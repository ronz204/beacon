package action

import (
	"context"
	"errors"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type DiffInput struct {
	Since string
	Base  string
}

type DiffResult struct {
	Revision string
	Entries  []Entry
	Pending  []Entry
}

type Differ struct {
	Changes deltas.ChangeSource
	Base    BaseResolver
}

func (d Differ) Pending(ctx context.Context) ([]Entry, error) {
	result, err := d.Diff(ctx, DiffInput{})

	var failure CheckError
	if errors.As(err, &failure) && failure.Code == BaseMissing {
		return nil, nil
	}

	return result.Pending, err
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

	return DiffResult{
		Revision: revision,
		Entries:  entries,
		Pending:  PendingOf(entries),
	}, nil
}
