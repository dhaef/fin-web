package util

import (
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
