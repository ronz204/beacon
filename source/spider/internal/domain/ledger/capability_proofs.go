package ledger

import "context"

type ProofOutcome string

const (
	ProofPassed ProofOutcome = "passed"
	ProofFailed ProofOutcome = "failed"
)

type ProofResult struct {
	Outcome ProofOutcome
	Commit  string
}

type ProofReader interface {
	ReadResults(ctx context.Context) (map[PromiseID]ProofResult, error)
}
