package controller

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"fin-web/internal/model"
	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCreateCategory(t *testing.T, db *sql.DB, label string, priority int, catType string) int {
	t.Helper()
	id, err := model.CreateCategory(db, label, priority, catType, false)
	require.NoError(t, err)
	return id
}

func seedTransaction(t *testing.T, db *sql.DB, id, name string, amount float64, date string, categoryID sql.NullInt32) {
	t.Helper()
	require.NoError(t, model.CreateTransaction(db, model.Transaction{
		ID:         id,
		Name:       name,
		Amount:     amount,
		Date:       date,
		Source:     "citi",
		Account:    "citi",
		CategoryID: categoryID,
	}))
}

func catID(id int) sql.NullInt32 {
	return sql.NullInt32{Valid: true, Int32: int32(id)}
}

func TestTransactionRenders(t *testing.T) {
	db := testutil.NewDB(t)
	seedTransaction(t, db, "tx-1", "WHOLE FOODS", 42.10, "2026-02-10", sql.NullInt32{})
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodGet, "/transactions/tx-1", nil)
	req.SetPathValue("id", "tx-1")
	rec := httptest.NewRecorder()
	require.NoError(t, c.transaction(rec, req))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "WHOLE FOODS")
}

func TestUpdateTransactionSuccess(t *testing.T) {
	db := testutil.NewDB(t)
	id := mustCreateCategory(t, db, "Groceries", 5, "fixed")
	seedTransaction(t, db, "tx-1", "WHOLE FOODS", 42.10, "2026-02-10", sql.NullInt32{})
	c := &Controller{db: db}

	req := newFormRequest("/transactions/tx-1", url.Values{
		"description":      {"weekly shop"},
		"category":         {strconv.Itoa(id)},
		"is_reimbursement": {"on"},
	})
	req.SetPathValue("id", "tx-1")
	rec := httptest.NewRecorder()
	require.NoError(t, c.updateTransaction(rec, req))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	// A short-lived success cookie is set on update.
	assert.Contains(t, rec.Result().Cookies()[0].Value, "success")

	got, err := model.GetTransaction(db, "tx-1")
	require.NoError(t, err)
	assert.Equal(t, "weekly shop", got.Description.String)
	assert.True(t, got.IsReimbursement)
	assert.Equal(t, int32(id), got.CategoryID.Int32)
}

func TestUpdateTransactionNonIntCategory(t *testing.T) {
	db := testutil.NewDB(t)
	seedTransaction(t, db, "tx-1", "WHOLE FOODS", 42.10, "2026-02-10", sql.NullInt32{})
	c := &Controller{db: db}

	req := newFormRequest("/transactions/tx-1", url.Values{"category": {"abc"}})
	req.SetPathValue("id", "tx-1")
	rec := httptest.NewRecorder()

	err := c.updateTransaction(rec, req)
	var apiErr APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
}

func TestDeleteTransaction(t *testing.T) {
	db := testutil.NewDB(t)
	seedTransaction(t, db, "tx-1", "WHOLE FOODS", 42.10, "2026-02-10", sql.NullInt32{})
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodPost, "/transactions/tx-1/delete", nil)
	req.SetPathValue("id", "tx-1")
	rec := httptest.NewRecorder()
	require.NoError(t, c.deleteTransaction(rec, req))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/", rec.Header().Get("Location"))

	txns, err := model.QueryTransactions(db, model.QueryTransactionsFilters{})
	require.NoError(t, err)
	assert.Empty(t, txns)
}

func TestUncategorizedTransactionsRenders(t *testing.T) {
	db := testutil.NewDB(t)
	categorized := mustCreateCategory(t, db, "Groceries", 5, "fixed")
	seedTransaction(t, db, "tx-1", "UNCATEGORIZED MERCHANT", 10, "2026-02-10", sql.NullInt32{})
	seedTransaction(t, db, "tx-2", "WHOLE FOODS", 20, "2026-02-11", catID(categorized))
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	require.NoError(t, c.uncategorizedTransactions(rec, httptest.NewRequest(http.MethodGet, "/transactions/uncategorized", nil)))

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	// Uncategorized rows (NULL category) must show; categorized ones must not.
	assert.Contains(t, body, "UNCATEGORIZED MERCHANT")
	assert.NotContains(t, body, "WHOLE FOODS")
}

func TestQueryTransactionsHidesIgnoredButKeepsUncategorized(t *testing.T) {
	db := testutil.NewDB(t)
	ignored, err := model.CreateCategory(db, "Hidden", 5, "fun", true)
	require.NoError(t, err)
	visible := mustCreateCategory(t, db, "Visible", 6, "fun")

	seedTransaction(t, db, "tx-ignored", "IGNORED", 10, "2026-02-10", catID(ignored))
	seedTransaction(t, db, "tx-visible", "VISIBLE", 20, "2026-02-11", catID(visible))
	seedTransaction(t, db, "tx-uncat", "UNCAT", 30, "2026-02-12", sql.NullInt32{})

	txns, err := model.QueryTransactions(db, model.QueryTransactionsFilters{})
	require.NoError(t, err)

	names := map[string]bool{}
	for _, tx := range txns {
		names[tx.Name] = true
	}
	assert.True(t, names["VISIBLE"], "non-ignored categorized transaction should show")
	assert.True(t, names["UNCAT"], "uncategorized transaction should show")
	assert.False(t, names["IGNORED"], "transaction in an is_ignored category should be hidden")
}

func TestTransactionsHomeRenders(t *testing.T) {
	db := testutil.NewDB(t)
	income := mustCreateCategory(t, db, "Salary", 1, "income")
	rent := mustCreateCategory(t, db, "Rent", 2, "fixed")
	// Income stored as a negative amount; expense as positive.
	seedTransaction(t, db, "tx-income", "PAYCHECK", -5000, "2026-02-01", catID(income))
	seedTransaction(t, db, "tx-rent", "RENT", 2000, "2026-02-05", catID(rent))
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodGet, "/?startDate=2026-02-01&endDate=2026-02-28", nil)
	rec := httptest.NewRecorder()
	require.NoError(t, c.transactions(rec, req))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "PAYCHECK")
}

func TestTransactionsHomeNonRootPathRendersNotFound(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}

	req := httptest.NewRequest(http.MethodGet, "/some/unknown/path", nil)
	rec := httptest.NewRecorder()
	require.NoError(t, c.transactions(rec, req))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
}

func TestGetNetCounts(t *testing.T) {
	expenses := []model.GroupByCounts{{Key: "02-2026", Value: 2000}}
	income := []model.GroupByCounts{
		{Key: "02-2026", Value: -5000},
		{Key: "03-2026", Value: -3000}, // no matching expense -> skipped
	}

	got := getNetCounts(expenses, income)
	require.Len(t, got, 1)
	assert.Equal(t, "02-2026", got[0].Key)
	// net = income(-5000) + expense(2000) = -3000; pct = (-3000 / -5000) * 100 = 60%.
	assert.InDelta(t, -3000, got[0].Net, 1e-9)
	assert.Equal(t, "60.00%", got[0].SavePercentage)
}

func TestGetNetCountsZeroIncomeNoDivideByZero(t *testing.T) {
	got := getNetCounts(
		[]model.GroupByCounts{{Key: "02-2026", Value: 100}},
		[]model.GroupByCounts{{Key: "02-2026", Value: 0}},
	)
	require.Len(t, got, 1)
	assert.Equal(t, "0.00%", got[0].SavePercentage)
}
