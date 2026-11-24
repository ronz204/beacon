package double

import (
	"context"
	"errors"
	"maps"
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
)

var errUnknownCommit = errors.New("unknown commit")

type Proofs map[ledger.PromiseID]ledger.ProofResult

type Changes map[string][]string

type History struct {
	Commits []memory.CodeCommit
	Guides  map[memory.GuideID]kernel.CalendarDate
}

func (p Proofs) ReadResults(
	context.Context,
) (map[ledger.PromiseID]ledger.ProofResult, error) {
	return maps.Clone(p), nil
}

func (c Changes) FilesChangedSince(
	_ context.Context,
	commit string,
) ([]string, error) {
	files, found := c[commit]
	if !found {
		return nil, kernel.FailHistory(errUnknownCommit)
	}

	return slices.Clone(files), nil
}

func (h History) CommitsSince(
	_ context.Context,
	day kernel.CalendarDate,
) ([]memory.CodeCommit, error) {
	return slices.DeleteFunc(
		slices.Clone(h.Commits),
		func(commit memory.CodeCommit) bool {
			return commit.Day.Before(day)
		},
	), nil
}

func (h History) GuideChangeDays(
	context.Context,
) (map[memory.GuideID]kernel.CalendarDate, error) {
	return maps.Clone(h.Guides), nil
}
