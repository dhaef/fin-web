// Package recurring detects recurring charges (subscriptions and bills) from a
// flat list of expense transactions. It groups charges by normalized merchant,
// measures cadence regularity and amount stability, and classifies each group
// as an active subscription, a recurring bill, a canceled series, or a
// low-confidence "possible" match. The logic is pure and DB-free so it can be
// unit-tested and reused by both the web page and the calibration tool.
package recurring

import (
	"math"
	"sort"
	"time"

	"fin-web/internal/util"
)

// Detection thresholds. These were tuned against real transaction data with the
// cmd/subsreport calibration tool.
const (
	// MinOccurrences is how many charges a merchant needs before we'll call it
	// recurring with confidence.
	MinOccurrences = 3
	// GapCVMax is the maximum coefficient of variation on the day-gaps between
	// charges for the cadence to count as regular.
	GapCVMax = 0.25
	// AmountCVMax is the CV under which a charge is considered fixed-price.
	AmountCVMax = 0.10
	// BillMonthlyMin splits large recurring bills (rent, mortgage) from
	// subscriptions by monthly-normalized cost.
	BillMonthlyMin = 150.0
	// staleFactor: a series is "canceled" once the last charge is older than
	// this many cadence-lengths.
	staleFactor = 1.5

	dateLayout = "2006-01-02"
)

// Charge is one expense transaction fed into detection.
type Charge struct {
	Name   string
	Amount float64
	Date   string // "2006-01-02"
}

// Recurring is a detected recurring charge for one merchant.
type Recurring struct {
	Merchant    string
	Cadence     string // weekly | biweekly | monthly | quarterly | annual
	Count       int
	TypicalAmt  float64
	AmountFixed bool
	Monthly     float64 // amount normalized to a monthly cost
	Annual      float64
	Last        string // date of the most recent charge
	Next        string // projected next charge date
	Active      bool
	Kind        string // "sub" | "bill"
}

// Report is the full detection result, bucketed for presentation.
type Report struct {
	Subscriptions    []Recurring // active, subscription-sized
	Bills            []Recurring // active, bill-sized (rent/utilities)
	Canceled         []Recurring // regular cadence but no recent charge
	Possible         []Recurring // only 2 charges: low-confidence new/annual
	MonthlySubTotal  float64
	MonthlyBillTotal float64
}

type parsedCharge struct {
	date   time.Time
	amount float64
}

// Detect groups charges by normalized merchant and classifies each group.
// `now` anchors the active/canceled decision (pass time.Now()).
func Detect(charges []Charge, now time.Time) Report {
	groups := map[string][]parsedCharge{}
	for _, c := range charges {
		d, err := time.Parse(dateLayout, c.Date)
		if err != nil {
			continue
		}
		key := util.NormalizeMerchant(c.Name)
		if key == "" {
			continue
		}
		groups[key] = append(groups[key], parsedCharge{date: d, amount: c.Amount})
	}

	var report Report
	for key, pcs := range groups {
		if len(pcs) < 2 {
			continue
		}
		r, gapCV, cadenceOK := analyze(key, pcs, now)
		regular := cadenceOK && gapCV <= GapCVMax

		switch {
		case r.Count >= MinOccurrences && regular:
			switch {
			case !r.Active:
				report.Canceled = append(report.Canceled, r)
			case r.Kind == "bill":
				report.Bills = append(report.Bills, r)
				report.MonthlyBillTotal += r.Monthly
			default:
				report.Subscriptions = append(report.Subscriptions, r)
				report.MonthlySubTotal += r.Monthly
			}
		case r.Count == 2 && r.AmountFixed && r.Active && isSlowCadence(r.Cadence):
			// Two same-priced charges a month/quarter/year apart: not enough to
			// confirm, but worth surfacing (catches brand-new or annual subs).
			report.Possible = append(report.Possible, r)
		}
	}

	byAnnualDesc := func(rs []Recurring) {
		sort.Slice(rs, func(i, j int) bool { return rs[i].Annual > rs[j].Annual })
	}
	byAnnualDesc(report.Subscriptions)
	byAnnualDesc(report.Bills)
	byAnnualDesc(report.Canceled)
	byAnnualDesc(report.Possible)

	return report
}

func analyze(key string, pcs []parsedCharge, now time.Time) (Recurring, float64, bool) {
	sort.Slice(pcs, func(i, j int) bool { return pcs[i].date.Before(pcs[j].date) })

	gaps := make([]float64, 0, len(pcs)-1)
	for i := 1; i < len(pcs); i++ {
		gaps = append(gaps, pcs[i].date.Sub(pcs[i-1].date).Hours()/24)
	}
	amounts := make([]float64, len(pcs))
	for i, p := range pcs {
		amounts[i] = p.amount
	}

	medGap := median(gaps)
	medAmt := median(amounts)
	cadence := classifyCadence(medGap)
	last := pcs[len(pcs)-1].date

	r := Recurring{
		Merchant:    key,
		Cadence:     cadence,
		Count:       len(pcs),
		TypicalAmt:  medAmt,
		AmountFixed: cv(amounts) <= AmountCVMax,
		Last:        last.Format(dateLayout),
		Kind:        "sub",
	}
	if medGap > 0 {
		r.Monthly = medAmt * 30.44 / medGap
		r.Annual = medAmt * 365.25 / medGap
		r.Next = last.AddDate(0, 0, int(math.Round(medGap))).Format(dateLayout)
		r.Active = now.Sub(last).Hours()/24 <= staleFactor*medGap
	}
	if r.Monthly >= BillMonthlyMin {
		r.Kind = "bill"
	}

	return r, cv(gaps), cadence != "irregular"
}

func classifyCadence(medGap float64) string {
	switch {
	case medGap >= 6 && medGap <= 8:
		return "weekly"
	case medGap >= 12 && medGap <= 16:
		return "biweekly"
	case medGap >= 26 && medGap <= 35:
		return "monthly"
	case medGap >= 82 && medGap <= 100:
		return "quarterly"
	case medGap >= 350 && medGap <= 380:
		return "annual"
	default:
		return "irregular"
	}
}

// isSlowCadence reports cadences slow enough that a 2-charge match is worth
// surfacing (fast cadences are too easily coincidental).
func isSlowCadence(cadence string) bool {
	switch cadence {
	case "monthly", "quarterly", "annual":
		return true
	default:
		return false
	}
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	mid := len(s) / 2
	if len(s)%2 == 0 {
		return (s[mid-1] + s[mid]) / 2
	}
	return s[mid]
}

// cv is the coefficient of variation (stddev/mean); 0 means perfectly regular.
func cv(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	if mean == 0 {
		return 0
	}
	var sq float64
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	return math.Sqrt(sq/float64(len(xs))) / mean
}
