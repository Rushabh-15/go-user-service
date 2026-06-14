package service

import (
	"testing"
	"time"
)

func TestCalculateAge(t *testing.T) {
	// Helper to parse a YYYY-MM-DD string into a time.Time for the table.
	date := func(s string) time.Time {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatalf("invalid test date %q: %v", s, err)
		}
		return d
	}

	tests := []struct {
		name string
		dob  string
		now  string
		want int
	}{
		{"birthday already passed this year", "1990-01-10", "2026-06-14", 36},
		{"birthday not yet this year", "1990-12-31", "2026-06-14", 35},
		{"birthday is today", "2000-06-14", "2026-06-14", 26},
		{"day before birthday", "2000-06-15", "2026-06-14", 25},
		{"born today, age zero", "2026-06-14", "2026-06-14", 0},
		{"leap-day dob, before Mar 1 in non-leap year", "2000-02-29", "2025-02-28", 24},
		{"leap-day dob, on Mar 1 in non-leap year", "2000-02-29", "2025-03-01", 25},
		{"born Mar 1, one year later (leap-shift guard)", "2000-03-01", "2001-03-01", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateAge(date(tt.dob), date(tt.now))
			if got != tt.want {
				t.Errorf("CalculateAge(dob=%s, now=%s) = %d; want %d",
					tt.dob, tt.now, got, tt.want)
			}
		})
	}
}
