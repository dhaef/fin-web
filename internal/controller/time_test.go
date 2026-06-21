package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetStartAndEndOfMonth(t *testing.T) {
	tests := []struct {
		name      string
		date      time.Time
		wantStart string
		wantEnd   string
	}{
		{
			name:      "mid month",
			date:      time.Date(2026, 2, 15, 13, 30, 0, 0, time.Local),
			wantStart: "2026-02-01",
			wantEnd:   "2026-02-28",
		},
		{
			name:      "leap year february",
			date:      time.Date(2024, 2, 10, 0, 0, 0, 0, time.Local),
			wantStart: "2024-02-01",
			wantEnd:   "2024-02-29",
		},
		{
			name:      "december rolls to year end",
			date:      time.Date(2025, 12, 25, 0, 0, 0, 0, time.Local),
			wantStart: "2025-12-01",
			wantEnd:   "2025-12-31",
		},
		{
			name:      "first of month",
			date:      time.Date(2026, 6, 1, 23, 59, 59, 0, time.Local),
			wantStart: "2026-06-01",
			wantEnd:   "2026-06-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := getStartAndEndOfMonth(tt.date)
			assert.Equal(t, tt.wantStart, start.Format("2006-01-02"))
			assert.Equal(t, tt.wantEnd, end.Format("2006-01-02"))
			// Start should be midnight local.
			assert.Equal(t, 0, start.Hour())
			assert.Equal(t, 0, start.Minute())
			assert.Equal(t, 0, start.Second())
		})
	}
}
