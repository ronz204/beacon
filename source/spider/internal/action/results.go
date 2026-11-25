package action

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
)

type Verdict struct {
	Promise ledger.PromiseID
	Outcome ledger.ProofOutcome
	Reason  ledger.ProofReason
}

type ProofInput struct {
	Verdicts []Verdict
	Changes  ledger.ChangeReader
	Matcher  kernel.PathMatcher
	Head     string
}

func ProofsOf(
	ctx context.Context,
	snapshot Snapshot,
	input ProofInput,
) (map[ledger.PromiseID]ledger.ProofResult, []string, error) {
	files, err := input.Changes.FilesChangedSince(ctx, input.Head)
	if err != nil {
		return nil, nil, err
	}

	owners := promiseOwners(snapshot)
	footprints := make(map[string]string)
	results := make(map[ledger.PromiseID]ledger.ProofResult)
	var unknown []string

	for _, verdict := range input.Verdicts {
		capability, found := owners[verdict.Promise.String()]
		if !found {
			unknown = append(unknown, verdict.Promise.String())
			continue
		}

		key := capability.ID().String()
		footprint, cached := footprints[key]
		if !cached {
			paths := capability.Paths()
			matched, err := globsMatched(input.Matcher, paths, files)
			if err != nil {
				return nil, nil, err
			}

			footprint, err = input.Changes.FootprintOf(ctx, matched)
			if err != nil {
				return nil, nil, err
			}
			footprints[key] = footprint
		}

		results[verdict.Promise] = ledger.ProofResult{
			Outcome:   verdict.Outcome,
			Reason:    verdict.Reason,
			Commit:    input.Head,
			Footprint: footprint,
		}
	}

	return results, unknown, nil
}

func promiseOwners(snapshot Snapshot) map[string]ledger.Capability {
	owners := make(map[string]ledger.Capability)

	for _, capability := range snapshot.Capabilities {
		for _, promise := range capability.Promises() {
			owners[promise.ID().String()] = capability
		}
	}

	return owners
}
