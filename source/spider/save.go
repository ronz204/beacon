package spider

import (
	"context"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/driver/random"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

type SaveCode string

const (
	SavePayloadInvalid      SaveCode = "payload_invalid"
	SaveGraphInvalid        SaveCode = "graph_invalid"
	SaveGuideNotArchivable  SaveCode = "guide_not_archivable"
	SaveTaskHasOpenChildren SaveCode = "task_has_open_children"
)

type SaveError = kernel.Error[SaveCode]

var saveErrors = kernel.ErrorCatalog[SaveCode]{
	SavePayloadInvalid:      "the payload does not describe a valid record",
	SaveGraphInvalid:        "the change breaks references or relations",
	SaveGuideNotArchivable:  "guides are deleted, not archived",
	SaveTaskHasOpenChildren: "finish or close the open subtasks first",
}

var passedThrough = []string{
	string(record.RecordMissing),
	string(record.RecordPayloadInvalid),
	string(record.RecordPromiseOmitted),
	string(record.RecordPromiseUnknown),
}

type Approval = kernel.ApprovalInput

type Prepared struct {
	Type      string
	ID        string
	File      string
	Before    *Record
	After     Record
	Ignored   []string
	Unchanged bool
	draft     record.Draft
}

func (p Project) Prepare(
	ctx context.Context,
	payload []byte,
	author string,
) (Prepared, error) {
	actor, err := kernel.ParseActor(author)
	if err != nil {
		return Prepared{}, err
	}

	snapshot := action.Load(ctx, p.repositories())
	if snapshot.Archived == nil {
		nodes, err := p.archive.ArchivedNodes(ctx)
		if err != nil {
			return Prepared{}, err
		}
		snapshot.Archived = nodes
	}

	ids := action.RecordIDs(snapshot)
	draft, failures := p.store.Draft(record.DraftInput{
		Payload: payload,
		Author:  actor,
		Today:   today(),
		Issuer:  random.NewIssuer(),
		Taken: func(id kernel.RecordID) bool {
			return ids[id.String()]
		},
	})
	if len(failures) > 0 {
		return Prepared{}, p.draftError(failures)
	}

	introduced := action.IntroducedFailures(snapshot, draft.Entity)
	if len(introduced) > 0 {
		return Prepared{}, p.saveError(SaveGraphInvalid, introduced)
	}

	return p.prepared(draft)
}

func (p Project) prepared(draft record.Draft) (Prepared, error) {
	after, err := recordOf(draft.File, draft.Output)
	if err != nil {
		return Prepared{}, err
	}

	before, err := recordOf(draft.File, draft.Before)
	if err != nil {
		return Prepared{}, err
	}

	id := draft.ID.String()
	if draft.Kind == record.ConfigKind {
		id = record.ConfigKind
	}

	return Prepared{
		Type:      draft.Kind,
		ID:        id,
		File:      p.relative(draft.File),
		Before:    before,
		After:     *after,
		Ignored:   draft.Ignored,
		Unchanged: draft.Unchanged(),
		draft:     draft,
	}, nil
}

func (p Project) Save(
	ctx context.Context,
	prepared Prepared,
	stamps map[string]Approval,
) error {
	approvals := make(map[string]kernel.Approval, len(stamps))
	for id, input := range stamps {
		approval, err := kernel.ParseApproval(input)
		if err != nil {
			return err
		}
		approvals[id] = approval
	}

	draft, failures := prepared.draft.Stamp(approvals)
	if len(failures) > 0 {
		return p.draftError(failures)
	}

	locks, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	return p.store.Write(draft, locks.LockFile(recordsLock))
}

func (p Project) draftError(failures []error) error {
	if len(failures) == 1 {
		code := kernel.DescribeFailure(failures[0]).Code
		if slices.Contains(passedThrough, code) {
			return failures[0]
		}
	}

	return p.saveError(SavePayloadInvalid, failures)
}

func (p Project) saveError(code SaveCode, failures []error) error {
	return saveErrors.Fail(code, map[string]any{
		"findings": p.findings(failures),
	})
}
