// Command subsreport is a calibration harness for recurring/subscription
// detection. It runs the real internal/recurring detector over the live
// database and prints what it flags, plus (with SHOW_RAW=1) the raw transaction
// name variants behind each normalized merchant key — the ground truth for
// tuning normalization. Run: `set -a; . ./.env; set +a; go run ./cmd/subsreport`.
package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"fin-web/internal/db"
	"fin-web/internal/model"
	"fin-web/internal/recurring"
	"fin-web/internal/util"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("DB_PATH is required (run: set -a; . ./.env; set +a; go run ./cmd/subsreport)")
	}

	conn, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	charges, err := model.RecurringCandidates(conn)
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("SHOW_RAW") != "" {
		dumpRawVariants(charges, []string{
			"SPOTIFY", "DIGITALOCEAN", "SHELL", "UBER", "AMAZON", "AMZN",
			"NETFLIX", "HULU", "DISNEY", "FITNESS", "HOUR", "PEPCO",
		})
	}

	report := recurring.Detect(charges, time.Now())

	fmt.Printf("Scanned %d expense charges.\n", len(charges))
	fmt.Printf("Thresholds: minOccurrences=%d gapCV<=%.2f amountCV<=%.2f billMonthly>=%.0f\n\n",
		recurring.MinOccurrences, recurring.GapCVMax, recurring.AmountCVMax, recurring.BillMonthlyMin)

	printList("ACTIVE SUBSCRIPTIONS", report.Subscriptions)
	fmt.Printf("Subscriptions: %.2f/mo  (%.2f/yr)\n\n", report.MonthlySubTotal, report.MonthlySubTotal*12)

	printList("ACTIVE BILLS", report.Bills)
	fmt.Printf("Bills: %.2f/mo  (%.2f/yr)\n\n", report.MonthlyBillTotal, report.MonthlyBillTotal*12)

	printList("POSSIBLE (2 charges, low confidence)", report.Possible)
	printList("LIKELY CANCELED", report.Canceled)
}

func printList(title string, rs []recurring.Recurring) {
	fmt.Printf("=== %s (%d) ===\n", title, len(rs))
	if len(rs) == 0 {
		fmt.Printf("(none)\n\n")
		return
	}
	fmt.Printf("%-30s %4s %-9s %8s %9s %9s  %-10s %-10s %s\n",
		"MERCHANT", "N", "CADENCE", "TYPICAL", "MONTHLY", "ANNUAL", "LAST", "NEXT", "AMOUNT")
	for _, r := range rs {
		fmt.Printf("%-30s %4d %-9s %8.2f %9.2f %9.2f  %-10s %-10s %s\n",
			trunc(r.Merchant, 30), r.Count, r.Cadence, r.TypicalAmt, r.Monthly, r.Annual,
			r.Last, r.Next, fixedLabel(r.AmountFixed))
	}
	fmt.Println()
}

// dumpRawVariants prints, for every merchant key containing any watch term, the
// distinct raw transaction names that normalized into it.
func dumpRawVariants(charges []recurring.Charge, watch []string) {
	raws := map[string]map[string]int{} // key -> raw name -> count
	for _, c := range charges {
		key := util.NormalizeMerchant(c.Name)
		if key == "" {
			continue
		}
		if raws[key] == nil {
			raws[key] = map[string]int{}
		}
		raws[key][c.Name]++
	}

	var keys []string
	for key := range raws {
		for _, w := range watch {
			if strings.Contains(key, w) { // keys are upper-cased
				keys = append(keys, key)
				break
			}
		}
	}
	sort.Strings(keys)

	fmt.Printf("=== RAW NAME VARIANTS (watchlist) ===\n")
	for _, key := range keys {
		fmt.Printf("KEY %q:\n", key)
		var names []string
		for n := range raws[key] {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Printf("    %3d  %s\n", raws[key][n], n)
		}
	}
	fmt.Printf("\n")
}

func fixedLabel(fixed bool) string {
	if fixed {
		return "fixed"
	}
	return "variable"
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
