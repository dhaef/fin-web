package model

import (
	"database/sql"
	"strconv"
	"strings"
)

type Transaction struct {
	ID             string
	Account        string
	Amount         float64
	Description    sql.NullString
	Date           string
	Name           string
	CustomCategory sql.NullString
	Category       string
	Source         string
	CategoryID     sql.NullInt32
}

type QueryTransactionsFilters struct {
	OrderBy             string
	OrderDirection      string
	StartDate           string
	EndDate             string
	Categories          []string
	CategoriesToExclude []string
	Limit               int
	Type                string
	EmptyCustomCategory *bool
}

func buildWhere(queryStr string, args []any, filters QueryTransactionsFilters) (string, []any) {
	filterStrings := []string{}

	if filters.StartDate != "" {
		filterStrings = append(filterStrings, "date >= ?")
		args = append(args, filters.StartDate)
	}

	if filters.EndDate != "" {
		filterStrings = append(filterStrings, "date <= ?")
		args = append(args, filters.EndDate)
	}

	if len(filters.Categories) > 0 && filters.Categories[0] != "" {
		cStr := "("
		for idx, val := range filters.Categories {
			args = append(args, val)

			if idx == len(filters.Categories)-1 {
				cStr += " c.category = ?)"
			} else {
				cStr += " c.category = ? OR"
			}
		}
		filterStrings = append(filterStrings, cStr)
	}

	if filters.Type == "income" {
		filterStrings = append(filterStrings, "amount < 0")
	}

	if filters.Type == "expenses" {
		filterStrings = append(filterStrings, "amount >= 0")
	}

	if len(filters.CategoriesToExclude) > 0 && filters.CategoriesToExclude[0] != "" {
		for _, val := range filters.CategoriesToExclude {
			args = append(args, val)
			filterStrings = append(filterStrings, "c.category != ?")
		}
	}

	if filters.EmptyCustomCategory != nil {
		if !*filters.EmptyCustomCategory {
			filterStrings = append(filterStrings, "c.category IS NOT NULL")
		} else {
			filterStrings = append(filterStrings, "c.category IS NULL")
		}
	}

	if len(filterStrings) > 0 {
		queryStr += " WHERE"

		for idx, s := range filterStrings {
			if idx != 0 {
				queryStr += " AND " + s
			} else {
				queryStr += " " + s
			}
		}
	}

	return queryStr, args
}

func QueryTransactions(conn *sql.DB, filters QueryTransactionsFilters) ([]Transaction, error) {
	queryStr := "select t.id, name, amount, date, account, source, description, c.id, c.category as category from transactions as t left join categories as c on category_id = c.id"
	args := []any{}

	queryStr, args = buildWhere(queryStr, args, filters)

	if filters.OrderBy != "" {
		queryStr += " ORDER BY " + filters.OrderBy + " " + filters.OrderDirection
	}

	if filters.Limit > 0 {
		queryStr += " LIMIT " + strconv.Itoa(filters.Limit)
	}

	rows, err := conn.Query(
		queryStr,
		args...,
	)
	if err != nil {
		return []Transaction{}, err
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		transaction := Transaction{}
		if err := rows.Scan(
			&transaction.ID,
			&transaction.Name,
			&transaction.Amount,
			&transaction.Date,
			&transaction.Account,
			&transaction.Source,
			&transaction.Description,
			&transaction.CategoryID,
			&transaction.CustomCategory,
		); err != nil {
			return []Transaction{}, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func CategoryCounts(conn *sql.DB, filters QueryTransactionsFilters) ([]GroupByCounts, error) {
	queryStr := "SELECT c.id, c.category as category, SUM(t.amount) FROM transactions as t left join categories as c on t.category_id = c.id"
	args := []any{}

	queryStr, args = buildWhere(queryStr, args, filters)

	queryStr += " GROUP BY c.id"

	rows, err := conn.Query(
		queryStr,
		args...,
	)
	if err != nil {
		return []GroupByCounts{}, err
	}
	defer rows.Close()

	counts := []GroupByCounts{}
	for rows.Next() {
		count := GroupByCounts{}
		if err := rows.Scan(
			&count.ID,
			&count.Key,
			&count.Value,
		); err != nil {
			return []GroupByCounts{}, err
		}

		counts = append(counts, count)
	}

	return counts, nil
}

func CountsByDate(conn *sql.DB, filters QueryTransactionsFilters, dateStr string) ([]GroupByCounts, error) {
	queryStr := "SELECT strftime(\"" + dateStr + "\", date), SUM(amount) FROM transactions as t left join categories as c on t.category_id = c.id"
	args := []any{}

	queryStr, args = buildWhere(queryStr, args, filters)

	queryStr += " GROUP BY strftime(\"" + dateStr + "\", date)"

	rows, err := conn.Query(
		queryStr,
		args...,
	)
	if err != nil {
		return []GroupByCounts{}, err
	}
	defer rows.Close()

	counts := []GroupByCounts{}
	for rows.Next() {
		count := GroupByCounts{}
		if err := rows.Scan(
			&count.Key,
			&count.Value,
		); err != nil {
			return []GroupByCounts{}, err
		}

		counts = append(counts, count)
	}

	return counts, nil
}

func GetTransaction(conn *sql.DB, ID string) (Transaction, error) {
	queryStr := "select t.id, name, amount, date, account, source, description, c.id, c.category as category from transactions as t left join categories as c on category_id = c.id where t.id = ?"

	transaction := Transaction{}
	err := conn.QueryRow(
		queryStr,
		ID,
	).Scan(
		&transaction.ID,
		&transaction.Name,
		&transaction.Amount,
		&transaction.Date,
		&transaction.Account,
		&transaction.Source,
		&transaction.Description,
		&transaction.CategoryID,
		&transaction.CustomCategory,
	)
	if err != nil {
		return Transaction{}, err
	}

	return transaction, nil
}

type UpdateTransactionParams struct {
	CategoryID  *int
	Description *string
}

func UpdateTransaction(conn *sql.DB, ID string, params UpdateTransactionParams) error {
	queryStr := "UPDATE transactions SET"
	updates := []string{}
	args := []any{}

	if params.CategoryID != nil {
		updates = append(updates, " category_id = ?")
		args = append(args, *params.CategoryID)
	}

	if params.Description != nil {
		updates = append(updates, " description = ?")
		args = append(args, *params.Description)
	}

	if len(updates) == 0 {
		return nil
	}

	queryStr += strings.Join(updates, ",")
	queryStr += " WHERE id = ?"
	args = append(args, ID)

	_, err := conn.Exec(
		queryStr,
		args...,
	)
	if err != nil {
		return err
	}

	return nil
}

func CreateTransaction(conn *sql.DB, transaction Transaction) error {
	queryStr := "INSERT INTO transactions(id, name, amount, date, source, account, category, category_id, description) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)"
	args := []any{
		transaction.ID,
		transaction.Name,
		transaction.Amount,
		transaction.Date,
		transaction.Source,
		transaction.Account,
		transaction.Category,
	}

	if transaction.CategoryID.Valid {
		args = append(args, transaction.CategoryID.Int32)
	} else {
		args = append(args, nil)
	}

	args = append(args, transaction.Description)

	_, err := conn.Exec(
		queryStr,
		args...,
	)
	if err != nil {
		return err
	}

	return nil
}
