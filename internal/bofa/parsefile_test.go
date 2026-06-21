package bofa

import (
	"testing"

	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	db := testutil.NewDB(t)
	testutil.SeedCategory(t, db, "Coffee", 5, "starbucks")

	p := NewBofaProvider(db)
	txns, err := p.ParseFile("testdata/sample.csv")
	require.NoError(t, err)
	require.Len(t, txns, 3)

	// Row 1: STARBUCKS debit of -5.75 -> amount flips to +5.75, categorized.
	starbucks := txns[0]
	assert.Equal(t, "STARBUCKS STORE 123", starbucks.Name)
	assert.Equal(t, "bank_of_america", starbucks.Source)
	assert.Equal(t, "bank_of_america", starbucks.Account)
	assert.Equal(t, "2026-02-04", starbucks.Date)
	assert.InDelta(t, 5.75, starbucks.Amount, 1e-9)
	assert.True(t, starbucks.CategoryID.Valid, "starbucks should be categorized")

	// Row 2: deposit -> negative amount, no matching category.
	deposit := txns[1]
	assert.InDelta(t, -2500.00, deposit.Amount, 1e-9)
	assert.False(t, deposit.CategoryID.Valid)

	// Row 3: uncategorized debit.
	wholeFoods := txns[2]
	assert.InDelta(t, 42.10, wholeFoods.Amount, 1e-9)
	assert.False(t, wholeFoods.CategoryID.Valid)

	// Every transaction gets a unique generated id.
	assert.NotEmpty(t, starbucks.ID)
	assert.NotEqual(t, starbucks.ID, deposit.ID)
}

func TestParseFileMissingFile(t *testing.T) {
	p := NewBofaProvider(testutil.NewDB(t))
	_, err := p.ParseFile("testdata/does-not-exist.csv")
	require.Error(t, err)
}

func TestGetPrefix(t *testing.T) {
	assert.Equal(t, "bofa", NewBofaProvider(nil).GetPrefix())
}
