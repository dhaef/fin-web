package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"fin-web/internal/model"
	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fullTradeForm() url.Values {
	return url.Values{
		"name":          {"Apple"},
		"ticker":        {"AAPL"},
		"purchase_date": {"2026-01-15"},
		"shares":        {"10"},
		"price":         {"150"},
		"type":          {"buy"},
		"account":       {"schwab"},
	}
}

func TestCreateTradeSuccess(t *testing.T) {
	db := testutil.NewDB(t)
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	require.NoError(t, c.createTrade(rec, newFormRequest("/trades/new", fullTradeForm())))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/trades/1", rec.Header().Get("Location"))

	trades, err := model.GetTrades(db)
	require.NoError(t, err)
	require.Len(t, trades, 1)
	assert.Equal(t, "AAPL", trades[0].Ticker)
	assert.InDelta(t, 10, trades[0].Shares, 1e-9)
}

func TestCreateTradeValidationRendersForm(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}

	form := fullTradeForm()
	form.Del("name")
	form.Set("shares", "notanumber")
	rec := httptest.NewRecorder()
	require.NoError(t, c.createTrade(rec, newFormRequest("/trades/new", form)))

	// Invalid input re-renders the form (HTTP 200) rather than redirecting.
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())

	trades, err := model.GetTrades(c.db)
	require.NoError(t, err)
	assert.Empty(t, trades, "no trade should be created on validation failure")
}

func TestUpdateTradeSuccess(t *testing.T) {
	db := testutil.NewDB(t)
	id, err := model.CreateTrade(db, "Apple", "AAPL", "2026-01-15", 10, 150, "buy", "schwab")
	require.NoError(t, err)
	c := &Controller{db: db}

	form := fullTradeForm()
	form.Set("ticker", "MSFT")
	form.Set("shares", "5")
	req := newFormRequest("/trades/"+strconv.Itoa(id), form)
	req.SetPathValue("id", strconv.Itoa(id))
	rec := httptest.NewRecorder()
	require.NoError(t, c.updateTrade(rec, req))

	assert.Equal(t, http.StatusSeeOther, rec.Code)

	trade, err := model.GetTrade(db, strconv.Itoa(id))
	require.NoError(t, err)
	assert.Equal(t, "MSFT", trade.Ticker)
	assert.InDelta(t, 5, trade.Shares, 1e-9)
}

func TestDeleteTrade(t *testing.T) {
	db := testutil.NewDB(t)
	id, err := model.CreateTrade(db, "Apple", "AAPL", "2026-01-15", 10, 150, "buy", "schwab")
	require.NoError(t, err)
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodPost, "/trades/"+strconv.Itoa(id)+"/delete", nil)
	req.SetPathValue("id", strconv.Itoa(id))
	rec := httptest.NewRecorder()
	require.NoError(t, c.deleteTrade(rec, req))

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/trades", rec.Header().Get("Location"))

	trades, err := model.GetTrades(db)
	require.NoError(t, err)
	assert.Empty(t, trades)
}

func TestTradeRenders(t *testing.T) {
	db := testutil.NewDB(t)
	id, err := model.CreateTrade(db, "Apple", "AAPL", "2026-01-15", 10, 150, "buy", "schwab")
	require.NoError(t, err)
	c := &Controller{db: db}

	req := httptest.NewRequest(http.MethodGet, "/trades/"+strconv.Itoa(id), nil)
	req.SetPathValue("id", strconv.Itoa(id))
	rec := httptest.NewRecorder()
	require.NoError(t, c.trade(rec, req))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "AAPL")
}

func TestNewTradeRenders(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}
	rec := httptest.NewRecorder()
	require.NoError(t, c.newTrade(rec, httptest.NewRequest(http.MethodGet, "/trades/new", nil)))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTradesListUsesCachedPrice(t *testing.T) {
	db := testutil.NewDB(t)
	_, err := model.CreateTrade(db, "Apple", "AAPL", "2026-01-15", 10, 150, "buy", "schwab")
	require.NoError(t, err)
	// Pre-seed the price cache so the handler doesn't reach out to Tiingo.
	require.NoError(t, model.PutKVItem(db, "AAPL", "200", time.Hour))
	c := &Controller{db: db}

	rec := httptest.NewRecorder()
	require.NoError(t, c.trades(rec, httptest.NewRequest(http.MethodGet, "/trades", nil)))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "AAPL")
}

func TestGetOrFetchPriceCacheHit(t *testing.T) {
	db := testutil.NewDB(t)
	require.NoError(t, model.PutKVItem(db, "AAPL", "200", time.Hour))
	c := &Controller{db: db}

	price, err := c.getOrFetchPrice("AAPL")
	require.NoError(t, err)
	assert.InDelta(t, 200, price, 1e-9)
}

func TestEnrichTradeData(t *testing.T) {
	t.Run("positive growth", func(t *testing.T) {
		tr := &model.Trade{Shares: 10, Total: 1000}
		enrichTradeData(tr, 150) // current value 1500 vs 1000 cost -> +50%
		require.NotNil(t, tr.CurrentValue)
		assert.InDelta(t, 1500, *tr.CurrentValue, 1e-9)
		require.NotNil(t, tr.GrowthRate)
		assert.Equal(t, "50.00", *tr.GrowthRate)
		assert.True(t, tr.HasPositiveGrowth)
	})

	t.Run("zero total avoids divide by zero", func(t *testing.T) {
		tr := &model.Trade{Shares: 10, Total: 0}
		enrichTradeData(tr, 150)
		require.NotNil(t, tr.GrowthRate)
		assert.Equal(t, "0.00", *tr.GrowthRate)
		assert.False(t, tr.HasPositiveGrowth)
	})
}
