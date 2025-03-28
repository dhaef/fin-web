package model

import (
	"database/sql"
	"strconv"
)

type QueryNetWorthItemsFilters struct {
	OrderBy        string
	OrderDirection string
	Limit          int
	Id             string
}

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

func buildNetWorthItemWhere(queryStr string, args []any, filters QueryNetWorthItemsFilters) (string, []any) {
	filterStrings := []string{}

	if filters.Id != "" {
		filterStrings = append(filterStrings, "id = ?")
		args = append(args, filters.Id)
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
