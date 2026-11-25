package action

import (
	"context"
	"slices"
	"sync"

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
	var cache *changeCache
	if sources.Changes != nil {
		cache = newChangeCache(sources.Changes)
		sources.Changes = cache
	}
	if sources.History != nil {
		sources.History = prefetchHistory(ctx, sources.History)
	}

	var results map[ledger.PromiseID]ledger.ProofResult
	var proofFailures []error
	var reading sync.WaitGroup
	reading.Go(func() {
		var err error
		results, err = sources.Proofs.ReadResults(ctx)
		if err != nil {
			proofFailures = append(proofFailures, err)
		}
		if cache != nil {
			cache.prefetch(ctx, resultCommits(results))
		}
	})

	snapshot := Load(ctx, sources.Repositories)
	graph, _ := Validate(snapshot)
	reading.Wait()

	verifications := verify(snapshot, results)
	var driftFailures, staleFailures []error
	var stale []StaleRecord
	var checking sync.WaitGroup
	if cache != nil {
		checking.Go(func() {
			driftFailures = markDrift(ctx, snapshot, verifications, sources)
		})
	}
	if sources.History != nil {
		checking.Go(func() {
			stale, staleFailures = staleRecords(ctx, snapshot, sources)
		})
	}
	checking.Wait()

	signals := slices.Concat(
		promiseSignalsOf(snapshot, graph, verifications),
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
