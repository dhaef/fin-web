package citi

import (
	"testing"

	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	testutil.SeedCategory(t, db, "Coffee", 5, "starbucks")

	p := NewCitiProvider(db)
	txns, err := p.ParseFile("testdata/sample.csv")
	require.NoError(t, err)
	require.Len(t, txns, 2)

	// Debit row, categorized via merchant name.
	debit := txns[0]
	assert.Equal(t, "STARBUCKS STORE 123", debit.Name)
	assert.Equal(t, "citi", debit.Source)
	assert.Equal(t, "citi", debit.Account)
	assert.Equal(t, "2026-02-04", debit.Date)
	assert.InDelta(t, 5.75, debit.Amount, 1e-9)
	assert.Equal(t, "Dining", debit.Category)
	assert.True(t, debit.CategoryID.Valid, "starbucks should match Coffee category")

	// Credit row.
	credit := txns[1]
	assert.Equal(t, "PAYMENT THANK YOU", credit.Name)
	assert.Equal(t, "2026-02-10", credit.Date)
	assert.InDelta(t, 200.00, credit.Amount, 1e-9)
	assert.Equal(t, "Payment", credit.Category)
	assert.False(t, credit.CategoryID.Valid)
}

func TestParseFileMissingFile(t *testing.T) {
	p := NewCitiProvider(testutil.NewCategoryDB(t))
	_, err := p.ParseFile("testdata/does-not-exist.csv")
	require.Error(t, err)
}

func TestGetPrefix(t *testing.T) {
	assert.Equal(t, "From", NewCitiProvider(nil).GetPrefix())
}
