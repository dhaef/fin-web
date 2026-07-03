package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fin-web/internal/model"
	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustCreateCategory and seedTransaction are shared helpers defined in
// transactions_handler_test.go.

func TestComputeSavingsRowsRolling(t *testing.T) {
	flows := []model.MonthlyFlow{
		{Month: "2026-01", Income: 1000, Expense: 500},
		{Month: "2026-02", Income: 1000, Expense: 700},
	}

	rows := computeSavingsRows(flows)
	require.Len(t, rows, 2)

	// First month: only itself in every window.
	assert.Equal(t, "50.0%", rows[0].Rate)
	assert.Equal(t, "50.0%", rows[0].Rolling3)
	assert.Equal(t, "50.0%", rows[0].Rolling12)
	assert.InDelta(t, 500, rows[0].Savings, 1e-6)

	// Second month: single-month 30%, but rolling aggregates both months:
	// (2000-1200)/2000 = 40%.
	assert.Equal(t, "30.0%", rows[1].Rate)
	assert.Equal(t, "40.0%", rows[1].Rolling3)
	assert.Equal(t, "40.0%", rows[1].Rolling12)
}

func TestComputeSavingsRowsWindowBoundary(t *testing.T) {
	// Four months; the 4th has all the spending. Rolling-3 excludes month 1,
	// rolling-12 includes it, so the two rates must differ.
	flows := []model.MonthlyFlow{
		{Month: "2026-01", Income: 100, Expense: 0},
		{Month: "2026-02", Income: 100, Expense: 0},
		{Month: "2026-03", Income: 100, Expense: 0},
		{Month: "2026-04", Income: 100, Expense: 90},
	}

	rows := computeSavingsRows(flows)

	// Rolling-3 over months 2..4: (300-90)/300 = 70%.
	assert.Equal(t, "70.0%", rows[3].Rolling3)
	// Rolling-12 over months 1..4: (400-90)/400 = 77.5%.
	assert.Equal(t, "77.5%", rows[3].Rolling12)
}

func TestSavingsRateNoIncome(t *testing.T) {
	assert.Equal(t, "—", formatSavingsRate(0, 100))
	assert.Equal(t, "—", formatSavingsRate(-50, 100))
	assert.Equal(t, "25.0%", formatSavingsRate(1000, 750))
}

func TestFormatPct(t *testing.T) {
	assert.Equal(t, "—", formatPct(400, 0))
	assert.Equal(t, "40.0%", formatPct(400, 1000))
	// Overspending shows as a negative rate rather than being hidden.
	assert.Equal(t, "-20.0%", formatPct(-200, 1000))
}

func TestTrailingWindowStart(t *testing.T) {
	assert.Equal(t, "", trailingWindowStart(nil, 12))

	few := []model.MonthlyFlow{{Month: "2026-01"}, {Month: "2026-02"}}
	assert.Equal(t, "2026-01-01", trailingWindowStart(few, 12))

	many := make([]model.MonthlyFlow, 15)
	for i := range many {
		many[i] = model.MonthlyFlow{Month: "m"}
	}
	many[3] = model.MonthlyFlow{Month: "2025-04"}
	// 15 months, window 12 -> start at index 3.
	assert.Equal(t, "2025-04-01", trailingWindowStart(many, 12))
}

func TestHealthHandlerRenders(t *testing.T) {
	db := testutil.NewDB(t)
	income := mustCreateCategory(t, db, "salary", 1, "income")
	fixed := mustCreateCategory(t, db, "rent", 2, "fixed")
	fun := mustCreateCategory(t, db, "dining", 3, "fun")

	seedTransaction(t, db, "jan-income", "salary", -1000, "2026-01-15", catID(income))
	seedTransaction(t, db, "jan-fixed", "rent", 400, "2026-01-16", catID(fixed))
	seedTransaction(t, db, "jan-fun", "dining", 100, "2026-01-17", catID(fun))

	c := &Controller{db: db}
	rec := httptest.NewRecorder()
	require.NoError(t, c.health(rec, httptest.NewRequest(http.MethodGet, "/health", nil)))

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	// 50/30/20 for the single month: needs 40%, wants 10%, savings 50%.
	assert.Contains(t, body, "40.0%")
	assert.Contains(t, body, "50.0%")
	assert.Contains(t, body, "2026-01")
}

func TestHealthHandlerEmpty(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}
	rec := httptest.NewRecorder()
	require.NoError(t, c.health(rec, httptest.NewRequest(http.MethodGet, "/health", nil)))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Not enough income")
}
