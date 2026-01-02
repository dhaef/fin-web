package controller

import (
	"fmt"
	"net/http"

	"fin-web/internal/model"
)

func annual(w http.ResponseWriter, r *http.Request) error {
	incomeCountsByYear, err := model.CountsByDate(transactionsDBConn, model.QueryTransactionsFilters{
		Type:                "income",
		CategoriesToExclude: ExcludedIncomeCategories,
	}, "%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching income counts by year: " + err.Error(),
		}
	}

	expenseCountsByYear, err := model.CountsByDate(transactionsDBConn, model.QueryTransactionsFilters{
		CategoriesToExclude: ExpenseCategoriesToExclude,
		Type:                "expenses",
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
	Net            float64
	Key            string
	SavePercentage string
}

func getNetCounts(expenses []model.GroupByCounts, income []model.GroupByCounts) []NetCounts {
	expenseMap := map[string]float64{}
	for _, item := range expenses {
		expenseMap[item.Key] = item.Value
	}

	amountAndPercents := []NetCounts{}
	for _, item := range income {
		expenseAmount, ok := expenseMap[item.Key]
		if ok {
			net := item.Value + expenseAmount
			roundedString := fmt.Sprintf("%.2f%%", (net/item.Value)*100)

			amountAndPercents = append(amountAndPercents, NetCounts{
				Net:            net,
				Key:            item.Key,
				SavePercentage: roundedString,
			})
		}
	}

	return amountAndPercents
}
