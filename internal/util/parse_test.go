package util

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
		{name: "plain", input: "12.34", want: 12.34},
		{name: "dollar sign", input: "$12.34", want: 12.34},
		{name: "thousands separator", input: "$1,234.56", want: 1234.56},
		{name: "negative", input: "-$1,000.00", want: -1000},
		{name: "integer", input: "100", want: 100},
		{name: "zero", input: "0", want: 0},
		{name: "empty", input: "", wantErr: true},
		{name: "non-numeric", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAmount(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeMerchant(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "store number stripped", input: "STARBUCKS STORE 123", want: "STARBUCKS STORE"},
		{name: "different store, same key", input: "STARBUCKS STORE 456", want: "STARBUCKS STORE"},
		{name: "lowercased input", input: "Netflix.com", want: "NETFLIX.COM"},
		{name: "star ref mid-string", input: "AMAZON MKTPL*2C4YN4MH3 SEATTLE WA", want: "AMAZON MKTPL"},
		{name: "amzn mktp us ref", input: "AMZN Mktp US*2A3B4C", want: "AMZN MKTP US"},
		{name: "amazon.com ref", input: "Amazon.com*6M9573OI3 Amzn.com/bill WA", want: "AMAZON.COM"},
		{name: "square prefix", input: "SQ *BLUE BOTTLE", want: "BLUE BOTTLE"},
		{name: "paypal prefix", input: "PAYPAL *SPOTIFY", want: "SPOTIFY"},
		{name: "domain drops city+state", input: "DIGITALOCEAN.COM NEW YORK CITY NY", want: "DIGITALOCEAN.COM"},
		{name: "domain drops other city", input: "DIGITALOCEAN.COM BROOMFIELD CO", want: "DIGITALOCEAN.COM"},
		{name: "phone number cut", input: "SPOTIFY 8777781161 NY", want: "SPOTIFY"},
		{name: "ref number then city", input: "SHELL OIL 57544800204 AUSTIN TX", want: "SHELL OIL"},
		{name: "billpay date suffix", input: "PEPCO PAYMENTUS BILLPAY 250830", want: "PEPCO PAYMENTUS BILLPAY"},
		{name: "trailing state", input: "WHOLE FOODS MARKET WA", want: "WHOLE FOODS MARKET"},
		{name: "hash number", input: "SHELL OIL #4021", want: "SHELL OIL"},
		{name: "embedded leading number kept as key", input: "1800FLOWERS 5551212 NY", want: "1800FLOWERS"},
		{name: "collapses spaces", input: "  DOORDASH   DASHPASS  ", want: "DOORDASH DASHPASS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeMerchant(tt.input))
		})
	}
}

func TestNormalizeMerchantGroupsSameVendor(t *testing.T) {
	// The key must be identical across noisy variants of one vendor.
	assert.Equal(t,
		NormalizeMerchant("STARBUCKS STORE 00123"),
		NormalizeMerchant("Starbucks Store 99999"),
	)
	// Different location suffixes on the same domain merchant collapse together.
	assert.Equal(t,
		NormalizeMerchant("DIGITALOCEAN.COM"),
		NormalizeMerchant("DIGITALOCEAN.COM NEW YORK NY"),
	)
}
