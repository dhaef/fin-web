package controller

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fin-web/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionsHandlerDetects(t *testing.T) {
	db := testutil.NewDB(t)

	// Five monthly Netflix charges ending recently -> active subscription.
	last := time.Now().AddDate(0, 0, -5)
	for i := 4; i >= 0; i-- {
		d := last.AddDate(0, 0, -30*i).Format("2006-01-02")
		seedTransaction(t, db, "nf-"+d, "NETFLIX.COM", 15.49, d, sql.NullInt32{})
	}

	c := &Controller{db: db}
	rec := httptest.NewRecorder()
	require.NoError(t, c.subscriptions(rec, httptest.NewRequest(http.MethodGet, "/subscriptions", nil)))

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "NETFLIX.COM")
	assert.Contains(t, body, "Active Subscriptions")
}

func TestSubscriptionsHandlerEmpty(t *testing.T) {
	c := &Controller{db: testutil.NewDB(t)}
	rec := httptest.NewRecorder()
	require.NoError(t, c.subscriptions(rec, httptest.NewRequest(http.MethodGet, "/subscriptions", nil)))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "None detected")
}
