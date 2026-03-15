package model

import (
	"database/sql"
)

type StockShare struct {
	Ticker string
	Shares float64
	Name   string
}

func GetStockShares(conn *sql.DB) ([]StockShare, error) {
	rows, err := conn.Query(
		"SELECT ticker, SUM(CASE WHEN type = 'sell' THEN -shares ELSE shares END) as shares, name FROM trades GROUP BY ticker",
	)
	if err != nil {
		return []StockShare{}, err
	}
	defer rows.Close()

	shares := []StockShare{}

	for rows.Next() {
		share := StockShare{}
		if err := rows.Scan(
			&share.Ticker,
			&share.Shares,
			&share.Name,
		); err != nil {
			return []StockShare{}, err
		}

		shares = append(shares, share)

	}

	return shares, nil
}

type Trade struct {
	ID           int
	Ticker       string
	PurchaseDate string
	Shares       float64
	Price        float64
	Type         string
	Account      string
	Name         sql.NullString
	Total        float64

	CurrentValue      *float64
	GrowthRate        *string
	HasPositiveGrowth bool
}

func GetTrades(conn *sql.DB) ([]Trade, error) {
	rows, err := conn.Query(
		"SELECT id, ticker, purchase_date, shares, price, type, account, name, price * shares as total  FROM trades ORDER BY purchase_date DESC",
	)
	if err != nil {
		return []Trade{}, err
	}
	defer rows.Close()

	trades := []Trade{}

	for rows.Next() {
		trade := Trade{}
		// var purchaseDateStr string
		if err := rows.Scan(
			&trade.ID,
			&trade.Ticker,
			&trade.PurchaseDate,
			&trade.Shares,
			&trade.Price,
			&trade.Type,
			&trade.Account,
			&trade.Name,
			&trade.Total,
		); err != nil {
			return []Trade{}, err
		}

		trades = append(trades, trade)

	}

	return trades, nil
}
