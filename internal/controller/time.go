package controller

import "time"

func getStartAndEndOfMonth(date time.Time) (time.Time, time.Time) {
	year, month, _ := date.Date()
	firstDayOfThisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	endOfThisMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)

	return firstDayOfThisMonth, endOfThisMonth
}
