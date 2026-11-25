package kernel_test

import (
	"errors"
	"io"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
)

func TestRecordFailureLocatesTheFile(t *testing.T) {
	err := kernel.NewRecordFailure("records/sample.yaml", 4, io.EOF)

	if err.Code != kernel.RecordFileInvalid {
		t.Fatalf("code = %q", err.Code)
	}
	if err.Details["file"] != "records/sample.yaml" {
		t.Fatalf("file = %v", err.Details["file"])
	}
	if err.Details["line"] != 4 {
		t.Fatalf("line = %v", err.Details["line"])
	}
	if !errors.Is(err, io.EOF) {
		t.Fatal("cause is not reachable")
	}
}

func TestRecordIDIssueFailureKeepsCause(t *testing.T) {
	err := kernel.FailRecordIDIssue(io.EOF)

	requireKernelCode(t, err, kernel.KernelIDIssueFailed)
	if !errors.Is(err, io.EOF) {
		t.Fatal("cause is not reachable")
	}
}

func TestRecordFailureExposesKernelCause(t *testing.T) {
	_, cause := kernel.ParseCalendarDate("2026-02-30")

	err := kernel.NewRecordFailure("sample.yaml", 2, cause)

	requireKernelCode(t, err, kernel.KernelDateInvalid)
}
