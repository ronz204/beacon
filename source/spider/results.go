package spider

import (
	"context"
	"time"

	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

type Verdict struct {
	Promise string
	Outcome string
	Reason  string
}

func (p Project) RecordResults(ctx context.Context, verdicts []Verdict) error {
	parsed, err := parseVerdicts(verdicts)
	if err != nil {
		return err
	}

	local, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	repository := commit.NewRepository(p.root)
	head, err := repository.Head(ctx)
	if err != nil {
		return err
	}

	snapshot := action.Load(ctx, p.repositories())
	results, unknown, err := action.ProofsOf(ctx, snapshot, action.ProofInput{
		Verdicts: parsed,
		Changes:  repository,
		Matcher:  match.NewMatcher(),
		Head:     head,
	})
	if err != nil {
		return err
	}
	if len(unknown) > 0 {
		return record.MissingFailure(unknown[0])
	}

	return local.RecordResults(ctx, results, time.Now())
}

func parseVerdicts(verdicts []Verdict) ([]action.Verdict, error) {
	parsed := make([]action.Verdict, 0, len(verdicts))

	for _, verdict := range verdicts {
		id, err := ledger.ParsePromiseID(verdict.Promise)
		if err != nil {
			return nil, err
		}

		outcome, err := ledger.ParseProofOutcome(verdict.Outcome)
		if err != nil {
			return nil, err
		}

		reason, err := ledger.ParseProofReason(verdict.Reason)
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, action.Verdict{
			Promise: id,
			Outcome: outcome,
			Reason:  reason,
		})
	}

	return parsed, nil
}
