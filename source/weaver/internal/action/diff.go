package action

import (
	"context"
	"errors"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type DiffInput struct {
	Since string
	Base  string
	Raise bool
}

type DiffResult struct {
	Revision string
	Entries  []Entry
	Pending  []Entry
	Risk     deltas.Assessment
}

type Differ struct {
	Changes deltas.ChangeSource
	Base    BaseResolver
	Zones   ZoneSource
}

func (d Differ) Pending(ctx context.Context) ([]Entry, error) {
	result, err := d.Diff(ctx, DiffInput{})

	return result.Pending, err
}

func IsBaseMissing(err error) bool {
	var failure CheckError

	return errors.As(err, &failure) && failure.Code == BaseMissing
}

func (d Differ) Diff(
	ctx context.Context,
	input DiffInput,
) (DiffResult, error) {
	revision := input.Since
	if revision == "" {
		base, err := d.Base.Base(ctx, input.Base)
		if err != nil {
			return DiffResult{}, err
		}

		revision = base
	}

	changes, err := d.Changes.ListChanges(ctx, revision)
	if err != nil {
		return DiffResult{}, err
	}

	entries := EntriesOf(changes)

	risk, err := d.risk(ctx, revision, entries, input.Raise)
	if err != nil {
		return DiffResult{}, err
	}

	return DiffResult{
		Revision: revision,
		Entries:  entries,
		Pending:  PendingOf(entries),
		Risk:     risk,
	}, nil
}

func (d Differ) risk(
	ctx context.Context,
	revision string,
	entries []Entry,
	raised bool,
) (deltas.Assessment, error) {
	var zones []deltas.Zone

	if d.Zones != nil {
		found, err := d.Zones.Zones(ctx, revision)
		if err != nil {
			return deltas.Assessment{}, err
		}

		zones = found
	}

	return deltas.AssessRisk(findingsOf(entries), zones, raised), nil
}
