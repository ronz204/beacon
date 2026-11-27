package action

import (
	"context"
	"errors"
	"iter"
	"maps"
	"runtime"
	"slices"
	"sync"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
)

type driftGroup struct {
	paths    []kernel.PathGlob
	promises []ledger.PromiseID
}

func markDrift(
	ctx context.Context,
	snapshot Snapshot,
	verifications Verifications,
	sources Sources,
) []error {
	groups := driftGroups(snapshot, verifications)
	changes := changedFiles(ctx, sources.Changes, maps.Keys(groups))
	var failures []error

	for _, commit := range slices.Sorted(maps.Keys(groups)) {
		files, err := changes[commit].files, changes[commit].err
		if err != nil && !errors.Is(err, ledger.ErrCommitUnknown) {
			failures = append(failures, err)
		}

		for _, group := range groups[commit] {
			current, groupErr := groupFootprint(ctx, sources, group, files)
			if groupErr != nil {
				failures = append(failures, groupErr)
			}

			for _, id := range group.promises {
				stale := verifications[id].Footprint() != current
				if err != nil || groupErr != nil || stale {
					verifications[id] = verifications[id].WithDrift()
				}
			}
		}
	}

	return failures
}

func driftGroups(
	snapshot Snapshot,
	verifications Verifications,
) map[string][]driftGroup {
	groups := make(map[string][]driftGroup)

	for _, capability := range snapshot.Capabilities {
		paths := capability.Paths()
		if len(paths) == 0 {
			continue
		}

		byCommit := make(map[string][]ledger.PromiseID)
		for _, promise := range capability.Promises() {
			commit := verifications[promise.ID()].Commit()
			if commit != "" {
				byCommit[commit] = append(byCommit[commit], promise.ID())
			}
		}

		for commit, promises := range byCommit {
			groups[commit] = append(groups[commit], driftGroup{
				paths:    paths,
				promises: promises,
			})
		}
	}

	return groups
}

type changeResult struct {
	files []string
	err   error
}

func changedFiles(
	ctx context.Context,
	reader ledger.ChangeReader,
	commits iter.Seq[string],
) map[string]changeResult {
	results := make(map[string]changeResult)
	var mutex sync.Mutex
	var group sync.WaitGroup

	slots := make(chan struct{}, runtime.NumCPU())

	for commit := range commits {
		slots <- struct{}{}

		group.Go(func() {
			defer func() { <-slots }()

			files, err := reader.FilesChangedSince(ctx, commit)

			mutex.Lock()
			defer mutex.Unlock()
			results[commit] = changeResult{files: files, err: err}
		})
	}
	group.Wait()

	return results
}

func groupFootprint(
	ctx context.Context,
	sources Sources,
	group driftGroup,
	files []string,
) (string, error) {
	matched, err := globsMatched(sources.Matcher, group.paths, files)
	if err != nil {
		return "", err
	}

	return sources.Changes.FootprintOf(ctx, matched)
}

func globsMatched(
	matcher kernel.PathMatcher,
	globs []kernel.PathGlob,
	files []string,
) ([]string, error) {
	var matched []string

	for _, file := range files {
		touched, err := globsTouch(matcher, globs, []string{file})
		if err != nil {
			return nil, err
		}

		if touched {
			matched = append(matched, file)
		}
	}

	return matched, nil
}
