package spider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/driver/archive"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

const (
	stateFolder   = ".beacon"
	gitFolder     = ".git"
	archiveFolder = "archive"
	recordsLock   = "records"
)

type Finding = action.Finding

const (
	FindingError     = action.FindingError
	CodeNotCanonical = action.CodeNotCanonical
	CodeLineEndings  = action.CodeLineEndings
)

type ProjectCode string

const ProjectMissing ProjectCode = "project_missing"

type ProjectError = kernel.Error[ProjectCode]

var projectErrors = kernel.ErrorCatalog[ProjectCode]{
	ProjectMissing: "no .beacon folder found here or in a parent folder",
}

type Project struct {
	root    string
	store   record.Store
	archive archive.Store
}

type noProofs struct{}

func (noProofs) ReadResults(
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

func (p Project) sources(
	ctx context.Context,
) (action.Sources, record.Config, []error) {
	config, failures := p.store.LoadConfig()
	sources := action.Sources{
		Repositories: p.repositories(),
		Proofs:       noProofs{},
		Matcher:      match.NewMatcher(),
		Review:       config.Review,
		Today:        today(),
	}

	if isRepositoryRoot(p.root) {
		repository := commit.NewRepository(p.root)
		sources.Changes = repository
		sources.History = repository

		local, err := states.Open(ctx, p.root)
		if err != nil {
			return sources, config, append(failures, err)
		}
		sources.Proofs = local
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
