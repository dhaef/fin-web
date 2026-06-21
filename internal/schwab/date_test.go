package schwab

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomDateUnmarshalJSON(t *testing.T) {
	t.Run("valid date", func(t *testing.T) {
		var cd CustomDate
		require.NoError(t, json.Unmarshal([]byte(`"02/04/2026"`), &cd))
		assert.Equal(t, "2026-02-04", cd.Format("2006-01-02"))
	})

	t.Run("null leaves zero value", func(t *testing.T) {
		var cd CustomDate
		require.NoError(t, json.Unmarshal([]byte(`"null"`), &cd))
		assert.True(t, cd.Time.IsZero())
	})

	t.Run("invalid format errors", func(t *testing.T) {
		var cd CustomDate
		require.Error(t, json.Unmarshal([]byte(`"2026-02-04"`), &cd))
	})

	t.Run("embedded in struct", func(t *testing.T) {
		var s struct {
			Date CustomDate `json:"Date"`
		}
		require.NoError(t, json.Unmarshal([]byte(`{"Date":"12/31/2025"}`), &s))
		assert.Equal(t, "2025-12-31", s.Date.Format("2006-01-02"))
	})
}
