package kernel_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
)

func TestParseCalendarDateAccepts(t *testing.T) {
	cases := map[string]string{
		"2026-11-22":   "2026-11-22",
		"2024-02-29":   "2024-02-29",
		" 2026-11-22 ": "2026-11-22",
	}

	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			got, err := kernel.ParseCalendarDate(raw)

			if err != nil {
				t.Fatal(err)
			}
			if got.String() != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestParseCalendarDateRejectsInvalid(t *testing.T) {
	invalid := []string{
		"",
		"2026-13-01",
		"2026-02-30",
		"2025-02-29",
		"2026-1-2",
		"26-11-22",
		"2026-11-22T10:00:00Z",
		"2026/11/22",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := kernel.ParseCalendarDate(raw)

			requireKernelCode(t, err, kernel.KernelDateInvalid)
		})
	}
}
