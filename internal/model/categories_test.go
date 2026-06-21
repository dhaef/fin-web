package model

import (
	"database/sql"
	"strconv"
	"testing"

	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndGetCategory(t *testing.T) {
	db := testutil.NewCategoryDB(t)

	id, err := CreateCategory(db, "Groceries", 10, "fixed", false)
	require.NoError(t, err)
	require.Positive(t, id)

	valID, err := CreateCategoryValue(db, id, "whole foods")
	require.NoError(t, err)
	require.Positive(t, valID)

	cat, err := GetCategory(db, strconv.Itoa(id))
	require.NoError(t, err)
	assert.Equal(t, "Groceries", cat.Label)
	assert.Equal(t, 10, cat.Priority)
	assert.True(t, cat.Type.Valid)
	assert.Equal(t, "fixed", cat.Type.String)
	require.Len(t, cat.Values, 1)
	assert.Equal(t, "whole foods", cat.Values[0].Value.String)
}

func TestCreateCategoryDuplicatePriorityRejected(t *testing.T) {
	db := testutil.NewCategoryDB(t)

	_, err := CreateCategory(db, "A", 1, "fun", false)
	require.NoError(t, err)

	_, err = CreateCategory(db, "B", 1, "fun", false)
	require.Error(t, err, "duplicate priority should be rejected by trigger")
}

func TestGetCategoriesOrderedByPriority(t *testing.T) {
	db := testutil.NewCategoryDB(t)

	mustCreateCategory(t, db, "Third", 30, "fun")
	mustCreateCategory(t, db, "First", 10, "fun")
	mustCreateCategory(t, db, "Second", 20, "fun")

	cats, err := GetCategories(db)
	require.NoError(t, err)
	require.Len(t, cats, 3)

	labels := []string{cats[0].Label, cats[1].Label, cats[2].Label}
	assert.Equal(t, []string{"First", "Second", "Third"}, labels)
}

func TestSearchCategories(t *testing.T) {
	db := testutil.NewCategoryDB(t)

	groceries := mustCreateCategory(t, db, "Groceries", 10, "fixed")
	coffee := mustCreateCategory(t, db, "Coffee", 5, "fun")
	mustCreateCategoryValue(t, db, groceries, "whole foods")
	mustCreateCategoryValue(t, db, coffee, "starbucks")

	t.Run("matches when query contains value as substring", func(t *testing.T) {
		got, err := SearchCategories(db, []string{"starbucks store #1234"})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "Coffee", got[0].Label)
	})

	t.Run("no match returns empty", func(t *testing.T) {
		got, err := SearchCategories(db, []string{"unknown merchant"})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("multiple queries ored together, ordered by priority", func(t *testing.T) {
		got, err := SearchCategories(db, []string{"whole foods market", "starbucks"})
		require.NoError(t, err)
		require.Len(t, got, 2)
		// Coffee has priority 5, Groceries 10, so Coffee comes first.
		assert.Equal(t, "Coffee", got[0].Label)
		assert.Equal(t, "Groceries", got[1].Label)
	})

	t.Run("empty query slice errors", func(t *testing.T) {
		_, err := SearchCategories(db, []string{})
		require.Error(t, err)
	})
}

func TestUpdateCategory(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	id := mustCreateCategory(t, db, "Old", 10, "fun")

	newLabel := "New"
	newPriority := 99
	require.NoError(t, UpdateCategory(db, strconv.Itoa(id), UpdateCategoryParams{
		Label:    &newLabel,
		Priority: &newPriority,
	}))

	cat, err := GetCategory(db, strconv.Itoa(id))
	require.NoError(t, err)
	assert.Equal(t, "New", cat.Label)
	assert.Equal(t, 99, cat.Priority)
}

func TestUpdateCategoryNoFieldsIsNoop(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	id := mustCreateCategory(t, db, "Stable", 10, "fun")

	require.NoError(t, UpdateCategory(db, strconv.Itoa(id), UpdateCategoryParams{}))

	cat, err := GetCategory(db, strconv.Itoa(id))
	require.NoError(t, err)
	assert.Equal(t, "Stable", cat.Label)
	assert.Equal(t, 10, cat.Priority)
}

func TestDeleteCategory(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	id := mustCreateCategory(t, db, "Doomed", 10, "fun")

	require.NoError(t, DeleteCategory(db, strconv.Itoa(id)))

	cats, err := GetCategories(db)
	require.NoError(t, err)
	assert.Empty(t, cats)
}

func TestCategoryValueLifecycle(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	catID := mustCreateCategory(t, db, "Shopping", 10, "fun")
	valID := mustCreateCategoryValue(t, db, catID, "amazon")

	newVal := "amazon.com"
	require.NoError(t, UpdateCategoryValue(db, strconv.Itoa(valID), UpdateCategoryValueParams{Value: &newVal}))

	cat, err := GetCategory(db, strconv.Itoa(catID))
	require.NoError(t, err)
	require.Len(t, cat.Values, 1)
	assert.Equal(t, "amazon.com", cat.Values[0].Value.String)

	require.NoError(t, DeleteCategoryValue(db, valID))

	cat, err = GetCategory(db, strconv.Itoa(catID))
	require.NoError(t, err)
	assert.Empty(t, cat.Values)
}

// --- helpers ---

func mustCreateCategory(t *testing.T, db *sql.DB, label string, priority int, catType string) int {
	t.Helper()
	id, err := CreateCategory(db, label, priority, catType, false)
	require.NoError(t, err)
	return id
}

func mustCreateCategoryValue(t *testing.T, db *sql.DB, categoryID int, value string) int {
	t.Helper()
	id, err := CreateCategoryValue(db, categoryID, value)
	require.NoError(t, err)
	return id
}
