package controller

import (
	"net/http"
	"time"

	"fin-web/internal/model"
	"fin-web/internal/recurring"
)

type SubscriptionsPage struct {
	Report           recurring.Report
	SubCount         int
	BillCount        int
	MonthlySubTotal  float64
	AnnualSubTotal   float64
	MonthlyBillTotal float64
}

func (c *Controller) subscriptions(w http.ResponseWriter, r *http.Request) error {
	charges, err := model.RecurringCandidates(c.db)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching recurring candidates: " + err.Error(),
		}
	}

	report := recurring.Detect(charges, time.Now())

	page := SubscriptionsPage{
		Report:           report,
		SubCount:         len(report.Subscriptions),
		BillCount:        len(report.Bills),
		MonthlySubTotal:  report.MonthlySubTotal,
		AnnualSubTotal:   report.MonthlySubTotal * 12,
		MonthlyBillTotal: report.MonthlyBillTotal,
	}

	if err := renderTemplate(w, Base[SubscriptionsPage]{Data: page}, "layout", []string{"subscriptions.html", "layout.html"}); err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return nil
}
