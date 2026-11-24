package action

import (
	"context"
	"maps"
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
)

type driftCandidate struct {
	promise ledger.PromiseID
	paths   []kernel.PathGlob
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

		for _, candidate := range groups[commit] {
			drifted := err != nil
			if !drifted {
				var matchErr error
				drifted, matchErr = globsTouch(
					sources.Matcher,
					candidate.paths,
					files,
				)
				if matchErr != nil {
					failures = append(failures, matchErr)
				}
			}

			if drifted {
				id := candidate.promise
				verifications[id] = verifications[id].WithDrift()
			}
		}
	}

	return failures
}

func driftGroups(
	snapshot Snapshot,
	verifications Verifications,
) map[string][]driftCandidate {
	groups := make(map[string][]driftCandidate)

	for _, capability := range snapshot.Capabilities {
		paths := capability.Paths()
		if len(paths) == 0 {
			continue
		}

		for _, promise := range capability.Promises() {
			commit := verifications[promise.ID()].Commit()
			if commit == "" {
				continue
			}

			groups[commit] = append(groups[commit], driftCandidate{
				promise: promise.ID(),
				paths:   paths,
			})
		}
	}

	return groups
}
