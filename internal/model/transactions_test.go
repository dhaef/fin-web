package model

import (
	"database/sql"
	"testing"

	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedTypedCategory inserts a category with an explicit type and returns its id.
func seedTypedCategory(t *testing.T, db *sql.DB, label string, priority int, catType string) int {
	t.Helper()
	res, err := db.Exec(
		"INSERT INTO categories(label, priority, type) VALUES(?, ?, ?)",
		label, priority, catType,
	)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

// seedTransaction inserts a single transaction. Amounts follow the app
// convention: income is stored negative, expenses positive.
func seedTransaction(t *testing.T, db *sql.DB, id string, amount float64, date string, categoryID int) {
	t.Helper()
	_, err := db.Exec(
		"INSERT INTO transactions(id, name, amount, date, account, source, category_id) VALUES(?, ?, ?, ?, ?, ?, ?)",
		id, "txn "+id, amount, date, "test", "test", categoryID,
	)
	require.NoError(t, err)
}

// seedFlows seeds two months: Jan (income 1000, fixed 400, fun 100) and Feb
// (income 1000, fixed 400, fun 300).
func seedFlows(t *testing.T, db *sql.DB) {
	t.Helper()
	income := seedTypedCategory(t, db, "salary", 1, "income")
	fixed := seedTypedCategory(t, db, "rent", 2, "fixed")
	fun := seedTypedCategory(t, db, "dining", 3, "fun")

	seedTransaction(t, db, "jan-income", -1000, "2026-01-15", income)
	seedTransaction(t, db, "jan-fixed", 400, "2026-01-16", fixed)
	seedTransaction(t, db, "jan-fun", 100, "2026-01-17", fun)

	seedTransaction(t, db, "feb-income", -1000, "2026-02-15", income)
	seedTransaction(t, db, "feb-fixed", 400, "2026-02-16", fixed)
	seedTransaction(t, db, "feb-fun", 300, "2026-02-17", fun)
}

func TestMonthlyFlows(t *testing.T) {
	db := testutil.NewDB(t)
	seedFlows(t, db)

	flows, err := MonthlyFlows(db, QueryTransactionsFilters{})
	require.NoError(t, err)
	require.Len(t, flows, 2)

	// Sorted chronologically; income is returned positive.
	assert.Equal(t, "2026-01", flows[0].Month)
	assert.InDelta(t, 1000, flows[0].Income, 1e-6)
	assert.InDelta(t, 500, flows[0].Expense, 1e-6)

	assert.Equal(t, "2026-02", flows[1].Month)
	assert.InDelta(t, 1000, flows[1].Income, 1e-6)
	assert.InDelta(t, 700, flows[1].Expense, 1e-6)
}

func TestMonthlyFlowsEmpty(t *testing.T) {
	db := testutil.NewDB(t)
	flows, err := MonthlyFlows(db, QueryTransactionsFilters{})
	require.NoError(t, err)
	assert.Empty(t, flows)
}

func TestSpendingBreakdown(t *testing.T) {
	db := testutil.NewDB(t)
	seedFlows(t, db)

	breakdown, err := SpendingBreakdown(db, QueryTransactionsFilters{})
	require.NoError(t, err)

	assert.InDelta(t, 2000, breakdown.Income, 1e-6)
	assert.InDelta(t, 800, breakdown.Needs, 1e-6)  // fixed both months
	assert.InDelta(t, 400, breakdown.Wants, 1e-6)  // fun both months
	assert.InDelta(t, 800, breakdown.Savings(), 1e-6)
}

func TestSpendingBreakdownRespectsStartDate(t *testing.T) {
	db := testutil.NewDB(t)
	seedFlows(t, db)

	breakdown, err := SpendingBreakdown(db, QueryTransactionsFilters{StartDate: "2026-02-01"})
	require.NoError(t, err)

	assert.InDelta(t, 1000, breakdown.Income, 1e-6)
	assert.InDelta(t, 400, breakdown.Needs, 1e-6)
	assert.InDelta(t, 300, breakdown.Wants, 1e-6)
	assert.InDelta(t, 300, breakdown.Savings(), 1e-6)
}
