package kernel_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
)

type sampleCode string

const (
	sampleBroken  sampleCode = "sample_broken"
	sampleMessage            = "sample_broken: sample is broken"
)

var sampleErrors = kernel.ErrorCatalog[sampleCode]{
	sampleBroken: "sample is broken",
}

func requireKernelCode(
	t *testing.T,
	err error,
	want kernel.KernelCode,
) {
	t.Helper()

	var got kernel.KernelError
	if !errors.As(err, &got) {
		t.Fatalf("expected KernelError, got %v", err)
	}
	if got.Code != want {
		t.Fatalf("code = %q, want %q", got.Code, want)
	}
}

func TestErrorCarriesCodeMessageAndDetails(t *testing.T) {
	details := map[string]any{"value": "input"}

	err := sampleErrors.Fail(sampleBroken, details)

	if err.Error() != sampleMessage {
		t.Fatalf("message = %q", err.Error())
	}
	if err.Details["value"] != "input" {
		t.Fatalf("details = %v", err.Details)
	}
}

func TestErrorKeepsCauseOutOfMessage(t *testing.T) {
	err := sampleErrors.Fail(sampleBroken, nil).WithCause(io.EOF)

	if !errors.Is(err, io.EOF) {
		t.Fatal("cause is not reachable through Unwrap")
	}
	if err.Error() != sampleMessage {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestErrorIsRecoverableThroughWrapping(t *testing.T) {
	cause := sampleErrors.Fail(sampleBroken, nil)
	wrapped := fmt.Errorf("load: %w", cause)

	var got kernel.Error[sampleCode]
	if !errors.As(wrapped, &got) || got.Code != sampleBroken {
		t.Fatalf("errors.As failed for %v", wrapped)
	}
}
