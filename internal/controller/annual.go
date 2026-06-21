package controller

import (
	"fmt"
	"net/http"

	"fin-web/internal/model"
)

type AnnualPage struct {
	IncomeCountsByYear  []model.GroupByCounts
	ExpenseCountsByYear []model.GroupByCounts
	NetCounts           []NetCounts
}

func (c *Controller) annual(w http.ResponseWriter, r *http.Request) error {
	incomeCountsByYear, err := model.CountsByDate(c.db, model.QueryTransactionsFilters{
		Type: "income",
	}, "%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching income counts by year: " + err.Error(),
		}
	}

	expenseCountsByYear, err := model.CountsByDate(c.db, model.QueryTransactionsFilters{
		Type: "expenses",
	}, "%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching expense counts by year: " + err.Error(),
		}
	}

	netCounts := getNetCounts(expenseCountsByYear, incomeCountsByYear)

	err = renderTemplate(w, Base[AnnualPage]{
		Data: AnnualPage{
			IncomeCountsByYear:  incomeCountsByYear,
			ExpenseCountsByYear: expenseCountsByYear,
			NetCounts:           netCounts,
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
			var roundedString string
			if item.Value != 0 {
				roundedString = fmt.Sprintf("%.2f%%", (net/item.Value)*100)
			} else {
				roundedString = "0.00%"
			}

			amountAndPercents = append(amountAndPercents, NetCounts{
				Net:            net,
				Key:            item.Key,
				SavePercentage: roundedString,
			})
		}
	}

	return amountAndPercents
}
