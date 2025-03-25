package controller

import (
	"fin-web/internal/model"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

func transactions(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path != "/" {
		return renderTemplate(w, "", "layout", []string{"not-found.html", "layout.html"})
	}

	q := r.URL.Query()
	startDate := q.Get("startDate")
	endDate := q.Get("endDate")
	orderBy := q.Get("sortBy")
	orderDirection := q.Get("sortDirection")
	categories := strings.Split(q.Get("categories"), ",")

	if orderBy == "" {
		orderBy = "amount"
	}

	if orderDirection == "" {
		orderDirection = "DESC"
	}

	if endDate == "" {
		transactions, err := model.QueryTransactions(transactionsDbConn, model.QueryTransactionsFilters{
			OrderBy:        "date",
			OrderDirection: "DESC",
			Limit:          1,
		})
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: "error fetching default dates: " + err.Error(),
			}
		}

		date, err := time.Parse("2006-01-02", transactions[0].Date)
		if err != nil {
			fmt.Println(err.Error())
		}

		firstDayOfThisMonth, endOfThisMonth := getStartAndEndOfMonth(date)

		startDate = firstDayOfThisMonth.Format("2006-01-02")
		endDate = endOfThisMonth.Format("2006-01-02")
	}

	transactions, err := model.QueryTransactions(transactionsDbConn, model.QueryTransactionsFilters{
		OrderBy:             orderBy,
		OrderDirection:      orderDirection,
		StartDate:           startDate,
		EndDate:             endDate,
		Categories:          categories,
		CategoriesToExclude: []string{"debit"},
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching transactions: " + err.Error(),
		}
	}

	var eTotal float64
	var iTotal float64
	for _, val := range transactions {
		if val.Amount >= 0 {
			eTotal += val.Amount
		} else {
			iTotal += math.Abs(val.Amount)
		}
	}

	expensesCategoryCounts, err := model.CategoryCounts(transactionsDbConn, model.QueryTransactionsFilters{
		OrderBy:             orderBy,
		OrderDirection:      orderDirection,
		StartDate:           startDate,
		EndDate:             endDate,
		Categories:          categories,
		CategoriesToExclude: []string{"debit", "work", "interest", "venmo", "miscellaneousIncome"},
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching expense category counts: " + err.Error(),
		}
	}

	incomeCategoryCounts, err := model.CategoryCounts(transactionsDbConn, model.QueryTransactionsFilters{
		OrderBy:             orderBy,
		OrderDirection:      orderDirection,
		StartDate:           startDate,
		EndDate:             endDate,
		Categories:          []string{"work", "interest", "venmo", "miscellaneousIncome"},
		CategoriesToExclude: []string{"debit"},
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching income category counts: " + err.Error(),
		}
	}

	date, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		fmt.Println(err.Error())
	}

	startOfMonthOneYearAgo, _ := getStartAndEndOfMonth(date.AddDate(0, -11, 0))

	expenseCountsByMonth, err := model.CountsByDate(transactionsDbConn, model.QueryTransactionsFilters{
		StartDate:           startOfMonthOneYearAgo.Format("2006-01-02"),
		EndDate:             endDate,
		Categories:          categories,
		CategoriesToExclude: []string{"debit", "work", "interest", "venmo", "miscellaneousIncome"},
	}, "%m-%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching expense counts by month: " + err.Error(),
		}
	}

	incomeCountsByMonth, err := model.CountsByDate(transactionsDbConn, model.QueryTransactionsFilters{
		StartDate:           startOfMonthOneYearAgo.Format("2006-01-02"),
		EndDate:             endDate,
		Categories:          []string{"work", "interest", "venmo", "miscellaneousIncome"},
		CategoriesToExclude: []string{"debit"},
	}, "%m-%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching income counts by month: " + err.Error(),
		}
	}

	netCounts := getNetCounts(expenseCountsByMonth, incomeCountsByMonth)

	selectedCatMap := map[string]bool{}

	for _, val := range categories {
		selectedCatMap[val] = true
	}

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"transactions":           transactions,
			"startDate":              startDate,
			"endDate":                endDate,
			"orderBy":                orderBy,
			"orderDirection":         orderDirection,
			"categories":             model.Categories(),
			"selectedCategories":     selectedCatMap,
			"expensesCategoryCounts": expensesCategoryCounts,
			"incomeCategoryCounts":   incomeCategoryCounts,
			"expenseCountsByMonth":   expenseCountsByMonth,
			"incomeCountsByMonth":    incomeCountsByMonth,
			"eTotal":                 eTotal,
			"iTotal":                 iTotal,
			"total":                  iTotal - eTotal,
			"netCounts":              netCounts,
		},
	}, "layout", []string{"transactions/transactions.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}
