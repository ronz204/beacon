package kernel_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
)

func TestDescribeFailureUsesTheInnermostDomainError(t *testing.T) {
	_, cause := kernel.ParseCalendarDate("2026-02-30")
	failure := kernel.NewRecordFailure("records/sample.yaml", 4, cause)

	view := kernel.DescribeFailure(failure)

	if view.Code != string(kernel.KernelDateInvalid) {
		t.Fatalf("code = %q", view.Code)
	}
	if view.File != "records/sample.yaml" || view.Line != 4 {
		t.Fatalf("location = %s:%d", view.File, view.Line)
	}
	if view.Details["value"] != "2026-02-30" {
		t.Fatalf("details = %v", view.Details)
	}
	if _, repeated := view.Details["file"]; repeated {
		t.Fatalf("details repeat the location: %v", view.Details)
	}
}

func TestDescribeFailureKeepsTechnicalCause(t *testing.T) {
	failure := kernel.NewRecordFailure("sample.yaml", 0, io.ErrUnexpectedEOF)

	view := kernel.DescribeFailure(failure)

	if view.Code != string(kernel.RecordFileInvalid) {
		t.Fatalf("code = %q", view.Code)
	}
	if view.Cause != io.ErrUnexpectedEOF.Error() {
		t.Fatalf("cause = %q", view.Cause)
	}
}

func TestDescribeFailureLooksThroughWrapping(t *testing.T) {
	_, cause := kernel.ParseActor("jane")
	wrapped := fmt.Errorf("load: %w", cause)

	view := kernel.DescribeFailure(wrapped)

	if view.Code != string(kernel.KernelActorInvalid) {
		t.Fatalf("code = %q", view.Code)
	}
}

func TestDescribeFailureHandlesPlainErrors(t *testing.T) {
	view := kernel.DescribeFailure(io.EOF)

	if view.Code != "" {
		t.Fatalf("code = %q", view.Code)
	}
	if view.Cause != io.EOF.Error() {
		t.Fatalf("cause = %q", view.Cause)
	}
}
