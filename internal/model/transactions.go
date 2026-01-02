package model

import (
	"database/sql"
	"strconv"
)

type Transactions struct {
	ID             string
	Account        string
	Amount         float64
	Description    sql.NullString
	Date           string
	Name           string
	CustomCategory sql.NullString
	Category       string
	Source         string
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

func Categories() []string {
	return []string{
		"work",
		"grocery",
		"foodOut",
		"flights",
		"utilities",
		"rent",
		"debit",
		"venmo",
		"gas",
		"car",
		"rentals",
		"transportation",
		"healthCare",
		"tech",
		"entertainment",
		"interest",
		"hotels",
		"gym",
		"insurance",
		"taxes",
		"government",
		"wedding",
		"mexico",
		"merchandise",
		"miscellaneousIncome",
	}
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
				cStr += " customCategory = ?)"
			} else {
				cStr += " customCategory = ? OR"
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
			filterStrings = append(filterStrings, "customCategory != ?")
		}
	}

	if filters.EmptyCustomCategory != nil {
		if !*filters.EmptyCustomCategory {
			filterStrings = append(filterStrings, "customCategory IS NOT NULL")
		} else {
			filterStrings = append(filterStrings, "customCategory IS NULL")
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

func QueryTransactions(conn *sql.DB, filters QueryTransactionsFilters) ([]Transactions, error) {
	queryStr := "SELECT * FROM transactions"
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
		return []Transactions{}, err
	}
	defer rows.Close()

	transactions := []Transactions{}
	for rows.Next() {
		transaction := Transactions{}
		if err := rows.Scan(
			&transaction.Name,
			&transaction.Amount,
			&transaction.Date,
			&transaction.Source,
			&transaction.Account,
			&transaction.Category,
			&transaction.ID,
			&transaction.CustomCategory,
			&transaction.Description,
		); err != nil {
			return []Transactions{}, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func CategoryCounts(conn *sql.DB, filters QueryTransactionsFilters) ([]GroupByCounts, error) {
	queryStr := "SELECT customCategory, SUM(amount) FROM transactions"
	args := []any{}

	queryStr, args = buildWhere(queryStr, args, filters)

	queryStr += " GROUP BY customCategory"

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

func CountsByDate(conn *sql.DB, filters QueryTransactionsFilters, dateStr string) ([]GroupByCounts, error) {
	queryStr := "SELECT strftime(\"" + dateStr + "\", date), SUM(amount) FROM transactions"
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
