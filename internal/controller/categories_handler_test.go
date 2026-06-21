package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"fin-web/internal/model"
	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFormRequest builds a form-encoded POST request the way the handlers
// expect to read it via r.FormValue.
func newFormRequest(target string, form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func decodeErrs(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body struct {
		Errs map[string]string `json:"errs"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body.Errs
}

func TestCategoriesHandlerRenders(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	testutil.SeedCategory(t, db, "Coffee", 5, "starbucks")
	testutil.SeedCategory(t, db, "Groceries", 10, "whole foods")
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	err := c.categories(rec, httptest.NewRequest(http.MethodGet, "/categories", nil))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Coffee")
	assert.Contains(t, body, "Groceries")
}

func TestCategoryHandlerRenders(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	id := testutil.SeedCategory(t, db, "Coffee", 5, "starbucks")
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodGet, "/categories/"+strconv.Itoa(id), nil)
	req.SetPathValue("id", strconv.Itoa(id))

	rec := httptest.NewRecorder()
	require.NoError(t, c.category(rec, req))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Coffee")
}

func TestNewCategoryHandlerRenders(t *testing.T) {
	c := &Controller{db: testutil.NewCategoryDB(t)}

	rec := httptest.NewRecorder()
	require.NoError(t, c.newCategory(rec, httptest.NewRequest(http.MethodGet, "/categories/new", nil)))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
}

func TestCreateCategorySuccess(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	err := c.createCategory(rec, newFormRequest("/categories/new", url.Values{
		"label":      {"Coffee"},
		"priority":   {"5"},
		"type":       {"fun"},
		"is_ignored": {"false"},
		"values":     {`[{"value":"starbucks"}]`},
	}))
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "/categories/1", body["redirect"])

	// Category and its value should be persisted.
	cat, err := model.GetCategory(db, "1")
	require.NoError(t, err)
	assert.Equal(t, "Coffee", cat.Label)
	require.Len(t, cat.Values, 1)
	assert.Equal(t, "starbucks", cat.Values[0].Value.String)
}

func TestCreateCategoryValidationErrors(t *testing.T) {
	c := &Controller{db: testutil.NewCategoryDB(t)}

	rec := httptest.NewRecorder()
	err := c.createCategory(rec, newFormRequest("/categories/new", url.Values{
		"label":      {""},
		"priority":   {""},
		"type":       {""},
		"is_ignored": {"false"},
		"values":     {`[]`},
	}))
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	errs := decodeErrs(t, rec)
	assert.Contains(t, errs, "label")
	assert.Contains(t, errs, "type")
	assert.Contains(t, errs, "priority")
}

func TestCreateCategoryEmptyValueRejected(t *testing.T) {
	c := &Controller{db: testutil.NewCategoryDB(t)}

	rec := httptest.NewRecorder()
	err := c.createCategory(rec, newFormRequest("/categories/new", url.Values{
		"label":      {"Coffee"},
		"priority":   {"5"},
		"type":       {"fun"},
		"is_ignored": {"false"},
		"values":     {`[{"value":""}]`},
	}))
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, decodeErrs(t, rec), "values")
}

func TestCreateCategoryMalformedValuesJSON(t *testing.T) {
	c := &Controller{db: testutil.NewCategoryDB(t)}

	rec := httptest.NewRecorder()
	err := c.createCategory(rec, newFormRequest("/categories/new", url.Values{
		"label":      {"Coffee"},
		"priority":   {"5"},
		"type":       {"fun"},
		"is_ignored": {"false"},
		"values":     {`not json`},
	}))

	// Malformed values JSON surfaces as a returned APIError, not a written body.
	var apiErr APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
}

func TestCreateCategoryDuplicatePriority(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	testutil.SeedCategory(t, db, "Existing", 5, "foo")
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	err := c.createCategory(rec, newFormRequest("/categories/new", url.Values{
		"label":      {"Coffee"},
		"priority":   {"5"},
		"type":       {"fun"},
		"is_ignored": {"false"},
		"values":     {`[{"value":"starbucks"}]`},
	}))
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	errs := decodeErrs(t, rec)
	assert.Equal(t, "This priority is already taken", errs["priority"])
}

func TestDeleteCategory(t *testing.T) {
	db := testutil.NewCategoryDB(t)
	id := testutil.SeedCategory(t, db, "Doomed", 5, "foo")
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodPost, "/categories/"+strconv.Itoa(id)+"/delete", nil)
	req.SetPathValue("id", strconv.Itoa(id))

	rec := httptest.NewRecorder()
	require.NoError(t, c.deleteCategory(rec, req))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/categories", rec.Header().Get("Location"))

	cats, err := model.GetCategories(db)
	require.NoError(t, err)
	assert.Empty(t, cats)
}
