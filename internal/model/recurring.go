package model

import (
	"database/sql"

	"fin-web/internal/recurring"
)

// RecurringCandidates returns every expense charge eligible for recurring
// detection: positive amounts (the app stores expenses positive), excluding
// reimbursements and ignored categories. Uncategorized rows are kept, since
// subscriptions are frequently uncategorized.
func RecurringCandidates(conn *sql.DB) ([]recurring.Charge, error) {
	rows, err := conn.Query(`
		SELECT t.name, t.amount, t.date
		FROM transactions AS t
		LEFT JOIN categories AS c ON t.category_id = c.id
		WHERE t.amount > 0
		  AND COALESCE(t.is_reimbursement, 0) = 0
		  AND COALESCE(c.is_ignored, 0) = 0
		  AND t.date IS NOT NULL AND t.date != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	charges := []recurring.Charge{}
	for rows.Next() {
		var c recurring.Charge
		if err := rows.Scan(&c.Name, &c.Amount, &c.Date); err != nil {
			return nil, err
		}
		charges = append(charges, c)
	}

	return charges, rows.Err()
}
