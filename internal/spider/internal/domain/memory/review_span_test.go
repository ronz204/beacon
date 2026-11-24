package memory_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
)

func TestReviewSpanDueAfter(t *testing.T) {
	base := must(kernel.ParseCalendarDate("2026-11-22"))
	cases := map[string]string{
		"6mo":  "2027-05-22",
		"12mo": "2027-11-22",
		"30d":  "2026-12-22",
	}

	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			span := must(memory.ParseReviewSpan(raw))

			if got := span.DueAfter(base).String(); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}
