package double

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
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

func (c Changes) FootprintOf(
	_ context.Context,
	files []string,
) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	ordered := slices.Clone(files)
	slices.Sort(ordered)

	return strings.Join(ordered, ","), nil
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

type Claims struct {
	items []agenda.TaskClaim
}

type Workspace struct {
	Branches  []agenda.TaskBranch
	Worktrees []string
}

type Injections map[string]map[string]bool

func NewClaims(items ...agenda.TaskClaim) *Claims {
	return &Claims{items: slices.Clone(items)}
}

func (c *Claims) ListClaims(context.Context) ([]agenda.TaskClaim, error) {
	return slices.Clone(c.items), nil
}

func (c *Claims) SaveClaim(_ context.Context, claim agenda.TaskClaim) error {
	c.items = slices.DeleteFunc(c.items, func(item agenda.TaskClaim) bool {
		return item.Task == claim.Task
	})
	c.items = append(c.items, claim)

	return nil
}

func (c *Claims) RemoveClaims(
	_ context.Context,
	tasks []agenda.TaskID,
) error {
	c.items = slices.DeleteFunc(c.items, func(item agenda.TaskClaim) bool {
		return slices.Contains(tasks, item.Task)
	})

	return nil
}

func (w Workspace) ListBranches(context.Context) ([]agenda.TaskBranch, error) {
	return slices.Clone(w.Branches), nil
}

func (w Workspace) WorktreeExists(
	_ context.Context,
	path string,
) (bool, error) {
	return slices.Contains(w.Worktrees, path), nil
}

func (i Injections) InjectedKeys(
	_ context.Context,
	session kernel.SessionID,
) (map[string]bool, error) {
	return maps.Clone(i[session.String()]), nil
}

func (i Injections) RecordInjected(
	_ context.Context,
	session kernel.SessionID,
	keys []string,
) error {
	if i[session.String()] == nil {
		i[session.String()] = make(map[string]bool)
	}

	for _, key := range keys {
		i[session.String()][key] = true
	}

	return nil
}
