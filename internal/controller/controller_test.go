package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTemplatePaths(t *testing.T) {
	got := buildTemplatePaths([]string{"categories/category.html", "layout.html"})
	assert.Equal(t, []string{"files/categories/category.html", "files/layout.html"}, got)
}

func TestMakeHandlerSuccess(t *testing.T) {
	called := false
	h := MakeHandler(func(w http.ResponseWriter, r *http.Request) error {
		called = true
		w.WriteHeader(http.StatusOK)
		return nil
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMakeHandlerJSONError(t *testing.T) {
	h := MakeHandler(func(w http.ResponseWriter, r *http.Request) error {
		return APIError{
			Status:       http.StatusBadRequest,
			Message:      "bad input",
			ResponseType: "JSON",
		}
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "bad input", body["message"])
}

func TestMakeHandlerHTMLError(t *testing.T) {
	h := MakeHandler(func(w http.ResponseWriter, r *http.Request) error {
		return APIError{
			Status:  http.StatusNotFound,
			Message: "missing",
		}
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// Non-JSON APIErrors fall back to rendering the not-found page (HTTP 200).
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.String())
}

func TestMakeHandlerNonAPIErrorIsSwallowed(t *testing.T) {
	// A plain (non-APIError) error fails the type assertion and produces no
	// response body — this documents the current behavior.
	h := MakeHandler(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("boom")
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}
