package states

import (
	"context"
	"maps"
	"path/filepath"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
)

const resultsFile = "results.json"

type resultsDocument struct {
	Version int                    `json:"version"`
	Results map[string]resultEntry `json:"results"`
}

type resultEntry struct {
	Outcome   string `json:"outcome"`
	Reason    string `json:"reason,omitempty"`
	Commit    string `json:"commit"`
	Footprint string `json:"footprint"`
	At        string `json:"at"`
}

func (s Store) ReadResults(
	context.Context,
) (map[ledger.PromiseID]ledger.ProofResult, error) {
	var document resultsDocument
	if err := s.read(resultsFile, &document); err != nil {
		return nil, err
	}

	results := make(map[ledger.PromiseID]ledger.ProofResult)
	for raw, entry := range document.Results {
		id, result, err := parseResult(raw, entry)
		if err != nil {
			return nil, s.failure(resultsFile, err)
		}

		results[id] = result
	}

	return results, nil
}

func parseResult(
	raw string,
	entry resultEntry,
) (ledger.PromiseID, ledger.ProofResult, error) {
	id, err := ledger.ParsePromiseID(raw)
	if err != nil {
		return ledger.PromiseID{}, ledger.ProofResult{}, err
	}

	outcome, err := ledger.ParseProofOutcome(entry.Outcome)
	if err != nil {
		return ledger.PromiseID{}, ledger.ProofResult{}, err
	}

	reason, err := ledger.ParseProofReason(entry.Reason)
	if err != nil {
		return ledger.PromiseID{}, ledger.ProofResult{}, err
	}

	return id, ledger.ProofResult{
		Outcome:   outcome,
		Reason:    reason,
		Commit:    entry.Commit,
		Footprint: entry.Footprint,
	}, nil
}

func (s Store) RecordResults(
	ctx context.Context,
	results map[ledger.PromiseID]ledger.ProofResult,
	at time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return kernel.FailHistory(err)
	}

	path := filepath.Join(s.directory, resultsFile)

	return s.locked(path, func() error {
		var document resultsDocument
		if err := s.read(resultsFile, &document); err != nil {
			return err
		}

		merged := make(map[string]resultEntry, len(document.Results))
		maps.Copy(merged, document.Results)
		for id, result := range results {
			merged[id.String()] = resultEntry{
				Outcome:   string(result.Outcome),
				Reason:    string(result.Reason),
				Commit:    result.Commit,
				Footprint: result.Footprint,
				At:        at.UTC().Format(time.RFC3339),
			}
		}

		return s.write(resultsFile, resultsDocument{
			Version: formatVersion,
			Results: merged,
		})
	})
}
