package bofa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		// BofA reports spend as negative; parseAmount flips the sign so a
		// debit becomes a positive transaction amount.
		{name: "debit flips to positive", input: "-25.00", want: 25},
		{name: "credit flips to negative", input: "100.00", want: -100},
		{name: "zero", input: "0", want: 0},
		{name: "non-numeric", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAmount(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
