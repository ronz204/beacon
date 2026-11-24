package ledger

import "context"

type ProofOutcome string

const (
	ProofPassed ProofOutcome = "passed"
	ProofFailed ProofOutcome = "failed"
)

type ProofReader interface {
	ReadOutcomes(ctx context.Context) (map[PromiseID]ProofOutcome, error)
}
