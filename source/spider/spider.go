package spider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

const (
	stateFolder = ".beacon"
	gitFolder   = ".git"
)

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
	root  string
	store record.Store
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
				root:  directory,
				store: record.NewStore(state),
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

func (p Project) Find(ctx context.Context, text string) (FindView, []Finding) {
	result, failures := action.Find(ctx, p.repositories(), text)

	return findView(result), p.findings(failures)
}

func (p Project) Get(ctx context.Context, id string) (StoredRecord, error) {
	parsed, err := kernel.ParseRecordID(id)
	if err != nil {
		return StoredRecord{}, err
	}

	file, err := p.store.Lookup(ctx, parsed)
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
