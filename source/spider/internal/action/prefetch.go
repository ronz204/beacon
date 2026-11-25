package action

import (
	"context"
	"sync"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
)

type changeQuery func() ([]string, error)

type changeCache struct {
	reader  ledger.ChangeReader
	mutex   sync.Mutex
	queries map[string]changeQuery
}

type prefetchedHistory struct {
	memory.HistoryReader
	guides func() (map[memory.GuideID]kernel.CalendarDate, error)
}

func newChangeCache(reader ledger.ChangeReader) *changeCache {
	return &changeCache{
		reader:  reader,
		queries: make(map[string]changeQuery),
	}
}

func (c *changeCache) prefetch(ctx context.Context, commits []string) {
	for _, commit := range commits {
		go c.query(ctx, commit)()
	}
}

func (c *changeCache) FilesChangedSince(
	ctx context.Context,
	commit string,
) ([]string, error) {
	return c.query(ctx, commit)()
}

func (c *changeCache) query(ctx context.Context, commit string) changeQuery {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if query, found := c.queries[commit]; found {
		return query
	}

	query := sync.OnceValues(func() ([]string, error) {
		return c.reader.FilesChangedSince(ctx, commit)
	})
	c.queries[commit] = query

	return query
}

func prefetchHistory(
	ctx context.Context,
	reader memory.HistoryReader,
) *prefetchedHistory {
	history := &prefetchedHistory{HistoryReader: reader}
	history.guides = sync.OnceValues(func() (
		map[memory.GuideID]kernel.CalendarDate,
		error,
	) {
		return reader.GuideChangeDays(ctx)
	})

	go history.guides()

	return history
}

func (h *prefetchedHistory) GuideChangeDays(
	context.Context,
) (map[memory.GuideID]kernel.CalendarDate, error) {
	return h.guides()
}

func resultCommits(results map[ledger.PromiseID]ledger.ProofResult) []string {
	seen := make(map[string]bool)
	var commits []string

	for _, result := range results {
		if result.Outcome == ledger.ProofPassed && !seen[result.Commit] {
			seen[result.Commit] = true
			commits = append(commits, result.Commit)
		}
	}

	return commits
}
