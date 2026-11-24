package action

import (
	"context"
	"maps"
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
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
	var failures []error

	for _, commit := range slices.Sorted(maps.Keys(groups)) {
		files, err := sources.Changes.FilesChangedSince(ctx, commit)
		if err != nil {
			failures = append(failures, err)
		}

		for _, group := range groups[commit] {
			drifted := err != nil
			if !drifted {
				var matchErr error
				drifted, matchErr = globsTouch(
					sources.Matcher,
					group.paths,
					files,
				)
				if matchErr != nil {
					failures = append(failures, matchErr)
				}
			}

			if drifted {
				for _, id := range group.promises {
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
