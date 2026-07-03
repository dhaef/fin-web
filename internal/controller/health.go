package controller

import (
	"fmt"
	"net/http"

	"fin-web/internal/model"
)

// windowMonths is the trailing window used for the 50/30/20 breakdown.
const windowMonths = 12

type HealthPage struct {
	Breakdown      model.Breakdown
	BreakdownDonut []BreakdownSlice
	NeedsPct       string
	WantsPct       string
	SavingsPct     string
	HasIncome      bool
	WindowMonths   int
	SavingsRows    []SavingsRow
}

// BreakdownSlice is one wedge of the needs/wants/savings donut. ID is unused by
// the chart (navigation is disabled) but kept so the JSON shape matches the
// other donut feeds.
type BreakdownSlice struct {
	Name  string
	Value float64
	ID    int
}

// SavingsRow is one month of the savings-rate table, including trailing
// rolling rates. Rate strings are preformatted ("42.1%" or "—" when there is
// no income to divide by).
type SavingsRow struct {
	Month     string
	Income    float64
	Expense   float64
	Savings   float64
	Rate      string
	Rolling3  string
	Rolling12 string
}

func (c *Controller) health(w http.ResponseWriter, r *http.Request) error {
	flows, err := model.MonthlyFlows(c.db, model.QueryTransactionsFilters{})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching monthly flows: " + err.Error(),
		}
	}

	// Rows are computed oldest-first (rolling windows look backwards) but shown
	// newest-first.
	rows := computeSavingsRows(flows)
	display := make([]SavingsRow, len(rows))
	for i, row := range rows {
		display[len(rows)-1-i] = row
	}

	breakdown, err := model.SpendingBreakdown(c.db, model.QueryTransactionsFilters{
		StartDate: trailingWindowStart(flows, windowMonths),
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching spending breakdown: " + err.Error(),
		}
	}

	savings := breakdown.Savings()
	// A deficit can't be drawn as a wedge; clamp so the donut still renders the
	// needs/wants split. The true (possibly negative) figure shows in the text.
	donutSavings := savings
	if donutSavings < 0 {
		donutSavings = 0
	}

	page := HealthPage{
		Breakdown: breakdown,
		BreakdownDonut: []BreakdownSlice{
			{Name: "Needs", Value: breakdown.Needs, ID: 0},
			{Name: "Wants", Value: breakdown.Wants, ID: 1},
			{Name: "Savings", Value: donutSavings, ID: 2},
		},
		NeedsPct:     formatPct(breakdown.Needs, breakdown.Income),
		WantsPct:     formatPct(breakdown.Wants, breakdown.Income),
		SavingsPct:   formatPct(savings, breakdown.Income),
		HasIncome:    breakdown.Income > 0,
		WindowMonths: windowMonths,
		SavingsRows:  display,
	}

	if err := renderTemplate(w, Base[HealthPage]{Data: page}, "layout", []string{"health.html", "layout.html"}); err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}

// computeSavingsRows turns a chronological flow series into per-month savings
// rows, including trailing 3- and 12-month rolling rates. Rolling rates
// aggregate income and expense over the window before dividing, so months are
// weighted by size rather than averaging noisy per-month percentages.
func computeSavingsRows(flows []model.MonthlyFlow) []SavingsRow {
	rows := make([]SavingsRow, len(flows))
	for i, flow := range flows {
		rows[i] = SavingsRow{
			Month:     flow.Month,
			Income:    flow.Income,
			Expense:   flow.Expense,
			Savings:   flow.Income - flow.Expense,
			Rate:      formatSavingsRate(flow.Income, flow.Expense),
			Rolling3:  rollingRate(flows, i, 3),
			Rolling12: rollingRate(flows, i, 12),
		}
	}
	return rows
}

// rollingRate is the savings rate over the trailing window ending at index end
// (inclusive), spanning at most window months.
func rollingRate(flows []model.MonthlyFlow, end, window int) string {
	start := max(end-window+1, 0)

	var income, expense float64
	for _, flow := range flows[start : end+1] {
		income += flow.Income
		expense += flow.Expense
	}

	return formatSavingsRate(income, expense)
}

// formatSavingsRate renders (income-expense)/income as a percentage, or "—"
// when there is no income to divide by.
func formatSavingsRate(income, expense float64) string {
	if income <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f%%", (income-expense)/income*100)
}

// formatPct renders part/whole as a percentage, or "—" when whole is not
// positive.
func formatPct(part, whole float64) string {
	if whole <= 0 {
		return "—"
	}
	return fmt.Sprintf("%.1f%%", part/whole*100)
}

// trailingWindowStart returns the "YYYY-MM-01" start date of the trailing
// window covering the last `months` months of data, suitable as an inclusive
// StartDate filter. It returns "" (no lower bound) when there is no data.
func trailingWindowStart(flows []model.MonthlyFlow, months int) string {
	if len(flows) == 0 {
		return ""
	}

	start := max(len(flows)-months, 0)

	return flows[start].Month + "-01"
}
