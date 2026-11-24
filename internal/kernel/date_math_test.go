package kernel_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
)

func day(t *testing.T, raw string) kernel.CalendarDate {
	t.Helper()

	parsed, err := kernel.ParseCalendarDate(raw)
	if err != nil {
		t.Fatal(err)
	}

	return parsed
}

func TestCalendarDatePlus(t *testing.T) {
	cases := []struct {
		start  string
		months int
		days   int
		want   string
	}{
		{"2026-11-22", 6, 0, "2027-05-22"},
		{"2026-11-22", 0, 30, "2026-12-22"},
		{"2024-02-29", 12, 0, "2025-03-01"},
	}

	for _, c := range cases {
		t.Run(c.start, func(t *testing.T) {
			got := day(t, c.start).Plus(c.months, c.days)

			if got.String() != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestCalendarDateBefore(t *testing.T) {
	earlier := day(t, "2026-09-30")
	later := day(t, "2026-10-01")

	if !earlier.Before(later) || later.Before(earlier) {
		t.Fatal("Before does not follow calendar order")
	}
	if earlier.Before(earlier) {
		t.Fatal("a day is not before itself")
	}
}
