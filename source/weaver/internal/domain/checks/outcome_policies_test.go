package checks_test

import (
	"testing"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
)

func TestOutcomeOfReadsTheExecutionResult(t *testing.T) {
	red := checks.ExecutionResult{ExitCode: 1}
	late := checks.ExecutionResult{ExitCode: -1, TimedOut: true}

	tests := []struct {
		name   string
		result checks.ExecutionResult
		status checks.OutcomeStatus
		reason checks.OutcomeReason
	}{
		{"green", checks.ExecutionResult{}, checks.OutcomePassed, ""},
		{"red", red, checks.OutcomeFailed, checks.ReasonExitCode},
		{"timeout", late, checks.OutcomeFailed, checks.ReasonTimeout},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outcome := checks.OutcomeOf(test.result)

			if outcome.Status() != test.status {
				t.Fatalf("status = %q", outcome.Status())
			}
			if outcome.Reason() != test.reason {
				t.Fatalf("reason = %q", outcome.Reason())
			}
		})
	}
}
