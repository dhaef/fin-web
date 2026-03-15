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
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "failed to get stock shares: " + err.Error(),
		}
	}

	prices := []StockPrice{}
	priceMap := map[string]float64{}

	for _, s := range ss {
		if s.Shares == 0 {
			continue
		}

		if s.Ticker == "" {
			// get price from db??
			continue
		}

		price := StockPrice{
			Ticker: s.Ticker,
		}

		// request data from cache or get and set
		item, err := model.GetKVItem(dbConn, s.Ticker)
		if err != nil {
			if errors.Is(err, model.ErrKVItemNotFound) {
				stockInfoItems, err := tiingo.GetTickerInfo(tiingoToken, s.Ticker)
				if err != nil {
					return APIError{
						Status:  http.StatusInternalServerError,
						Message: "failed to get info from tiingo: " + err.Error(),
					}
				}

				if len(stockInfoItems) == 0 {
					return APIError{
						Status:  http.StatusInternalServerError,
						Message: "no stock info found for " + s.Ticker,
					}
				}

				tickerInfo := stockInfoItems[0]
				v := strconv.FormatFloat(float64(tickerInfo.Close), 'f', -1, 32)

				err = model.PutKVItem(dbConn, s.Ticker, v, time.Hour*24)
				if err != nil {
					return APIError{
						Status:  http.StatusInternalServerError,
						Message: "failed to cache item: " + err.Error(),
					}
				}

				price.Price = float64(tickerInfo.Close)
				value := float64(tickerInfo.Close) * s.Shares
				price.Value = value
				prices = append(prices, price)
				priceMap[s.Ticker] = float64(tickerInfo.Close)
				continue
			}

			return APIError{
				Status:  http.StatusBadRequest,
				Message: "failed to get item from cache: " + err.Error(),
			}
		}

		f, err := strconv.ParseFloat(item.Value, 64)
		if err != nil {
			return APIError{
				Status:  http.StatusBadRequest,
				Message: "failed to convert cache string to float: " + item.Value,
			}
		}
		price.Price = f
		priceMap[s.Ticker] = f

		value := f * s.Shares
		price.Value = value
		prices = append(prices, price)

	}

	trades, err := model.GetTrades(dbConn)
	if err != nil {
		return APIError{
			Status:  http.StatusBadRequest,
			Message: "failed to get trades: " + err.Error(),
		}
	}

	for idx, trade := range trades {
		currPrice, ok := priceMap[trade.Ticker]
		if ok {
			cv := currPrice * trade.Shares
			trade.CurrentValue = &cv
			gr := ((cv - trade.Total) / trade.Total) * 100
			grStr := fmt.Sprintf("%.2f", gr)
			trade.GrowthRate = &grStr

			if gr > 0 {
				trade.HasPositiveGrowth = true
			}

			trades[idx] = trade
		}
	}

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"prices": prices,
			"trades": trades,
		},
	}, "layout", []string{"trades/trades.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
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

	err = renderTemplate(w, Base{
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
	err := renderTemplate(w, Base{
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
		err := renderTemplate(w, Base{
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
		err := renderTemplate(w, Base{
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
