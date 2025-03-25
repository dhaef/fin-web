package controller

import (
	"fin-web/internal/model"
	"net/http"
)

func netWorth(w http.ResponseWriter, r *http.Request) error {
	netWorthItems, err := model.QueryNetWorthItems(netWorthDbConn, model.QueryTransactionsFilters{
		OrderBy:        "date",
		OrderDirection: "DESC",
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching net worth items: " + err.Error(),
		}
	}

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"netWorthItems": netWorthItems,
		},
	}, "layout", []string{"net-worth.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}
