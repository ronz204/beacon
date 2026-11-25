package ledger

import (
	"context"
	"slices"
	"strings"
)

type ProofOutcome string

const (
	ProofPassed ProofOutcome = "passed"
	ProofFailed ProofOutcome = "failed"
)

type ProofReason string

const (
	ProofExitCode ProofReason = "exit_code"
	ProofTimeout  ProofReason = "timeout"
)

type ProofResult struct {
	Outcome   ProofOutcome
	Reason    ProofReason
	Commit    string
	Footprint string
}

type ProofReader interface {
	ReadResults(ctx context.Context) (map[PromiseID]ProofResult, error)
}

func ProofOutcomes() []ProofOutcome {
	return []ProofOutcome{ProofPassed, ProofFailed}
}

func ParseProofOutcome(raw string) (ProofOutcome, error) {
	outcome := ProofOutcome(strings.TrimSpace(raw))

	if !slices.Contains(ProofOutcomes(), outcome) {
		return "", invalidCapabilityValue(ProofOutcomeInvalid, raw)
	}

	return outcome, nil
}

func ProofReasons() []ProofReason {
	return []ProofReason{ProofExitCode, ProofTimeout}
}

func ParseProofReason(raw string) (ProofReason, error) {
	reason := ProofReason(strings.TrimSpace(raw))

	if reason != "" && !slices.Contains(ProofReasons(), reason) {
		return "", invalidCapabilityValue(ProofReasonInvalid, raw)
	}

	return reason, nil
}
