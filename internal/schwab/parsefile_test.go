package schwab

import (
	"testing"

	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	testutil.SeedCategory(t, db, "Coffee", 5, "starbucks")

	p := NewSchwabProvider(db)
	txns, err := p.ParseFile("testdata/sample.json")
	require.NoError(t, err)
	require.Len(t, txns, 2)

	// Withdrawal -> positive amount, categorized.
	withdrawal := txns[0]
	assert.Equal(t, "STARBUCKS STORE 123", withdrawal.Name)
	assert.Equal(t, "schwab", withdrawal.Source)
	assert.Equal(t, "schwab", withdrawal.Account)
	assert.Equal(t, "2026-02-04", withdrawal.Date)
	assert.InDelta(t, 5.75, withdrawal.Amount, 1e-9)
	assert.True(t, withdrawal.CategoryID.Valid, "starbucks should be categorized")

	// Deposit -> negated amount (and comma-separated value parsed).
	deposit := txns[1]
	assert.Equal(t, "PAYROLL DEPOSIT", deposit.Name)
	assert.Equal(t, "2026-02-05", deposit.Date)
	assert.InDelta(t, -2500.00, deposit.Amount, 1e-9)
	assert.False(t, deposit.CategoryID.Valid)
}

func TestParseFileMissingFile(t *testing.T) {
	p := NewSchwabProvider(testutil.NewCategoryDB(t))
	_, err := p.ParseFile("testdata/does-not-exist.json")
	require.Error(t, err)
}

func TestParseFileInvalidJSON(t *testing.T) {
	p := NewSchwabProvider(testutil.NewCategoryDB(t))
	_, err := p.ParseFile("testdata/invalid.txt")
	require.Error(t, err)
}

func TestGetPrefix(t *testing.T) {
	assert.Equal(t, "Checking", NewSchwabProvider(nil).GetPrefix())
}
