package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := encode(rec, req, http.StatusCreated, map[string]string{"redirect": "/categories/1"})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "/categories/1", body["redirect"])
}

func TestDecode(t *testing.T) {
	t.Run("valid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"value":"starbucks"}`))
		got, err := decode[CategoryValueFormItem](req)
		require.NoError(t, err)
		assert.Equal(t, "starbucks", got.Value)
	})

	t.Run("malformed body errors", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{not json`))
		_, err := decode[CategoryValueFormItem](req)
		require.Error(t, err)
	})
}
