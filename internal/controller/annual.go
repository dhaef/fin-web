package controller

import (
	"fin-web/internal/model"
	"net/http"
)

func annual(w http.ResponseWriter, r *http.Request) error {
	incomeCountsByYear, err := model.CountsByDate(transactionsDbConn, model.QueryTransactionsFilters{
		Categories:          IncomeCategories,
		CategoriesToExclude: ExcludedIncomeCategories,
	}, "%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching income counts by year: " + err.Error(),
		}
	}

	expenseCountsByYear, err := model.CountsByDate(transactionsDbConn, model.QueryTransactionsFilters{
		CategoriesToExclude: ExpenseCategoriesToExclude,
	}, "%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching expense counts by year: " + err.Error(),
		}
	}

	netCounts := getNetCounts(expenseCountsByYear, incomeCountsByYear)

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"incomeCountsByYear":  incomeCountsByYear,
			"expenseCountsByYear": expenseCountsByYear,
			"netCounts":           netCounts,
		},
	}, "layout", []string{"annual.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

type NetCounts struct {
	Net float64
	Key string
}

func getNetCounts(expenses []model.GroupByCounts, income []model.GroupByCounts) []NetCounts {
	expenseMap := map[string]float64{}
	for _, item := range expenses {
		expenseMap[item.Key] = item.Value
	}

	amountAndPercents := []NetCounts{}
	for _, item := range income {
		net := item.Value + expenseMap[item.Key]
		amountAndPercents = append(amountAndPercents, NetCounts{
			Net: net,
			Key: item.Key,
		})
	}

	return amountAndPercents
}
