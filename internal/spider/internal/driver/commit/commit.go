package commit

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
	"github.com/ronz204/beacon/internal/system"
)

const (
	recordsFolder = ".beacon/records/"
	guidesFolder  = ".beacon/records/guides"
	commitMark    = "\x00"
	commitFormat  = "--format=%x00%cs"
)

var (
	commitPattern    = regexp.MustCompile(`^[0-9a-f]{4,64}$`)
	guideFilePattern = regexp.MustCompile(`^(bc-[0-9a-f]{4,})-.+\.md$`)

	errInvalidCommit   = errors.New("commit must be a hexadecimal hash")
	errInvalidRevision = errors.New("revision must be a plain git reference")
)

type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeModified ChangeKind = "modified"
	ChangeDeleted  ChangeKind = "deleted"
)

type StateChange struct {
	Path     string
	Kind     ChangeKind
	Baseline bool
	Before   []byte
	After    []byte
}

type Repository struct {
	root string
}

func NewRepository(root string) Repository {
	return Repository{root: root}
}

func (r Repository) FilesChangedSince(
	ctx context.Context,
	commit string,
) ([]string, error) {
	if !commitPattern.MatchString(commit) {
		return nil, kernel.FailHistory(errInvalidCommit)
	}

	tracked, err := r.git(
		ctx,
		"diff", "--name-only", "--relative", "--no-renames", commit, "--",
	)
	if err != nil {
		return nil, err
	}

	untracked, err := r.untracked(ctx)
	if err != nil {
		return nil, err
	}

	return append(lines(tracked), untracked...), nil
}

func (r Repository) CommitsSince(
	ctx context.Context,
	day kernel.CalendarDate,
) ([]memory.CodeCommit, error) {
	output, err := r.git(
		ctx,
		"log", "--since="+day.String()+"T00:00:00",
		"--relative", "--name-only", commitFormat,
	)
	if err != nil {
		return nil, err
	}

	return parseCommits(output)
}

func (r Repository) GuideChangeDays(
	ctx context.Context,
) (map[memory.GuideID]kernel.CalendarDate, error) {
	output, err := r.git(
		ctx,
		"log", "--relative", "--name-only", commitFormat, "--", guidesFolder,
	)
	if err != nil {
		return nil, err
	}

	commits, err := parseCommits(output)
	if err != nil {
		return nil, err
	}

	days := make(map[memory.GuideID]kernel.CalendarDate)
	for _, commit := range commits {
		for _, file := range commit.Files {
			id, found := guideIDOf(file)
			if _, seen := days[id]; found && !seen {
				days[id] = commit.Day
			}
		}
	}

	return days, nil
}

func (r Repository) ListBranches(
	ctx context.Context,
) ([]agenda.TaskBranch, error) {
	output, err := r.git(
		ctx,
		"for-each-ref", "--format=%(refname:short)", "refs/heads",
	)
	if err != nil {
		return nil, err
	}

	var branches []agenda.TaskBranch
	for _, line := range lines(output) {
		if branch, err := agenda.ParseTaskBranch(line); err == nil {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

func (r Repository) WorktreeExists(
	_ context.Context,
	path string,
) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, kernel.FailHistory(err)
	}

	return info.IsDir(), nil
}

func (r Repository) StateChanges(
	ctx context.Context,
	revision string,
	baselines []kernel.PathGlob,
	matcher kernel.PathMatcher,
) ([]StateChange, error) {
	if !isRevision(revision) {
		return nil, kernel.FailHistory(errInvalidRevision)
	}

	entries, err := r.changedEntries(ctx, revision)
	if err != nil {
		return nil, err
	}

	var changes []StateChange
	for _, entry := range entries {
		baseline, err := matchesAny(matcher, baselines, entry.Path)
		if err != nil {
			return nil, err
		}

		if !baseline && !strings.HasPrefix(entry.Path, recordsFolder) {
			continue
		}

		entry.Baseline = baseline
		if err := r.fillContents(ctx, revision, &entry); err != nil {
			return nil, err
		}

		changes = append(changes, entry)
	}

	return changes, nil
}

func (r Repository) changedEntries(
	ctx context.Context,
	revision string,
) ([]StateChange, error) {
	output, err := r.git(
		ctx,
		"diff", "--name-status", "--relative", "--no-renames", revision, "--",
	)
	if err != nil {
		return nil, err
	}

	var entries []StateChange
	for _, line := range lines(output) {
		status, file, found := strings.Cut(line, "\t")
		if found {
			entries = append(entries, StateChange{
				Path: file,
				Kind: changeKind(status),
			})
		}
	}

	untracked, err := r.untracked(ctx)
	if err != nil {
		return nil, err
	}

	for _, file := range untracked {
		entries = append(entries, StateChange{Path: file, Kind: ChangeAdded})
	}

	return entries, nil
}

func (r Repository) fillContents(
	ctx context.Context,
	revision string,
	change *StateChange,
) error {
	if change.Kind != ChangeAdded {
		before, err := r.git(ctx, "show", revision+":./"+change.Path)
		if err != nil {
			return err
		}

		change.Before = []byte(before)
	}

	if change.Kind == ChangeDeleted {
		return nil
	}

	path := filepath.Join(r.root, filepath.FromSlash(change.Path))

	after, err := os.ReadFile(path)
	if err != nil {
		return kernel.FailHistory(err)
	}

	change.After = after

	return nil
}

func (r Repository) untracked(ctx context.Context) ([]string, error) {
	output, err := r.git(ctx, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	return lines(output), nil
}

func (r Repository) git(ctx context.Context, args ...string) (string, error) {
	output, err := system.Git(ctx, r.root, args...)
	if err != nil {
		return "", kernel.FailHistory(err)
	}

	return output, nil
}

func parseCommits(output string) ([]memory.CodeCommit, error) {
	var commits []memory.CodeCommit

	for _, block := range strings.Split(output, commitMark) {
		blockLines := lines(block)
		if len(blockLines) == 0 {
			continue
		}

		day, err := kernel.ParseCalendarDate(blockLines[0])
		if err != nil {
			return nil, kernel.FailHistory(err)
		}

		commits = append(commits, memory.CodeCommit{
			Day:   day,
			Files: blockLines[1:],
		})
	}

	return commits, nil
}

func lines(output string) []string {
	var found []string

	for line := range strings.Lines(output) {
		if trimmed := strings.TrimRight(line, "\r\n"); trimmed != "" {
			found = append(found, trimmed)
		}
	}

	return found
}

func guideIDOf(file string) (memory.GuideID, bool) {
	match := guideFilePattern.FindStringSubmatch(path.Base(file))
	if match == nil {
		return memory.GuideID{}, false
	}

	id, err := memory.ParseGuideID(match[1])

	return id, err == nil
}

func changeKind(status string) ChangeKind {
	switch status {
	case "A":
		return ChangeAdded
	case "D":
		return ChangeDeleted
	}

	return ChangeModified
}

func isRevision(revision string) bool {
	if revision == "" || strings.HasPrefix(revision, "-") {
		return false
	}

	return !strings.ContainsFunc(revision, isRevisionBreak)
}

func isRevisionBreak(char rune) bool {
	if char == ':' || unicode.IsSpace(char) {
		return true
	}

	return unicode.IsControl(char)
}

func matchesAny(
	matcher kernel.PathMatcher,
	globs []kernel.PathGlob,
	file string,
) (bool, error) {
	for _, glob := range globs {
		matched, err := matcher.MatchPath(glob, file)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}
