package schwab

import (
	"strings"
	"time"
)

type CustomDate struct {
	time.Time
}

func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	// Remove quotes from the JSON string
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		return nil
	}
	// Parse using your specific MM/DD/YYYY layout
	t, err := time.Parse("01/02/2006", s)
	if err != nil {
		return err
	}
	cd.Time = t
	return nil
}
