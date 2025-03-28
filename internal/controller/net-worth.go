package controller

import (
	"fin-web/internal/model"
	"fmt"
	"net/http"
)

func netWorth(w http.ResponseWriter, r *http.Request) error {
	netWorthItems, err := model.QueryNetWorthItems(netWorthDbConn, model.QueryNetWorthItemsFilters{
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
	}, "layout", []string{"net-worth/net-worth.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func netWorthItem(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	fmt.Println(id)

	netWorthItems, err := model.QueryNetWorthItems(netWorthDbConn, model.QueryNetWorthItemsFilters{
		Id: id,
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching net worth item: " + err.Error(),
		}
	}

	netWorthItem := model.NetWorthItem{}
	if len(netWorthItems) > 0 {
		netWorthItem = netWorthItems[0]
	}

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"netWorthItem": netWorthItem,
		},
	}, "layout", []string{"net-worth/net-worth-form.html", "net-worth/net-worth-item.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}
