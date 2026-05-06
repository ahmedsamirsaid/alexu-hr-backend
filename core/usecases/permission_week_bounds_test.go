package usecases

import (
	"testing"
	"time"
)

func TestWeekBoundsForDate(t *testing.T) {
	loc := time.UTC

	cases := []struct {
		name      string
		weekend   []int
		date      time.Time
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "Fri-Sat weekend, query Wednesday → week starts Sunday",
			weekend:   []int{5, 6},
			date:      time.Date(2026, 5, 6, 0, 0, 0, 0, loc), // Wed
			wantStart: time.Date(2026, 5, 3, 0, 0, 0, 0, loc), // Sun
			wantEnd:   time.Date(2026, 5, 9, 0, 0, 0, 0, loc), // Sat (last calendar day of 7-day span)
		},
		{
			name:      "Fri-Sat weekend, query Sunday → week starts that Sunday",
			weekend:   []int{5, 6},
			date:      time.Date(2026, 5, 3, 0, 0, 0, 0, loc), // Sun
			wantStart: time.Date(2026, 5, 3, 0, 0, 0, 0, loc),
			wantEnd:   time.Date(2026, 5, 9, 0, 0, 0, 0, loc),
		},
		{
			name:      "Sat-only weekend, query Tuesday → week starts Sunday",
			weekend:   []int{6},
			date:      time.Date(2026, 5, 5, 0, 0, 0, 0, loc), // Tue
			wantStart: time.Date(2026, 5, 3, 0, 0, 0, 0, loc), // Sun (the day after Sat=weekend)
			wantEnd:   time.Date(2026, 5, 9, 0, 0, 0, 0, loc),
		},
		{
			name:      "Sun-only weekend, query Tuesday → week starts Monday",
			weekend:   []int{0},
			date:      time.Date(2026, 5, 5, 0, 0, 0, 0, loc), // Tue
			wantStart: time.Date(2026, 5, 4, 0, 0, 0, 0, loc), // Mon
			wantEnd:   time.Date(2026, 5, 10, 0, 0, 0, 0, loc),
		},
		{
			name:      "Three-day weekend Fri-Sat-Sun, query Tuesday → week starts Monday",
			weekend:   []int{0, 5, 6},
			date:      time.Date(2026, 5, 5, 0, 0, 0, 0, loc), // Tue
			wantStart: time.Date(2026, 5, 4, 0, 0, 0, 0, loc), // Mon
			wantEnd:   time.Date(2026, 5, 10, 0, 0, 0, 0, loc),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotStart, gotEnd := WeekBoundsForDate(tc.weekend, tc.date)
			if !gotStart.Equal(tc.wantStart) {
				t.Errorf("start: got %s want %s", gotStart, tc.wantStart)
			}
			if !gotEnd.Equal(tc.wantEnd) {
				t.Errorf("end: got %s want %s", gotEnd, tc.wantEnd)
			}
		})
	}
}
