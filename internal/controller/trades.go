package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"fin-web/internal/model"
	"fin-web/internal/tiingo"
)

type StockPrice struct {
	Price  float64
	Value  float64
	Ticker string
}

func trades(w http.ResponseWriter, r *http.Request) error {
	ss, err := model.GetStockShares(dbConn)
	if err != nil {
		return APIError{Status: http.StatusInternalServerError, Message: "failed to get stock shares: " + err.Error()}
	}

	prices, priceMap, err := processStockPrices(ss)
	if err != nil {
		return err // processStockPrices returns APIError
	}

	trades, err := model.GetTrades(dbConn)
	if err != nil {
		return APIError{Status: http.StatusBadRequest, Message: "failed to get trades: " + err.Error()}
	}

	for i := range trades {
		if price, ok := priceMap[trades[i].Ticker]; ok {
			enrichTradeData(&trades[i], price)
		}
	}

	return renderTemplate(w, Base[map[string]any]{
		Data: map[string]any{
			"prices": prices,
			"trades": trades,
		},
	}, "layout", []string{"trades/trades.html", "layout.html"})
}

func processStockPrices(ss []model.StockShare) ([]StockPrice, map[string]float64, error) {
	prices := []StockPrice{}
	priceMap := map[string]float64{}

	for _, s := range ss {
		if s.Shares <= 0 || s.Ticker == "" {
			continue
		}

		currentPrice, err := getOrFetchPrice(s.Ticker)
		if err != nil {
			return nil, nil, err
		}

		priceMap[s.Ticker] = currentPrice
		prices = append(prices, StockPrice{
			Ticker: s.Ticker,
			Price:  currentPrice,
			Value:  currentPrice * s.Shares,
		})
	}
	return prices, priceMap, nil
}

func getOrFetchPrice(ticker string) (float64, error) {
	// 1. Attempt to get from Cache (KV Store)
	item, err := model.GetKVItem(dbConn, ticker)
	if err == nil {
		// Cache Hit: Parse and return
		f, err := strconv.ParseFloat(item.Value, 64)
		if err != nil {
			return 0, APIError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("failed to convert cache string to float for %s: %s", ticker, item.Value),
			}
		}
		return f, nil
	}

	// 2. Handle Cache Miss vs. Actual DB Error
	if !errors.Is(err, model.ErrKVItemNotFound) {
		return 0, APIError{
			Status:  http.StatusInternalServerError,
			Message: "failed to get item from cache: " + err.Error(),
		}
	}

	// 3. Cache Miss: Fetch from Tiingo API
	stockInfoItems, err := tiingo.GetTickerInfo(tiingoToken, ticker)
	if err != nil {
		return 0, APIError{
			Status:  http.StatusInternalServerError,
			Message: "failed to get info from tiingo: " + err.Error(),
		}
	}

	if len(stockInfoItems) == 0 {
		return 0, APIError{
			Status:  http.StatusInternalServerError,
			Message: "no stock info found for " + ticker,
		}
	}

	// 4. Process API Result
	tickerInfo := stockInfoItems[0]
	price := float64(tickerInfo.Close)

	// Convert to string for KV storage (matching your original implementation)
	v := strconv.FormatFloat(price, 'f', -1, 32)

	// 5. Update Cache for 24 hours
	err = model.PutKVItem(dbConn, ticker, v, time.Hour*24)
	if err != nil {
		return 0, APIError{
			Status:  http.StatusInternalServerError,
			Message: "failed to cache item: " + err.Error(),
		}
	}

	return price, nil
}

func enrichTradeData(t *model.Trade, currentPrice float64) {
	cv := currentPrice * t.Shares
	t.CurrentValue = &cv

	// Prevent division by zero if Total is 0
	if t.Total == 0 {
		zero := "0.00"
		t.GrowthRate = &zero
		return
	}

	// Formula: ((Current - Total) / Total) * 100
	growth := ((cv - t.Total) / t.Total) * 100
	growthStr := fmt.Sprintf("%.2f", growth)

	t.GrowthRate = &growthStr
	t.HasPositiveGrowth = growth > 0
}

func trade(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	t, err := model.GetTrade(dbConn, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error getting trade: " + err.Error(),
		}
	}

	err = renderTemplate(w, Base[map[string]any]{
		Data: map[string]any{
			"trade": t,
		},
	}, "layout", []string{"trades/trade.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

func deleteTrade(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	err := model.DeleteTrade(dbConn, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error deleting trade: " + err.Error(),
		}
	}

	http.Redirect(w, r, "/trades", http.StatusSeeOther)
	return nil
}

func newTrade(w http.ResponseWriter, r *http.Request) error {
	err := renderTemplate(w, Base[map[string]any]{
		Data: map[string]any{
			"type": "create",
		},
	}, "layout", []string{"trades/trade.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func createTrade(w http.ResponseWriter, r *http.Request) error {
	errs := map[string]string{}
	var err error

	name := r.FormValue("name")
	if name == "" {
		errs["name"] = "name can't be empty"
	}

	ticker := r.FormValue("ticker")

	purchaseDate := r.FormValue("purchase_date")
	if purchaseDate == "" {
		errs["purchase_date"] = "purchase_date can't be empty"
	}

	sharesStr := r.FormValue("shares")
	var shares float64
	if sharesStr != "" {
		shares, err = strconv.ParseFloat(sharesStr, 64)
		if err != nil {
			errs["shares"] = "shares is not a valid float"
		}
	} else {
		errs["shares"] = "shares can't be empty"
	}

	priceStr := r.FormValue("price")
	var price float64
	if priceStr != "" {
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			errs["price"] = "price is not a valid float"
		}
	} else {
		errs["price"] = "price can't be empty"
	}

	tradeType := r.FormValue("type")
	if tradeType == "" {
		errs["type"] = "type can't be empty"
	}

	account := r.FormValue("account")
	if account == "" {
		errs["account"] = "account can't be empty"
	}

	if len(errs) != 0 {
		err := renderTemplate(w, Base[map[string]any]{
			Data: map[string]any{
				"errs": errs,
				"trade": map[string]any{
					"Name": sql.NullString{
						Valid:  true,
						String: name,
					},
					"Ticker":       ticker,
					"PurchaseDate": purchaseDate,
					"Shares":       sharesStr,
					"Price":        priceStr,
					"Type":         tradeType,
					"Account":      account,
				},
			},
		}, "layout", []string{"trades/trade.html", "layout.html"})
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		return nil
	}

	id, err := model.CreateTrade(dbConn, name, ticker, purchaseDate, shares, price, tradeType, account)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error creating trade: " + err.Error(),
		}
	}

	http.Redirect(w, r, "/trades/"+strconv.Itoa(id), http.StatusSeeOther)
	return nil
}

func updateTrade(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	errs := map[string]string{}
	var err error

	name := r.FormValue("name")
	if name == "" {
		errs["name"] = "name can't be empty"
	}

	ticker := r.FormValue("ticker")

	purchaseDate := r.FormValue("purchase_date")
	if purchaseDate == "" {
		errs["purchase_date"] = "purchase_date can't be empty"
	}

	sharesStr := r.FormValue("shares")
	var shares float64
	if sharesStr != "" {
		shares, err = strconv.ParseFloat(sharesStr, 64)
		if err != nil {
			errs["shares"] = "shares is not a valid float"
		}
	} else {
		errs["shares"] = "shares can't be empty"
	}

	priceStr := r.FormValue("price")
	var price float64
	if priceStr != "" {
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			errs["price"] = "price is not a valid float"
		}
	} else {
		errs["price"] = "price can't be empty"
	}

	tradeType := r.FormValue("type")
	if tradeType == "" {
		errs["type"] = "type can't be empty"
	}

	account := r.FormValue("account")
	if account == "" {
		errs["account"] = "account can't be empty"
	}

	if len(errs) != 0 {
		err := renderTemplate(w, Base[map[string]any]{
			Data: map[string]any{
				"errs": errs,
				"trade": map[string]any{
					"Name": sql.NullString{
						Valid:  true,
						String: name,
					},
					"Ticker":       ticker,
					"PurchaseDate": purchaseDate,
					"Shares":       sharesStr,
					"Price":        priceStr,
					"Type":         tradeType,
					"Account":      account,
				},
			},
		}, "layout", []string{"trades/trade.html", "layout.html"})
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		return nil
	}

	params := model.UpdateTradeParams{
		Name:         &name,
		Ticker:       &ticker,
		PurchaseDate: &purchaseDate,
		Shares:       &shares,
		Price:        &price,
		Type:         &tradeType,
		Account:      &account,
	}

	err = model.UpdateTrade(dbConn, id, params)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error updating trade: " + err.Error(),
		}
	}

	http.Redirect(w, r, "/trades/"+id, http.StatusSeeOther)
	return nil
}
