package model

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type QueryNetWorthItemsFilters struct {
	OrderBy        string
	OrderDirection string
	Limit          int
	ID             string
}

type NetWorthItem struct {
	ID            string
	Date          string
	Cash          float32
	Investment    float32
	Debit         float32
	Credit        float32
	Savings       float32
	Retirement    float32
	Loans         float32
	NetWorth      float32
	Change        float32
	ChangePercent string
}

func buildNetWorthItemWhere(queryStr string, args []any, filters QueryNetWorthItemsFilters) (string, []any) {
	filterStrings := []string{}

	if filters.ID != "" {
		filterStrings = append(filterStrings, "id = ?")
		args = append(args, filters.ID)
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

func QueryNetWorthItems(conn *sql.DB, filters QueryNetWorthItemsFilters) ([]NetWorthItem, error) {
	queryStr := "SELECT * FROM net_worth"
	args := []any{}

	queryStr, args = buildNetWorthItemWhere(queryStr, args, filters)

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
		return []NetWorthItem{}, err
	}
	defer rows.Close()

	netWorthItems := []NetWorthItem{}
	for rows.Next() {
		netWorthItem := NetWorthItem{}
		if err := rows.Scan(
			&netWorthItem.ID,
			&netWorthItem.Date,
			&netWorthItem.Cash,
			&netWorthItem.Investment,
			&netWorthItem.Debit,
			&netWorthItem.Credit,
			&netWorthItem.Savings,
			&netWorthItem.Retirement,
			&netWorthItem.Loans,
		); err != nil {
			return []NetWorthItem{}, err
		}

		netWorthItems = append(netWorthItems, netWorthItem)
	}

	return netWorthItems, nil
}

func GetNetWorthItem(conn *sql.DB, ID string) (NetWorthItem, error) {
	queryStr := "SELECT * FROM net_worth WHERE id = ?"

	netWorthItem := NetWorthItem{}
	err := conn.QueryRow(
		queryStr,
		ID,
	).Scan(
		&netWorthItem.ID,
		&netWorthItem.Date,
		&netWorthItem.Cash,
		&netWorthItem.Investment,
		&netWorthItem.Debit,
		&netWorthItem.Credit,
		&netWorthItem.Savings,
		&netWorthItem.Retirement,
		&netWorthItem.Loans,
	)
	if err != nil {
		return NetWorthItem{}, err
	}

	return netWorthItem, nil
}

type NetWorthItemParams struct {
	Date       *string
	Cash       *float32
	Investment *float32
	Debit      *float32
	Credit     *float32
	Savings    *float32
	Retirement *float32
	Loans      *float32
}

func UpdateNetWorthItem(conn *sql.DB, ID string, params NetWorthItemParams) error {
	queryStr := "UPDATE net_worth SET"
	updates := []string{}
	args := []any{}

	if params.Date != nil {
		updates = append(updates, " date = ?")
		args = append(args, *params.Date)
	}

	if params.Cash != nil {
		updates = append(updates, " cash = ?")
		args = append(args, *params.Cash)
	}

	if params.Investment != nil {
		updates = append(updates, " investment = ?")
		args = append(args, *params.Investment)
	}

	if params.Debit != nil {
		updates = append(updates, " debit = ?")
		args = append(args, *params.Debit)
	}

	if params.Credit != nil {
		updates = append(updates, " credit = ?")
		args = append(args, *params.Credit)
	}

	if params.Savings != nil {
		updates = append(updates, " savings = ?")
		args = append(args, *params.Savings)
	}

	if params.Retirement != nil {
		updates = append(updates, " retirement = ?")
		args = append(args, *params.Retirement)
	}

	if params.Loans != nil {
		updates = append(updates, " loans = ?")
		args = append(args, *params.Loans)
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

func CreateNetWorthItem(conn *sql.DB, params NetWorthItemParams) (string, error) {
	ID := uuid.NewString()
	queryStr := "INSERT INTO net_worth(id, date, cash, investment, debit, credit, savings, retirement, loans) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	args := []any{}

	args = append(args, ID)

	if params.Date != nil {
		args = append(args, *params.Date)
	}

	if params.Cash != nil {
		args = append(args, *params.Cash)
	}

	if params.Investment != nil {
		args = append(args, *params.Investment)
	}

	if params.Debit != nil {
		args = append(args, *params.Debit)
	}

	if params.Credit != nil {
		args = append(args, *params.Credit)
	}

	if params.Savings != nil {
		args = append(args, *params.Savings)
	}

	if params.Retirement != nil {
		args = append(args, *params.Retirement)
	}

	if params.Loans != nil {
		args = append(args, *params.Loans)
	}

	_, err := conn.Exec(
		queryStr,
		args...,
	)
	if err != nil {
		return "", err
	}

	return ID, nil
}
