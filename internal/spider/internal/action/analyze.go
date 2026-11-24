package action

import (
	"context"
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
	"github.com/ronz204/beacon/internal/spider/internal/domain/thread"
)

type Sources struct {
	Repositories Repositories
	Proofs       ledger.ProofReader
	Changes      ledger.ChangeReader
	History      memory.HistoryReader
	Matcher      kernel.PathMatcher
	Review       map[string]memory.ReviewSpan
	Today        kernel.CalendarDate
}

type Analysis struct {
	Snapshot      Snapshot
	Graph         thread.Graph
	Verifications Verifications
	Stale         []StaleRecord
	Signals       []thread.Signal
}

func Analyze(ctx context.Context, sources Sources) (Analysis, []error) {
	snapshot := Load(ctx, sources.Repositories)
	graph, _ := Validate(snapshot)
	verifications, proofFailures := Verify(ctx, snapshot, sources.Proofs)
	driftFailures := markDrift(ctx, snapshot, verifications, sources)
	stale, staleFailures := staleRecords(ctx, snapshot, sources)

	signals := slices.Concat(
		Signals(snapshot, graph, verifications),
		staleSignals(stale),
	)
	slices.SortFunc(signals, thread.CompareSignals)

	analysis := Analysis{
		Snapshot:      snapshot,
		Graph:         graph,
		Verifications: verifications,
		Stale:         stale,
		Signals:       signals,
	}
	failures := slices.Concat(
		snapshot.Failures,
		proofFailures,
		driftFailures,
		staleFailures,
	)

	return analysis, failures
}

func globsTouch(
	matcher kernel.PathMatcher,
	globs []kernel.PathGlob,
	files []string,
) (bool, error) {
	for _, glob := range globs {
		for _, file := range files {
			matched, err := matcher.MatchPath(glob, file)
			if err != nil {
				return false, err
			}

			if matched {
				return true, nil
			}
		}
	}

	return false, nil
}
