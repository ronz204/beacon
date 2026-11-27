package spider

import (
	"context"
	"os"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/driver/archive"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
	"github.com/ronz204/beacon/source/system"
)

type ClosePrepared struct {
	Type     string
	ID       string
	File     string
	Before   *Record
	After    *Record
	Archived Record
	file     string
	before   []byte
	record   []byte
	closing  record.Closing
	draft    *record.Draft
}

func (p Project) PrepareClose(
	ctx context.Context,
	id string,
	reason string,
) (ClosePrepared, error) {
	parsed, err := kernel.ParseRecordID(id)
	if err != nil {
		return ClosePrepared{}, err
	}

	text, err := kernel.ParseRecordText(reason)
	if err != nil {
		return ClosePrepared{}, err
	}

	snapshot := action.Load(ctx, p.repositories())
	closing := record.Closing{Closed: today(), Reason: text}

	if owner, found := action.PromiseOwner(snapshot, id); found {
		closing.Capability = owner
		return p.preparePromiseClose(ctx, id, closing)
	}

	file, err := p.store.Lookup(ctx, parsed)
	if err != nil {
		return ClosePrepared{}, err
	}

	prepared, err := p.closedRecord(file, id, closing)
	if err != nil {
		return ClosePrepared{}, err
	}

	switch prepared.Type {
	case "guide":
		return ClosePrepared{}, saveErrors.Fail(SaveGuideNotArchivable, nil)
	case "task":
		if open := action.OpenSubtasks(snapshot, id); len(open) > 0 {
			details := map[string]any{"subtasks": open}
			failure := saveErrors.Fail(SaveTaskHasOpenChildren, details)
			return ClosePrepared{}, failure
		}
	}

	return prepared, nil
}

func (p Project) closedRecord(
	file string,
	id string,
	closing record.Closing,
) (ClosePrepared, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return ClosePrepared{}, kernel.NewRecordFailure(file, 0, err)
	}

	document, err := record.ReadDocument(file, data)
	if err != nil {
		return ClosePrepared{}, err
	}

	line, err := record.ClosedJSON(document.JSON, closing)
	if err != nil {
		return ClosePrepared{}, err
	}

	return ClosePrepared{
		Type:     document.Kind,
		ID:       id,
		File:     p.relative(file),
		Before:   &Record{Type: document.Kind, Data: document.JSON},
		Archived: Record{Type: document.Kind, Data: line},
		file:     file,
		before:   data,
		record:   document.JSON,
		closing:  closing,
	}, nil
}

func (p Project) preparePromiseClose(
	ctx context.Context,
	id string,
	closing record.Closing,
) (ClosePrepared, error) {
	owner, err := kernel.ParseRecordID(closing.Capability)
	if err != nil {
		return ClosePrepared{}, err
	}

	file, err := p.store.Lookup(ctx, owner)
	if err != nil {
		return ClosePrepared{}, err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return ClosePrepared{}, kernel.NewRecordFailure(file, 0, err)
	}

	promise, _ := record.PromiseJSON(file, data, id)
	line, err := record.ClosedJSON(promise, closing)
	if err != nil {
		return ClosePrepared{}, err
	}

	draft, failures := p.store.DraftWithoutPromise(file, id)
	if len(failures) > 0 {
		return ClosePrepared{}, p.draftError(failures)
	}

	before, err := recordOf(file, data)
	if err != nil {
		return ClosePrepared{}, err
	}

	after, err := recordOf(file, draft.Output)
	if err != nil {
		return ClosePrepared{}, err
	}

	return ClosePrepared{
		Type:     "promise",
		ID:       id,
		File:     p.relative(file),
		Before:   before,
		After:    after,
		Archived: Record{Type: "promise", Data: line},
		file:     file,
		before:   data,
		record:   promise,
		closing:  closing,
		draft:    &draft,
	}, nil
}

func (p Project) Close(
	ctx context.Context,
	prepared ClosePrepared,
	stamps map[string]Approval,
) error {
	closing := prepared.closing

	for id, input := range stamps {
		if id != prepared.ID {
			return record.StampTargetFailure(id)
		}

		approval, err := kernel.ParseApproval(input)
		if err != nil {
			return err
		}
		closing.Approval = &approval
	}

	line, err := record.ClosedJSON(prepared.record, closing)
	if err != nil {
		return err
	}

	locks, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	segment := archive.SegmentName(closing.Closed)
	lock := locks.LockFile(recordsLock)

	return system.WithFileLock(lock, func() error {
		if err := record.Unchanged(prepared.file, prepared.before); err != nil {
			return err
		}

		err := p.archive.Append(segment, prepared.Type, line)
		if err != nil {
			return err
		}

		if prepared.draft != nil {
			return prepared.draft.Commit()
		}

		return record.Remove(prepared.file, prepared.before)
	})
}
