package model

import (
	"database/sql"
	"strings"
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

func GetTrade(conn *sql.DB, ID string) (Trade, error) {
	queryStr := "SELECT id, ticker, purchase_date, shares, price, type, account, name FROM trades where id = ?"

	trade := Trade{}
	err := conn.QueryRow(
		queryStr,
		ID,
	).Scan(
		&trade.ID,
		&trade.Ticker,
		&trade.PurchaseDate,
		&trade.Shares,
		&trade.Price,
		&trade.Type,
		&trade.Account,
		&trade.Name,
	)
	if err != nil {
		return Trade{}, err
	}

	return trade, nil
}

func CreateTrade(conn *sql.DB, name string, ticker string, purchaseDate string, shares float64, price float64, tradeType string, account string) (int, error) {
	queryStr := "INSERT INTO trades (name, ticker, purchase_date, shares, price, type, account) VALUES(?, ?, ?, ?, ?, ?, ?) RETURNING id"
	args := []any{
		name,
		ticker,
		purchaseDate,
		shares,
		price,
		tradeType,
		account,
	}

	var lastInsertID int
	err := conn.QueryRow(
		queryStr,
		args...,
	).Scan(&lastInsertID)
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

type UpdateTradeParams struct {
	Ticker       *string
	PurchaseDate *string
	Shares       *float64
	Price        *float64
	Type         *string
	Account      *string
	Name         *string
}

func UpdateTrade(conn *sql.DB, ID string, params UpdateTradeParams) error {
	queryStr := "UPDATE trades SET"
	updates := []string{}
	args := []any{}

	if params.Ticker != nil {
		updates = append(updates, " ticker = ?")
		args = append(args, *params.Ticker)
	}

	if params.PurchaseDate != nil {
		updates = append(updates, " purchase_date = ?")
		args = append(args, *params.PurchaseDate)
	}

	if params.Shares != nil {
		updates = append(updates, " shares = ?")
		args = append(args, *params.Shares)
	}

	if params.Price != nil {
		updates = append(updates, " price = ?")
		args = append(args, *params.Price)
	}

	if params.Type != nil {
		updates = append(updates, " type = ?")
		args = append(args, *params.Type)
	}

	if params.Account != nil {
		updates = append(updates, " account = ?")
		args = append(args, *params.Account)
	}

	if params.Name != nil {
		updates = append(updates, " name = ?")
		args = append(args, *params.Name)
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

func DeleteTrade(conn *sql.DB, ID string) error {
	queryStr := "DELETE FROM trades WHERE id = ?"

	_, err := conn.Exec(
		queryStr,
		ID,
	)
	if err != nil {
		return err
	}

	return nil
}
