package model

import (
	"database/sql"
	"strconv"
)

type NetWorthItem struct {
	Id            string
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
	ChangePercent float32
}

func QueryNetWorthItems(conn *sql.DB, filters QueryTransactionsFilters) ([]NetWorthItem, error) {
	queryStr := "SELECT * FROM net_worth"
	args := []any{}

	// queryStr, args = buildWhere(queryStr, args, filters)

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
			&netWorthItem.Id,
			&netWorthItem.Date,
			&netWorthItem.Cash,
			&netWorthItem.Investment,
			&netWorthItem.Debit,
			&netWorthItem.Credit,
			&netWorthItem.Savings,
			&netWorthItem.Retirement,
			&netWorthItem.Loans,
			&netWorthItem.NetWorth,
			&netWorthItem.Change,
			&netWorthItem.ChangePercent,
		); err != nil {
			return []NetWorthItem{}, err
		}

		netWorthItems = append(netWorthItems, netWorthItem)
	}

	return netWorthItems, nil
}
