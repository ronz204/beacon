package spider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/action"
	"github.com/ronz204/beacon/internal/spider/internal/driver/record"
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

	findings := action.Doctor(ctx, action.DoctorInput{
		Repositories: action.Repositories{
			Goals:        p.store,
			Capabilities: p.store,
			Decisions:    p.store,
			Questions:    p.store,
			Guides:       p.store,
			Works:        p.store,
			Tasks:        p.store,
		},
		Formatter:      p.store,
		GuideLimit:     config.Limits.Guide,
		ConfigFailures: configFailures,
	})

	return p.relativeFindings(findings)
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
