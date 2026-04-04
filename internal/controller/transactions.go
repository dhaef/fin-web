package controller

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fin-web/internal/model"
)

var ExcludedIncomeCategories = []string{"34"}

var ExpenseCategoriesToExclude = []string{"34"}

func transactions(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path != "/" {
		return renderTemplate(w, Base[any]{}, "layout", []string{"not-found.html", "layout.html"})
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
		CategoriesToExclude: ExcludedIncomeCategories,
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
		CategoriesToExclude: ExpenseCategoriesToExclude,
		Type:                "expenses",
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
		Type:                "income",
		CategoriesToExclude: ExcludedIncomeCategories,
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
		Type:                "expenses",
		CategoriesToExclude: ExpenseCategoriesToExclude,
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
		Type:                "income",
		CategoriesToExclude: ExcludedIncomeCategories,
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

	cs, err := model.GetCategories(dbConn)
	if err != nil {
		fmt.Println("faile to get categories from DB: ", err.Error())
	}

	err = renderTemplate(w, Base[map[string]any]{
		Data: map[string]any{
			"transactions":           transactions,
			"startDate":              startDate,
			"endDate":                endDate,
			"orderBy":                orderBy,
			"orderDirection":         orderDirection,
			"categories":             cs,
			"selectedCategories":     selectedCatMap,
			"expensesCategoryCounts": expensesCategoryCounts,
			"incomeCategoryCounts":   incomeCategoryCounts,
			"expenseCountsByMonth":   expenseCountsByMonth,
			"incomeCountsByMonth":    incomeCountsByMonth,
			"eTotal":                 eTotal,
			"iTotal":                 iTotal,
			"total":                  iTotal - eTotal,
			"savedPercent":           math.Round(((iTotal - eTotal) / iTotal) * 100),
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

type UncategorizedTransactionsPage struct {
	Transactions []model.Transaction
}

func uncategorizedTransactions(w http.ResponseWriter, r *http.Request) error {
	emptyCustomCategory := true
	transactions, err := model.QueryTransactions(
		dbConn,
		model.QueryTransactionsFilters{
			EmptyCustomCategory: &emptyCustomCategory,
		},
	)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching transactions: " + err.Error(),
		}
	}

	err = renderTemplate(w, Base[UncategorizedTransactionsPage]{
		Data: UncategorizedTransactionsPage{
			Transactions: transactions,
		},
	}, "layout", []string{"transactions/uncategorized-transactions.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

type TransactionPage struct {
	Transaction model.Transaction
	Categories  []model.Category
	Success     bool
}

func transaction(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	transaction, err := model.GetTransaction(
		dbConn,
		id,
	)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching transaction: " + err.Error(),
		}
	}

	cs, err := model.GetCategories(dbConn)
	if err != nil {
		fmt.Println("faile to get categories from DB: ", err.Error())
	}

	responseCookie, err := r.Cookie("response")
	if err != nil && err != http.ErrNoCookie {
		fmt.Println("error getting cookie: " + err.Error())
	}

	success := responseCookie != nil && responseCookie.Value == "success"

	err = renderTemplate(w, Base[TransactionPage]{
		Data: TransactionPage{
			Transaction: transaction,
			Categories:  cs,
			Success:     success,
		},
	}, "layout", []string{"transactions/transaction.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

func updateTransaction(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	description := r.FormValue("description")
	category := r.FormValue("category")

	var categoryID *int
	if category != "" {
		c, err := strconv.Atoi(category)
		if err != nil {
			return APIError{
				Status:  http.StatusBadRequest,
				Message: "category must be an int",
			}
		}

		categoryID = &c
	}

	err := model.UpdateTransaction(dbConn, id, model.UpdateTransactionParams{
		Description: &description,
		CategoryID:  categoryID,
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error updating transaction: " + err.Error(),
		}
	}

	cookie := &http.Cookie{
		Name:     "response",
		Value:    "success",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(1 * time.Second),
	}

	http.SetCookie(w, cookie)

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	return nil
}

func deleteTransaction(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	err := model.DeleteTransaction(dbConn, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error deleting transaction: " + err.Error(),
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
