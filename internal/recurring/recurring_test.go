package recurring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

// monthlyCharges builds n charges roughly one month apart ending `lastDaysAgo`
// days before `now`, all for the same merchant/amount.
func monthlyCharges(name string, amount float64, n, lastDaysAgo int) []Charge {
	var out []Charge
	last := now.AddDate(0, 0, -lastDaysAgo)
	for i := n - 1; i >= 0; i-- {
		d := last.AddDate(0, 0, -30*i)
		out = append(out, Charge{Name: name, Amount: amount, Date: d.Format(dateLayout)})
	}
	return out
}

func merchants(rs []Recurring) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Merchant
	}
	return out
}

func TestDetectClassifies(t *testing.T) {
	var charges []Charge
	charges = append(charges, monthlyCharges("NETFLIX.COM", 15.49, 5, 10)...) // active sub
	charges = append(charges, monthlyCharges("RENT WEB PMTS", 2000, 6, 5)...) // active bill
	charges = append(charges, monthlyCharges("OLD GYM", 40, 4, 500)...)       // canceled

	report := Detect(charges, now)

	require.Len(t, report.Subscriptions, 1)
	assert.Equal(t, "NETFLIX.COM", report.Subscriptions[0].Merchant)
	assert.Equal(t, "monthly", report.Subscriptions[0].Cadence)
	assert.True(t, report.Subscriptions[0].Active)
	assert.Equal(t, "sub", report.Subscriptions[0].Kind)

	require.Len(t, report.Bills, 1)
	assert.Equal(t, "RENT WEB PMTS", report.Bills[0].Merchant)
	assert.Equal(t, "bill", report.Bills[0].Kind)

	assert.Contains(t, merchants(report.Canceled), "OLD GYM")

	// Netflix ~$15.49 * 30.44/30 ≈ 15.72/mo.
	assert.InDelta(t, 15.72, report.MonthlySubTotal, 0.2)
	assert.InDelta(t, 2000, report.MonthlyBillTotal, 70) // rent, monthly-normalized
}

func TestDetectMergesNormalizedVariants(t *testing.T) {
	// Same vendor, noisy location suffixes: must group into one subscription.
	charges := []Charge{
		{Name: "DIGITALOCEAN.COM", Amount: 6.40, Date: "2026-03-06"},
		{Name: "DIGITALOCEAN.COM NEW YORK NY", Amount: 6.40, Date: "2026-04-06"},
		{Name: "DIGITALOCEAN.COM BROOMFIELD CO", Amount: 6.40, Date: "2026-05-06"},
		{Name: "DIGITALOCEAN.COM", Amount: 6.40, Date: "2026-06-06"},
	}
	report := Detect(charges, now)
	require.Len(t, report.Subscriptions, 1)
	assert.Equal(t, "DIGITALOCEAN.COM", report.Subscriptions[0].Merchant)
	assert.Equal(t, 4, report.Subscriptions[0].Count)
}

func TestDetectRejectsIrregular(t *testing.T) {
	// Frequent, uneven, variable-amount charges (groceries) must not detect.
	charges := []Charge{
		{Name: "H-E-B 123", Amount: 42.10, Date: "2026-06-01"},
		{Name: "H-E-B 456", Amount: 8.75, Date: "2026-06-03"},
		{Name: "H-E-B 789", Amount: 91.20, Date: "2026-06-19"},
		{Name: "H-E-B 111", Amount: 15.00, Date: "2026-06-25"},
	}
	report := Detect(charges, now)
	assert.Empty(t, report.Subscriptions)
	assert.Empty(t, report.Bills)
}

func TestDetectPossibleTwoCharges(t *testing.T) {
	// Two same-priced monthly charges: below MinOccurrences, surfaced as
	// low-confidence rather than missed.
	charges := []Charge{
		{Name: "Spotify USA New York NY", Amount: 11.99, Date: "2026-05-06"},
		{Name: "Spotify USA New York NY", Amount: 11.99, Date: "2026-06-06"},
	}
	report := Detect(charges, now)
	assert.Empty(t, report.Subscriptions)
	require.Len(t, report.Possible, 1)
	assert.Equal(t, "monthly", report.Possible[0].Cadence)
}

func TestDetectNextExpected(t *testing.T) {
	charges := monthlyCharges("NETFLIX.COM", 15.49, 4, 0) // last charge == now
	report := Detect(charges, now)
	require.Len(t, report.Subscriptions, 1)
	// Next expected ~30 days after the last (today).
	next, err := time.Parse(dateLayout, report.Subscriptions[0].Next)
	require.NoError(t, err)
	assert.InDelta(t, 30, next.Sub(now).Hours()/24, 1)
}
