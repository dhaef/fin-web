package controller

import (
	"fin-web/internal/model"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

func favicon(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func home(w http.ResponseWriter, r *http.Request) error {
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
		transactions, err := model.QueryTransactions(dbConn, model.QueryTransactionsFilters{
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

	transactions, err := model.QueryTransactions(dbConn, model.QueryTransactionsFilters{
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

	expensesCategoryCounts, err := model.CategoryCounts(dbConn, model.QueryTransactionsFilters{
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

	incomeCategoryCounts, err := model.CategoryCounts(dbConn, model.QueryTransactionsFilters{
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

	expenseCountsByMonth, err := model.CountsByDate(dbConn, model.QueryTransactionsFilters{
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

	incomeCountsByMonth, err := model.CountsByDate(dbConn, model.QueryTransactionsFilters{
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

func annual(w http.ResponseWriter, r *http.Request) error {
	incomeCountsByYear, err := model.CountsByDate(dbConn, model.QueryTransactionsFilters{
		Categories:          []string{"work", "interest", "venmo", "miscellaneousIncome"},
		CategoriesToExclude: []string{"debit"},
	}, "%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching income counts by year: " + err.Error(),
		}
	}

	expenseCountsByYear, err := model.CountsByDate(dbConn, model.QueryTransactionsFilters{
		CategoriesToExclude: []string{"debit", "work", "interest", "venmo", "miscellaneousIncome"},
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
