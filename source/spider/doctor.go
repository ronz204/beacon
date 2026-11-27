package spider

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
	"github.com/ronz204/beacon/source/system"
)

func (p Project) Doctor(ctx context.Context) []Finding {
	config, configFailures := p.store.LoadConfig()
	repository := commit.NewRepository(p.root)
	claims, err := states.Open(ctx, p.root)

	input := action.DoctorInput{
		Repositories:   p.repositories(),
		Formatter:      p.store,
		GuideLimit:     config.Limits.Guide,
		ConfigFailures: configFailures,
	}
	var repairs []Finding
	if err == nil {
		input.Claims = claims
		input.Workspace = repository
		repairs = p.completeCloses(ctx, claims.LockFile(recordsLock))
	}

	findings := append(action.Doctor(ctx, input), repairs...)

	return p.relativeFindings(refineLineEndings(findings))
}

func (p Project) completeCloses(ctx context.Context, lock string) []Finding {
	entries, err := p.archive.Entries(ctx)
	if err != nil {
		return action.FindingsFrom([]error{err})
	}

	closed := make([]action.ClosedRecord, 0, len(entries))
	for _, entry := range entries {
		closed = append(closed, action.ClosedRecord{
			Kind: thread.NodeKind(entry.Type),
			ID:   entry.ID,
		})
	}

	repositories := p.repositories()
	repositories.Archive = nil
	snapshot := action.Load(ctx, repositories)

	interrupted := action.InterruptedCloses(snapshot, closed)
	if len(interrupted) == 0 {
		return nil
	}

	err = system.WithFileLock(lock, func() error {
		for _, record := range interrupted {
			if err := p.finishClose(ctx, snapshot, record); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return action.FindingsFrom([]error{err})
	}

	return action.CompletedCloseFindings(interrupted)
}

func (p Project) finishClose(
	ctx context.Context,
	snapshot action.Snapshot,
	closed action.ClosedRecord,
) error {
	if closed.Kind != thread.NodePromise {
		id, err := kernel.ParseRecordID(closed.ID)
		if err != nil {
			return err
		}

		file, err := p.store.Lookup(ctx, id)
		if err != nil {
			return err
		}

		return system.RemoveFile(file)
	}

	owner, _ := action.PromiseOwner(snapshot, closed.ID)
	ownerID, err := kernel.ParseRecordID(owner)
	if err != nil {
		return err
	}

	file, err := p.store.Lookup(ctx, ownerID)
	if err != nil {
		return err
	}

	draft, failures := p.store.DraftWithoutPromise(file, closed.ID)
	if len(failures) > 0 {
		return failures[0]
	}

	return draft.Commit()
}
