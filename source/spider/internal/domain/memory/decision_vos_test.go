package memory_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
)

var sampleNow = time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func requireCode[C ~string](t *testing.T, err error, want C) {
	t.Helper()

	var got kernel.Error[C]
	if !errors.As(err, &got) {
		t.Fatalf("expected domain error %q, got %v", want, err)
	}
	if got.Code != want {
		t.Fatalf("code = %q, want %q", got.Code, want)
	}
}

func sampleText(raw string) kernel.RecordText {
	return must(kernel.ParseRecordText(raw))
}

func TestParseDecisionID(t *testing.T) {
	got, err := memory.ParseDecisionID("bc-a1b2")

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "bc-a1b2" {
		t.Fatalf("got %q", got)
	}

	_, err = memory.ParseDecisionID("bc-a1b2.1")
	requireCode(t, err, memory.DecisionIDInvalid)
}

func TestParseDecisionStatus(t *testing.T) {
	for _, status := range memory.DecisionStatuses() {
		got, err := memory.ParseDecisionStatus(string(status))

		if err != nil {
			t.Fatal(err)
		}
		if got != status {
			t.Fatalf("got %q", got)
		}
	}

	_, err := memory.ParseDecisionStatus("draft")
	requireCode(t, err, memory.DecisionStatusInvalid)
}

func TestParseReviewSpanAccepts(t *testing.T) {
	cases := map[string]string{
		"6mo":   "6mo",
		"12mo":  "12mo",
		"30d":   "30d",
		" 6mo ": "6mo",
	}

	for raw, want := range cases {
		t.Run(raw, func(t *testing.T) {
			got, err := memory.ParseReviewSpan(raw)

			if err != nil {
				t.Fatal(err)
			}
			if got.String() != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestParseReviewSpanRejectsInvalid(t *testing.T) {
	invalid := []string{
		"",
		"0mo",
		"06mo",
		"6m",
		"6",
		"-6mo",
		"6MO",
		"1y",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := memory.ParseReviewSpan(raw)

			requireCode(t, err, memory.ReviewSpanInvalid)
		})
	}
}
