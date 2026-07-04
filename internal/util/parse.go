package util

import (
	"regexp"
	"strconv"
	"strings"
)

func ParseAmount(amount string) (float64, error) {
	r := strings.NewReplacer("$", "", ",", "")
	cleanInput := r.Replace(amount)

	val, err := strconv.ParseFloat(cleanInput, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}

// merchantPrefixes are payment-processor / transaction-type prefixes that carry
// no merchant identity and should be stripped before grouping.
var merchantPrefixes = []string{
	"SQ *", "SQ*", "TST* ", "TST*", "PP*", "PAYPAL *", "PAYPAL*",
	"POS ", "CHECKCARD ", "PURCHASE ", "DEBIT CARD PURCHASE ", "RECURRING ",
}

var (
	// A domain-like token, e.g. DIGITALOCEAN.COM or AWS.AMAZON.CO. Everything
	// after the merchant's domain is location noise.
	merchantDomain = regexp.MustCompile(`^[A-Z0-9&'-]+\.[A-Z]{2,}`)
	// A space-delimited run of 4+ digits: a phone number, reference, or date.
	// The merchant name ends where such a run begins. Anchored on a preceding
	// space so embedded digits (e.g. "1800FLOWERS") aren't cut.
	merchantLongNum = regexp.MustCompile(`\s\d{4,}`)
	// Leftover store/reference numbers: 2+ digit runs, optionally "#"-prefixed.
	merchantNumbers = regexp.MustCompile(`#?\b\d{2,}\b`)
	merchantSpace   = regexp.MustCompile(`\s+`)
)

// merchantLocation are trailing tokens that denote a place, not the merchant:
// US state USPS codes plus a few country/territory codes and "CITY" seen in the
// data. Stripped only when trailing, so "PEPCO" keeps all its tokens.
var merchantLocation = map[string]bool{
	"AL": true, "AK": true, "AZ": true, "AR": true, "CA": true, "CO": true,
	"CT": true, "DE": true, "FL": true, "GA": true, "HI": true, "ID": true,
	"IL": true, "IN": true, "IA": true, "KS": true, "KY": true, "LA": true,
	"ME": true, "MD": true, "MA": true, "MI": true, "MN": true, "MS": true,
	"MO": true, "MT": true, "NE": true, "NV": true, "NH": true, "NJ": true,
	"NM": true, "NY": true, "NC": true, "ND": true, "OH": true, "OK": true,
	"OR": true, "PA": true, "RI": true, "SC": true, "SD": true, "TN": true,
	"TX": true, "UT": true, "VT": true, "VA": true, "WA": true, "WV": true,
	"WI": true, "WY": true, "DC": true, "USA": true, "MEX": true, "CITY": true,
}

// NormalizeMerchant reduces a raw transaction name to a stable merchant key so
// charges from the same vendor group together despite reference numbers,
// city/state suffixes, and processor prefixes. It optimizes for a *consistent*
// key, not a pretty label: every "DIGITALOCEAN.COM <city> <state>" variant
// collapses to "DIGITALOCEAN.COM", which is what matters for recurrence
// grouping.
func NormalizeMerchant(name string) string {
	s := strings.ToUpper(strings.TrimSpace(name))

	for _, prefix := range merchantPrefixes {
		if strings.HasPrefix(s, strings.ToUpper(prefix)) {
			s = s[len(prefix):]
			break
		}
	}

	// Cut everything from the first '*' processor-reference marker.
	if i := strings.IndexByte(s, '*'); i >= 0 {
		s = s[:i]
	}

	// If a domain token is present, the merchant ends there — drop trailing
	// location words.
	tokens := strings.Fields(s)
	for i, tok := range tokens {
		if merchantDomain.MatchString(tok) {
			s = strings.Join(tokens[:i+1], " ")
			break
		}
	}

	// Otherwise the merchant ends at the first phone/reference/date number.
	if loc := merchantLongNum.FindStringIndex(s); loc != nil {
		s = s[:loc[0]]
	}

	// Strip trailing location tokens (state / country / "CITY").
	tokens = strings.Fields(s)
	for len(tokens) > 1 && merchantLocation[tokens[len(tokens)-1]] {
		tokens = tokens[:len(tokens)-1]
	}
	s = strings.Join(tokens, " ")

	// Remove any leftover embedded numbers and tidy whitespace.
	s = merchantNumbers.ReplaceAllString(s, "")
	s = merchantSpace.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}
