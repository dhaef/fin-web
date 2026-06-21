package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"fin-web/internal/model"
	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fullNetWorthForm returns a complete, valid net-worth form. Individual tests
// override fields to exercise validation branches.
func fullNetWorthForm() url.Values {
	return url.Values{
		"date":       {"2026-02-01"},
		"cash":       {"100"},
		"investment": {"200"},
		"debit":      {"-50"},
		"credit":     {"-25"},
		"savings":    {"300"},
		"retirement": {"400"},
		"loans":      {"-1000"},
	}
}

// fullNetWorthParams returns params with every numeric column set. The model's
// scanners reject NULL float columns, so seeded rows must populate all fields
// (production always does, via the form).
func fullNetWorthParams(date string, cash float32) model.NetWorthItemParams {
	return model.NetWorthItemParams{
		Date:       ToPtr(date),
		Cash:       ToPtr(cash),
		Investment: ToPtr(float32(0)),
		Debit:      ToPtr(float32(0)),
		Credit:     ToPtr(float32(0)),
		Savings:    ToPtr(float32(0)),
		Retirement: ToPtr(float32(0)),
		Loans:      ToPtr(float32(0)),
	}
}

func TestValidateNetWorthForm(t *testing.T) {
	t.Run("valid form has no errors", func(t *testing.T) {
		req := newFormRequest("/net-worth/new", fullNetWorthForm())
		params, errs := validateNetWorthForm(req)
		assert.Empty(t, errs)
		require.NotNil(t, params.Cash)
		assert.InDelta(t, 100, *params.Cash, 1e-6)
		assert.Equal(t, "2026-02-01", *params.Date)
	})

	t.Run("empty fields produce errors", func(t *testing.T) {
		form := fullNetWorthForm()
		form.Del("cash")
		form.Del("loans")
		req := newFormRequest("/net-worth/new", form)
		_, errs := validateNetWorthForm(req)
		assert.Contains(t, errs, "cash")
		assert.Contains(t, errs, "loans")
		assert.NotContains(t, errs, "savings")
	})

	t.Run("non-numeric value produces error", func(t *testing.T) {
		form := fullNetWorthForm()
		form.Set("cash", "abc")
		req := newFormRequest("/net-worth/new", form)
		_, errs := validateNetWorthForm(req)
		assert.Contains(t, errs, "cash")
	})
}

func TestCreateNetWorthItemSuccess(t *testing.T) {
	db := testutil.NewDB(t)
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	require.NoError(t, c.createNetWorthItem(rec, newFormRequest("/net-worth/new", fullNetWorthForm())))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/net-worth", rec.Header().Get("Location"))

	items, err := model.QueryNetWorthItems(db, model.QueryNetWorthItemsFilters{})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.InDelta(t, 100, items[0].Cash, 1e-6)
}

func TestCreateNetWorthItemValidationRendersForm(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}

	form := fullNetWorthForm()
	form.Del("cash")
	rec := httptest.NewRecorder()
	require.NoError(t, c.createNetWorthItem(rec, newFormRequest("/net-worth/new", form)))

	// Validation failures re-render the form (HTTP 200), not a redirect.
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
}

func TestUpdateNetWorthItemSuccess(t *testing.T) {
	db := testutil.NewDB(t)
	id, err := model.CreateNetWorthItem(db, fullNetWorthParams("2026-01-01", 1))
	require.NoError(t, err)
	c := &Controller{db: db}

	req := newFormRequest("/net-worth/"+id, fullNetWorthForm())
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	require.NoError(t, c.updateNetWorthItem(rec, req))

	assert.Equal(t, http.StatusSeeOther, rec.Code)

	item, err := model.GetNetWorthItem(db, id)
	require.NoError(t, err)
	assert.InDelta(t, 100, item.Cash, 1e-6)
	assert.InDelta(t, 200, item.Investment, 1e-6)
}

func TestUpdateNetWorthItemMissingReturnsError(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}

	req := newFormRequest("/net-worth/nope", fullNetWorthForm())
	req.SetPathValue("id", "nope")
	rec := httptest.NewRecorder()

	err := c.updateNetWorthItem(rec, req)
	var apiErr APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusInternalServerError, apiErr.Status)
}

func TestDeleteNetWorthItem(t *testing.T) {
	db := testutil.NewDB(t)
	id, err := model.CreateNetWorthItem(db, model.NetWorthItemParams{Date: ToPtr("2026-01-01")})
	require.NoError(t, err)
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodPost, "/net-worth/"+id+"/delete", nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	require.NoError(t, c.deleteNetWorthItem(rec, req))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/net-worth", rec.Header().Get("Location"))

	items, err := model.QueryNetWorthItems(db, model.QueryNetWorthItemsFilters{})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestNetWorthListComputesChange(t *testing.T) {
	db := testutil.NewDB(t)
	// Older item: net worth 100. Newer item: net worth 150 -> +50 (50%).
	_, err := model.CreateNetWorthItem(db, fullNetWorthParams("2026-01-01", 100))
	require.NoError(t, err)
	_, err = model.CreateNetWorthItem(db, fullNetWorthParams("2026-02-01", 150))
	require.NoError(t, err)
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	require.NoError(t, c.netWorth(rec, httptest.NewRequest(http.MethodGet, "/net-worth", nil)))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "50.00%")
}

func TestNewNetWorthItemRenders(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}
	rec := httptest.NewRecorder()
	require.NoError(t, c.newNetWorthItem(rec, httptest.NewRequest(http.MethodGet, "/net-worth/new", nil)))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
}

func TestNetWorthItemRenders(t *testing.T) {
	db := testutil.NewDB(t)
	id, err := model.CreateNetWorthItem(db, fullNetWorthParams("2026-02-01", 123))
	require.NoError(t, err)
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodGet, "/net-worth/"+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	require.NoError(t, c.netWorthItem(rec, req))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPointerHelpers(t *testing.T) {
	assert.Equal(t, "", ptrToString(nil))
	assert.Equal(t, "hi", ptrToString(ToPtr("hi")))
	assert.Equal(t, "", ptrToStringFloat32(nil))
	assert.Equal(t, "1.500000", ptrToStringFloat32(ToPtr(float32(1.5))))
}
