package controller

import (
	"errors"
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

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"prices": prices,
			"trades": trades,
		},
	}, "layout", []string{"trades.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}
