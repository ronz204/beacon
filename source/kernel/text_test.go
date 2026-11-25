package kernel_test

import (
	"testing"

	"github.com/ronz204/beacon/source/kernel"
)

func TestParseRecordTextTrims(t *testing.T) {
	got, err := kernel.ParseRecordText("  first line\nsecond line  ")

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "first line\nsecond line" {
		t.Fatalf("got %q", got)
	}
}

func TestParseRecordTextRejectsBlank(t *testing.T) {
	blank := []string{
		"",
		"   ",
		"\n\t",
	}

	for _, raw := range blank {
		t.Run(raw, func(t *testing.T) {
			_, err := kernel.ParseRecordText(raw)

			requireKernelCode(t, err, kernel.KernelTextInvalid)
		})
	}
}
