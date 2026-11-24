package double

import (
	"context"
	"maps"

	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
)

type Proofs map[ledger.PromiseID]ledger.ProofOutcome

func (p Proofs) ReadOutcomes(
	context.Context,
) (map[ledger.PromiseID]ledger.ProofOutcome, error) {
	return maps.Clone(p), nil
}
