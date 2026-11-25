package states

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/system"
)

const (
	stateFolder      = "beacon"
	claimsFile       = "claims.json"
	injectedFile     = "injected.json"
	formatVersion    = 1
	sessionRetention = 7
)

type Store struct {
	directory string
}

type claimsDocument struct {
	Version int          `json:"version"`
	Claims  []claimEntry `json:"claims"`
}

type claimEntry struct {
	Task     string `json:"task"`
	Branch   string `json:"branch"`
	Worktree string `json:"worktree"`
	Day      string `json:"day"`
}

type injectedDocument struct {
	Version  int                     `json:"version"`
	Sessions map[string]sessionEntry `json:"sessions"`
}

type sessionEntry struct {
	Day  string   `json:"day"`
	Keys []string `json:"keys"`
}

func Open(ctx context.Context, root string) (Store, error) {
	output, err := system.Git(ctx, root, "rev-parse", "--git-common-dir")
	if err != nil {
		return Store{}, kernel.FailHistory(err)
	}

	common := strings.TrimSpace(output)
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}

	return NewStore(filepath.Join(common, stateFolder))
}

func NewStore(directory string) (Store, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return Store{}, kernel.NewRecordFailure(directory, 0, err)
	}

	return Store{directory: directory}, nil
}

func (s Store) LockFile(name string) string {
	return filepath.Join(s.directory, name)
}

func (s Store) ListClaims(ctx context.Context) ([]agenda.TaskClaim, error) {
	var document claimsDocument
	if err := s.read(claimsFile, &document); err != nil {
		return nil, err
	}

	claims := make([]agenda.TaskClaim, 0, len(document.Claims))
	for _, entry := range document.Claims {
		claim, err := entry.claim()
		if err != nil {
			return nil, s.failure(claimsFile, err)
		}

		claims = append(claims, claim)
	}

	return claims, nil
}

func (s Store) SaveClaim(ctx context.Context, claim agenda.TaskClaim) error {
	return s.updateClaims(func(entries []claimEntry) []claimEntry {
		task := claim.Task.String()
		kept := slices.DeleteFunc(entries, func(entry claimEntry) bool {
			return entry.Task == task
		})

		return append(kept, claimEntry{
			Task:     task,
			Branch:   claim.Branch.String(),
			Worktree: claim.Worktree,
			Day:      claim.Day.String(),
		})
	})
}

func (s Store) RemoveClaims(ctx context.Context, tasks []agenda.TaskID) error {
	removed := make(map[string]bool, len(tasks))
	for _, task := range tasks {
		removed[task.String()] = true
	}

	return s.updateClaims(func(entries []claimEntry) []claimEntry {
		return slices.DeleteFunc(entries, func(entry claimEntry) bool {
			return removed[entry.Task]
		})
	})
}

func (s Store) InjectedKeys(
	ctx context.Context,
	session kernel.SessionID,
) (map[string]bool, error) {
	var document injectedDocument
	if err := s.read(injectedFile, &document); err != nil {
		return nil, err
	}

	keys := make(map[string]bool)
	for _, key := range document.Sessions[session.String()].Keys {
		keys[key] = true
	}

	return keys, nil
}

func (s Store) RecordInjected(
	ctx context.Context,
	session kernel.SessionID,
	keys []string,
) error {
	path := filepath.Join(s.directory, injectedFile)
	today := time.Now().Format(time.DateOnly)
	oldest := time.Now().AddDate(0, 0, -sessionRetention).Format(time.DateOnly)

	return s.locked(path, func() error {
		var document injectedDocument
		if err := s.read(injectedFile, &document); err != nil {
			return err
		}
		if document.Sessions == nil {
			document.Sessions = make(map[string]sessionEntry)
		}

		for name, entry := range document.Sessions {
			if entry.Day < oldest {
				delete(document.Sessions, name)
			}
		}

		entry := document.Sessions[session.String()]
		entry.Day = today
		entry.Keys = mergeKeys(entry.Keys, keys)
		document.Sessions[session.String()] = entry
		document.Version = formatVersion

		return s.write(injectedFile, document)
	})
}

func (s Store) updateClaims(change func([]claimEntry) []claimEntry) error {
	path := filepath.Join(s.directory, claimsFile)

	return s.locked(path, func() error {
		var document claimsDocument
		if err := s.read(claimsFile, &document); err != nil {
			return err
		}

		document.Version = formatVersion
		document.Claims = change(document.Claims)
		slices.SortFunc(document.Claims, func(a, b claimEntry) int {
			return cmp.Compare(a.Task, b.Task)
		})

		return s.write(claimsFile, document)
	})
}

func (s Store) locked(path string, work func() error) error {
	if err := system.WithFileLock(path, work); err != nil {
		return s.failure(filepath.Base(path), err)
	}

	return nil
}

func (s Store) read(name string, target any) error {
	data, err := os.ReadFile(filepath.Join(s.directory, name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return s.failure(name, err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return s.failure(name, err)
	}

	return nil
}

func (s Store) write(name string, document any) error {
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return s.failure(name, err)
	}

	path := filepath.Join(s.directory, name)
	if err := system.WriteFileAtomic(path, append(data, '\n')); err != nil {
		return s.failure(name, err)
	}

	return nil
}

func (s Store) failure(name string, err error) error {
	var located kernel.RecordFailure
	if errors.As(err, &located) {
		return err
	}

	return kernel.NewRecordFailure(filepath.Join(s.directory, name), 0, err)
}

func (e claimEntry) claim() (agenda.TaskClaim, error) {
	task, err := agenda.ParseTaskID(e.Task)
	if err != nil {
		return agenda.TaskClaim{}, err
	}

	branch, err := agenda.ParseTaskBranch(e.Branch)
	if err != nil {
		return agenda.TaskClaim{}, err
	}

	day, err := kernel.ParseCalendarDate(e.Day)
	if err != nil {
		return agenda.TaskClaim{}, err
	}

	return agenda.TaskClaim{
		Task:     task,
		Branch:   branch,
		Worktree: e.Worktree,
		Day:      day,
	}, nil
}

func mergeKeys(existing, added []string) []string {
	merged := slices.Concat(existing, added)
	slices.Sort(merged)

	return slices.Compact(merged)
}
