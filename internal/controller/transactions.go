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

type TransactionsPage struct {
	Transactions           []model.Transaction
	StartDate              string
	EndDate                string
	OrderBy                string
	OrderDirection         string
	Categories             []model.Category
	SelectedCategories     map[string]bool
	ExpensesCategoryCounts []model.GroupByCounts
	IncomeCategoryCounts   []model.GroupByCounts
	ExpenseCountsByMonth   []model.GroupByCounts
	IncomeCountsByMonth    []model.GroupByCounts
	ETotal                 float64
	ITotal                 float64
	Total                  float64
	SavedPercent           int
	NetCounts              []NetCounts
	FixedCosts             float64
	FixedCostsPercent      int
	GuiltFree              float64
	GuiltFreePercent       int
}

func (c *Controller) transactions(w http.ResponseWriter, r *http.Request) error {
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
		transactions, err := model.QueryTransactions(c.db, model.QueryTransactionsFilters{
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

	transactions, err := model.QueryTransactions(c.db, model.QueryTransactionsFilters{
		OrderBy:        orderBy,
		OrderDirection: orderDirection,
		StartDate:      startDate,
		EndDate:        endDate,
		Categories:     categories,
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching transactions: " + err.Error(),
		}
	}

	cs, err := model.GetCategories(c.db)
	if err != nil {
		fmt.Println("faile to get categories from DB: ", err.Error())
	}

	eTotal, err := model.SumTransactions(c.db, model.QueryTransactionsFilters{
		StartDate:  startDate,
		EndDate:    endDate,
		Categories: categories,
		Type:       "expenses",
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error suming transactions: " + err.Error(),
		}
	}

	iTotal, err := model.SumTransactions(c.db, model.QueryTransactionsFilters{
		StartDate:  startDate,
		EndDate:    endDate,
		Categories: categories,
		Type:       "income",
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error suming transactions: " + err.Error(),
		}
	}
	iTotal = math.Abs(iTotal)

	fixedCosts, err := model.SumTransactions(c.db, model.QueryTransactionsFilters{
		StartDate:  startDate,
		EndDate:    endDate,
		Categories: categories,
		Type:       "fixed",
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error suming transactions: " + err.Error(),
		}
	}

	guiltFree, err := model.SumTransactions(c.db, model.QueryTransactionsFilters{
		StartDate:  startDate,
		EndDate:    endDate,
		Categories: categories,
		Type:       "fun",
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error suming transactions: " + err.Error(),
		}
	}

	expensesCategoryCounts, err := model.CategoryCounts(c.db, model.QueryTransactionsFilters{
		OrderBy:        orderBy,
		OrderDirection: orderDirection,
		StartDate:      startDate,
		EndDate:        endDate,
		Categories:     categories,
		Type:           "expenses",
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching expense category counts: " + err.Error(),
		}
	}

	incomeCategoryCounts, err := model.CategoryCounts(c.db, model.QueryTransactionsFilters{
		OrderBy:        orderBy,
		OrderDirection: orderDirection,
		StartDate:      startDate,
		EndDate:        endDate,
		Type:           "income",
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

	expenseCountsByMonth, err := model.CountsByDate(c.db, model.QueryTransactionsFilters{
		StartDate:  startOfMonthOneYearAgo.Format("2006-01-02"),
		EndDate:    endDate,
		Categories: categories,
		Type:       "expenses",
	}, "%m-%Y")
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching expense counts by month: " + err.Error(),
		}
	}

	incomeCountsByMonth, err := model.CountsByDate(c.db, model.QueryTransactionsFilters{
		StartDate: startOfMonthOneYearAgo.Format("2006-01-02"),
		EndDate:   endDate,
		Type:      "income",
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

	err = renderTemplate(w, Base[TransactionsPage]{
		Data: TransactionsPage{
			Transactions:           transactions,
			StartDate:              startDate,
			EndDate:                endDate,
			OrderBy:                orderBy,
			OrderDirection:         orderDirection,
			Categories:             cs,
			SelectedCategories:     selectedCatMap,
			ExpensesCategoryCounts: expensesCategoryCounts,
			IncomeCategoryCounts:   incomeCategoryCounts,
			ExpenseCountsByMonth:   expenseCountsByMonth,
			IncomeCountsByMonth:    incomeCountsByMonth,
			ETotal:                 eTotal,
			ITotal:                 iTotal,
			Total:                  iTotal - eTotal,
			SavedPercent:           int(math.Round(((iTotal - eTotal) / iTotal) * 100)),
			NetCounts:              netCounts,
			FixedCosts:             fixedCosts,
			FixedCostsPercent:      int(math.Round((fixedCosts / iTotal) * 100)),
			GuiltFree:              guiltFree,
			GuiltFreePercent:       int(math.Round((guiltFree / iTotal) * 100)),
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

func (c *Controller) uncategorizedTransactions(w http.ResponseWriter, r *http.Request) error {
	emptyCustomCategory := true
	transactions, err := model.QueryTransactions(
		c.db,
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

func (c *Controller) transaction(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	transaction, err := model.GetTransaction(
		c.db,
		id,
	)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching transaction: " + err.Error(),
		}
	}

	cs, err := model.GetCategories(c.db)
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

func (c *Controller) updateTransaction(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	description := r.FormValue("description")
	category := r.FormValue("category")
	isReimbursementStr := r.FormValue("is_reimbursement")

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

	var isReimbursement bool
	if isReimbursementStr == "on" {
		isReimbursement = true
	}

	err := model.UpdateTransaction(c.db, id, model.UpdateTransactionParams{
		Description:     &description,
		CategoryID:      categoryID,
		IsReimbursement: &isReimbursement,
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

func (c *Controller) deleteTransaction(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	err := model.DeleteTransaction(c.db, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error deleting transaction: " + err.Error(),
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
