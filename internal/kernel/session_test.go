package kernel_test

import (
	"strings"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
)

func TestParseSessionID(t *testing.T) {
	got, err := kernel.ParseSessionID(" sample-session_1.a ")
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "sample-session_1.a" {
		t.Fatalf("got %q", got)
	}

	invalid := []string{
		"",
		"with space",
		"../escape",
		"a/b",
		strings.Repeat("a", 129),
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := kernel.ParseSessionID(raw)

			requireKernelCode(t, err, kernel.KernelSessionInvalid)
		})
	}
}
