package spider

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
	"github.com/ronz204/beacon/source/spider/internal/driver/archive"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
	"github.com/ronz204/beacon/source/spider/internal/driver/random"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
	"github.com/ronz204/beacon/source/system"
)

const (
	stateFolder   = ".beacon"
	gitFolder     = ".git"
	archiveFolder = "archive"
	recordsLock   = "records"
)

var searchedFields = []string{
	"statement", "success", "name", "intent", "text",
	"what", "why", "request", "title",
}

type Finding = action.Finding

const (
	FindingError     = action.FindingError
	CodeNotCanonical = action.CodeNotCanonical
)

type ProjectCode string

const ProjectMissing ProjectCode = "project_missing"

type ProjectError = kernel.Error[ProjectCode]

var projectErrors = kernel.ErrorCatalog[ProjectCode]{
	ProjectMissing: "no .beacon folder found here or in a parent folder",
}

type FormatResult struct {
	Changed  []string
	Findings []Finding
}

type Project struct {
	root    string
	store   record.Store
	archive archive.Store
}

type pendingProofs struct{}

func (pendingProofs) ReadResults(
	context.Context,
) (map[ledger.PromiseID]ledger.ProofResult, error) {
	return map[ledger.PromiseID]ledger.ProofResult{}, nil
}

func Locate(start string) (Project, error) {
	directory, err := filepath.Abs(start)
	if err != nil {
		return Project{}, projectErrors.Fail(ProjectMissing, nil).WithCause(err)
	}

	for {
		state := filepath.Join(directory, stateFolder)
		if info, err := os.Stat(state); err == nil && info.IsDir() {
			return Project{
				root:    directory,
				store:   record.NewStore(state),
				archive: archive.NewStore(filepath.Join(state, archiveFolder)),
			}, nil
		}

		parent := filepath.Dir(directory)
		if parent == directory || isRepositoryRoot(directory) {
			return Project{}, projectErrors.Fail(ProjectMissing, nil)
		}

		directory = parent
	}
}

func isRepositoryRoot(directory string) bool {
	_, err := os.Stat(filepath.Join(directory, gitFolder))
	return err == nil
}

func (p Project) Format(ctx context.Context) FormatResult {
	changed, failures := p.store.Format(ctx, true)

	for index, file := range changed {
		changed[index] = p.relative(file)
	}

	return FormatResult{
		Changed:  changed,
		Findings: p.relativeFindings(action.FindingsFrom(failures)),
	}
}

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
	if err == nil {
		input.Claims = claims
		input.Workspace = repository
	}

	return p.relativeFindings(action.Doctor(ctx, input))
}

func (p Project) Brief(
	ctx context.Context,
	pending []string,
) (BriefView, []Finding) {
	sources, config, failures := p.sources()
	result, briefFailures := action.Brief(
		ctx,
		sources,
		config.Limits.Brief,
		pending,
	)

	return briefView(result), p.findings(failures, briefFailures)
}

func (p Project) Context(
	ctx context.Context,
	paths []string,
) (ContextView, []Finding) {
	sources, _, failures := p.sources()
	result, contextFailures := action.Context(ctx, sources, paths)

	return contextView(result), p.findings(failures, contextFailures)
}

func (p Project) HookContext(
	ctx context.Context,
	session string,
	paths []string,
) (HookView, []Finding, error) {
	id, err := kernel.ParseSessionID(session)
	if err != nil {
		return HookView{}, nil, err
	}

	log, err := states.Open(ctx, p.root)
	if err != nil {
		return HookView{}, nil, err
	}

	sources, config, failures := p.sources()
	result, hookFailures := action.HookContext(ctx, sources, action.HookInput{
		Session: id,
		Paths:   paths,
		Limit:   config.Limits.Hook,
		Log:     log,
	})
	view := HookView{Lines: result.Lines, Omitted: result.Omitted}

	return view, p.findings(failures, hookFailures), nil
}

func (p Project) Ready(
	ctx context.Context,
	branch string,
) (ReadyView, []Finding, error) {
	parsed, err := agenda.ParseTaskBranch(branch)
	if err != nil {
		return ReadyView{}, nil, err
	}

	claims, err := states.Open(ctx, p.root)
	if err != nil {
		return ReadyView{}, nil, err
	}

	result, failures := action.Ready(ctx, p.repositories(), claims, parsed)

	return readyView(result), p.findings(failures), nil
}

func (p Project) Analyze(ctx context.Context) (AnalysisView, []Finding) {
	sources, _, failures := p.sources()
	analysis, analysisFailures := action.Analyze(ctx, sources)

	return analysisView(analysis), p.findings(failures, analysisFailures)
}

func (p Project) Find(
	ctx context.Context,
	text string,
	archived bool,
) (FindView, []Finding) {
	var candidates []action.FindCandidate
	var failures []error

	if archived {
		var err error
		candidates, err = p.archivedCandidates(ctx)
		if err != nil {
			failures = append(failures, err)
		}
	}

	result, findFailures := action.Find(
		ctx,
		p.repositories(),
		text,
		candidates,
	)

	return findView(result), p.findings(failures, findFailures)
}

func (p Project) archivedCandidates(
	ctx context.Context,
) ([]action.FindCandidate, error) {
	entries, err := p.archive.Entries(ctx)
	if err != nil {
		return nil, err
	}

	candidates := make([]action.FindCandidate, 0, len(entries))
	for _, entry := range entries {
		var fields map[string]any
		if json.Unmarshal(entry.Record, &fields) != nil {
			continue
		}

		id, _ := fields["id"].(string)
		candidate := action.FindCandidate{
			Kind: thread.NodeKind(entry.Type),
			ID:   id,
		}
		for _, key := range searchedFields {
			if text, isText := fields[key].(string); isText {
				candidate.Texts = append(candidate.Texts, text)
			}
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (p Project) Get(ctx context.Context, id string) (StoredRecord, error) {
	parsed, err := kernel.ParseRecordID(id)
	if err != nil {
		return StoredRecord{}, err
	}

	file, err := p.store.Lookup(ctx, parsed)
	if isMissing(err) {
		return p.archived(ctx, id, err)
	}
	if err != nil {
		return StoredRecord{}, err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return StoredRecord{}, kernel.NewRecordFailure(p.relative(file), 0, err)
	}

	document, err := record.ReadDocument(file, data)
	if err != nil {
		return StoredRecord{}, err
	}

	return StoredRecord{
		Record: Record{Type: document.Kind, Data: document.JSON},
		File:   p.relative(file),
	}, nil
}

func isMissing(err error) bool {
	code := kernel.DescribeFailure(err).Code
	return err != nil && code == string(record.RecordMissing)
}

func (p Project) archived(
	ctx context.Context,
	id string,
	missing error,
) (StoredRecord, error) {
	entry, found, err := p.archive.Find(ctx, id)
	if err != nil {
		return StoredRecord{}, err
	}
	if !found {
		return StoredRecord{}, missing
	}

	return StoredRecord{
		Record:   Record{Type: entry.Type, Data: entry.Record},
		File:     p.relative(entry.File),
		Archived: true,
	}, nil
}

func (p Project) StateChanges(
	ctx context.Context,
	revision string,
) ([]StateChange, error) {
	repository := commit.NewRepository(p.root)
	changes, err := repository.StateChanges(
		ctx,
		revision,
		nil,
		match.NewMatcher(),
	)
	if err != nil {
		return nil, err
	}

	views := make([]StateChange, 0, len(changes))
	for _, change := range changes {
		view, err := p.stateChange(change)
		if err != nil {
			return nil, err
		}

		views = append(views, view)
	}

	return views, nil
}

func (p Project) stateChange(change commit.StateChange) (StateChange, error) {
	view := StateChange{
		Path:     change.Path,
		Kind:     string(change.Kind),
		Baseline: change.Baseline,
	}
	if change.Baseline {
		return view, nil
	}

	file := filepath.Join(p.root, filepath.FromSlash(change.Path))
	var err error

	if view.Before, err = recordOf(file, change.Before); err != nil {
		return StateChange{}, err
	}
	if view.After, err = recordOf(file, change.After); err != nil {
		return StateChange{}, err
	}

	return view, nil
}

func recordOf(file string, data []byte) (*Record, error) {
	if data == nil {
		return nil, nil
	}

	document, err := record.ReadDocument(file, data)
	if err != nil {
		return nil, err
	}

	return &Record{Type: document.Kind, Data: document.JSON}, nil
}

func (p Project) sources() (action.Sources, record.Config, []error) {
	config, failures := p.store.LoadConfig()
	sources := action.Sources{
		Repositories: p.repositories(),
		Proofs:       pendingProofs{},
		Matcher:      match.NewMatcher(),
		Review:       config.Review,
		Today:        today(),
	}

	if isRepositoryRoot(p.root) {
		repository := commit.NewRepository(p.root)
		sources.Changes = repository
		sources.History = repository
	}

	return sources, config, failures
}

func (p Project) repositories() action.Repositories {
	return action.Repositories{
		Goals:        p.store,
		Capabilities: p.store,
		Decisions:    p.store,
		Questions:    p.store,
		Guides:       p.store,
		Works:        p.store,
		Tasks:        p.store,
		Archive:      p.archive,
	}
}

func today() kernel.CalendarDate {
	day, _ := kernel.CalendarDateOf(time.Now())
	return day
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

func (p Project) findings(failures ...[]error) []Finding {
	var all []error
	for _, group := range failures {
		all = append(all, group...)
	}

	return p.relativeFindings(action.FindingsFrom(all))
}

func (p Project) relativeFindings(findings []Finding) []Finding {
	for index := range findings {
		if findings[index].File != "" {
			findings[index].File = p.relative(findings[index].File)
		}
	}

	return findings
}

func (p Project) relative(file string) string {
	relative, err := filepath.Rel(p.root, file)
	if err != nil {
		return filepath.ToSlash(file)
	}

	return filepath.ToSlash(relative)
}

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

	return Prepared{
		Type:      draft.Kind,
		ID:        draft.ID.String(),
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
