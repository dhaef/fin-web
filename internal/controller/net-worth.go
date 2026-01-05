package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"fin-web/internal/model"
)

func netWorth(w http.ResponseWriter, r *http.Request) error {
	netWorthItems, err := model.QueryNetWorthItems(netWorthDBConn, model.QueryNetWorthItemsFilters{
		OrderBy:        "date",
		OrderDirection: "DESC",
	})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching net worth items: " + err.Error(),
		}
	}

	for idx, item := range netWorthItems {
		netWorth := item.Cash + item.Investment + item.Debit + item.Credit + item.Savings + item.Retirement + item.Loans
		netWorthItems[idx].NetWorth = netWorth

		var changeAmt float32
		var changePercent float32
		if idx+1 != len(netWorthItems) {
			prevNetWorth := netWorthItems[idx+1].Cash + netWorthItems[idx+1].Investment + netWorthItems[idx+1].Debit + netWorthItems[idx+1].Credit + netWorthItems[idx+1].Savings + netWorthItems[idx+1].Retirement + netWorthItems[idx+1].Loans
			changeAmt = netWorth - prevNetWorth
			changePercent = (changeAmt / prevNetWorth) * 100

			netWorthItems[idx].Change = changeAmt
			netWorthItems[idx].ChangePercent = fmt.Sprintf("%.2f%%", changePercent)
		}

	}

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"netWorthItems": netWorthItems,
		},
	}, "layout", []string{"net-worth/net-worth.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func netWorthItem(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	netWorthItem, err := model.GetNetWorthItem(netWorthDBConn, id)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching net worth item: " + err.Error(),
		}
	}

	err = renderTemplate(w, Base{
		Data: map[string]any{
			"netWorthItem": netWorthItem,
		},
	}, "layout", []string{"net-worth/net-worth-form.html", "net-worth/net-worth-item.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func validateNetWorthForm(r *http.Request) (model.NetWorthItemParams, map[string]string) {
	errs := map[string]string{}
	params := model.NetWorthItemParams{}
	params.Date = ToPtr(r.FormValue("date"))

	cashStr := r.FormValue("cash")
	if cashStr != "" {
		cash, err := strconv.ParseFloat(cashStr, 32)
		if err != nil {
			errs["cash"] = "cash is not a valid float"
		}

		params.Cash = ToPtr(float32(cash))
	} else {
		errs["cash"] = "cash can't be empty"
	}

	investmentStr := r.FormValue("investment")
	if investmentStr != "" {
		investment, err := strconv.ParseFloat(investmentStr, 32)
		if err != nil {
			errs["investment"] = "investment is not a valid float"
		}

		params.Investment = ToPtr(float32(investment))
	} else {
		errs["investment"] = "investment can't be empty"
	}

	debitStr := r.FormValue("debit")
	if debitStr != "" {
		debit, err := strconv.ParseFloat(debitStr, 32)
		if err != nil {
			errs["debit"] = "debit is not a valid float"
		}

		params.Debit = ToPtr(float32(debit))
	} else {
		errs["debit"] = "debit can't be empty"
	}

	creditStr := r.FormValue("credit")
	if creditStr != "" {
		credit, err := strconv.ParseFloat(creditStr, 32)
		if err != nil {
			errs["credit"] = "credit is not a valid float"
		}

		params.Credit = ToPtr(float32(credit))
	} else {
		errs["credit"] = "credit can't be empty"
	}

	savingsStr := r.FormValue("savings")
	if savingsStr != "" {
		savings, err := strconv.ParseFloat(savingsStr, 32)
		if err != nil {
			errs["savings"] = "savings is not a valid float"
		}

		params.Savings = ToPtr(float32(savings))
	} else {
		errs["savings"] = "savings can't be empty"
	}

	retirementStr := r.FormValue("retirement")
	if retirementStr != "" {
		retirement, err := strconv.ParseFloat(retirementStr, 32)
		if err != nil {
			errs["retirement"] = "retirement is not a valid float"
		}

		params.Retirement = ToPtr(float32(retirement))
	} else {
		errs["retirement"] = "retirement can't be empty"
	}

	loansStr := r.FormValue("loans")
	if loansStr != "" {
		loans, err := strconv.ParseFloat(loansStr, 32)
		if err != nil {
			errs["loans"] = "loans is not a valid float"
		}

		params.Loans = ToPtr(float32(loans))
	} else {
		errs["loans"] = "loans can't be empty"
	}

	return params, errs
}

func ToPtr[T any](v T) *T {
	return &v
}

func updateNetWorthItem(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")

	params, errs := validateNetWorthForm(r)
	if len(errs) != 0 {
		err := renderTemplate(w, Base{
			Data: map[string]any{
				"errs":         errs,
				"netWorthItem": params,
			},
		}, "layout", []string{"net-worth/net-worth-form.html", "net-worth/net-worth-item.html", "layout.html"})
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		return nil
	}

	_, err := model.GetNetWorthItem(
		netWorthDBConn,
		id,
	)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error fetching net worth item: " + err.Error(),
		}
	}

	err = model.UpdateNetWorthItem(netWorthDBConn, id, params)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error updating net worth item: " + err.Error(),
		}
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	return nil
}

func newNetWorthItem(w http.ResponseWriter, r *http.Request) error {
	err := renderTemplate(w, Base{
		Data: map[string]any{},
	}, "layout", []string{"net-worth/net-worth-form.html", "net-worth/net-worth-item.html", "layout.html"})
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func createNetWorthItem(w http.ResponseWriter, r *http.Request) error {
	params, errs := validateNetWorthForm(r)

	if len(errs) != 0 {
		err := renderTemplate(w, Base{
			Data: map[string]any{
				"errs":         errs,
				"netWorthItem": params,
			},
		}, "layout", []string{"net-worth/net-worth-form.html", "net-worth/net-worth-item.html", "layout.html"})
		if err != nil {
			return APIError{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		return nil
	}

	ID, err := model.CreateNetWorthItem(netWorthDBConn, params)
	if err != nil {
		return APIError{
			Status:  http.StatusInternalServerError,
			Message: "error creating net worth item: " + err.Error(),
		}
	}

	http.Redirect(w, r, "/net-worth/"+ID, http.StatusSeeOther)
	return nil
}
