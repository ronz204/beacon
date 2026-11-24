package kernel_test

import (
	"testing"
	"time"

	"github.com/ronz204/beacon/internal/kernel"
)

func TestIsMissingDetectsZeroValue(t *testing.T) {
	text, err := kernel.ParseRecordText("present")
	if err != nil {
		t.Fatal(err)
	}

	if !kernel.IsMissing(kernel.RecordText{}) {
		t.Fatal("zero value should be missing")
	}
	if kernel.IsMissing(text) {
		t.Fatal("parsed value should be present")
	}
}

func TestParseRootRecordID(t *testing.T) {
	got, err := kernel.ParseRootRecordID("bc-91cd")

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "bc-91cd" {
		t.Fatalf("got %q", got)
	}

	invalid := []string{
		"",
		"bc-91cd.1",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := kernel.ParseRootRecordID(raw)

			requireKernelCode(t, err, kernel.KernelRecordIDInvalid)
		})
	}
}

func TestCalendarDateOfUsesTheMomentZone(t *testing.T) {
	zone := time.FixedZone("sample", -6*60*60)
	moment := time.Date(2026, 11, 22, 23, 0, 0, 0, zone)

	got, err := kernel.CalendarDateOf(moment)

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "2026-11-22" {
		t.Fatalf("got %q", got)
	}
}
