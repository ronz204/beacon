package states

import (
	"context"
	"path/filepath"
	"time"
)

const (
	activityFile    = "activity.json"
	activityWindow  = 24 * time.Hour
	activityRefresh = 10 * time.Minute
)

type activityDocument struct {
	Version   int                      `json:"version"`
	Worktrees map[string]activityEntry `json:"worktrees"`
}

type activityEntry struct {
	Agent   string `json:"agent"`
	Session string `json:"session"`
	At      string `json:"at"`
}

func (s Store) NoteAgent(
	_ context.Context,
	worktree, agent, session string,
	now time.Time,
) error {
	var current activityDocument
	if err := s.read(activityFile, &current); err != nil {
		return err
	}
	if isFresh(current.Worktrees[worktree], agent, session, now) {
		return nil
	}

	path := filepath.Join(s.directory, activityFile)

	return s.locked(path, func() error {
		var document activityDocument
		if err := s.read(activityFile, &document); err != nil {
			return err
		}
		if document.Worktrees == nil {
			document.Worktrees = make(map[string]activityEntry)
		}

		oldest := now.AddDate(0, 0, -sessionRetention)
		for name, entry := range document.Worktrees {
			if when, ok := parseMoment(entry.At); !ok || when.Before(oldest) {
				delete(document.Worktrees, name)
			}
		}

		document.Worktrees[worktree] = activityEntry{
			Agent:   agent,
			Session: session,
			At:      now.UTC().Format(time.RFC3339),
		}
		document.Version = formatVersion

		return s.write(activityFile, document)
	})
}

func (s Store) ActiveAgent(
	_ context.Context,
	worktree string,
	now time.Time,
) (string, error) {
	var document activityDocument
	if err := s.read(activityFile, &document); err != nil {
		return "", err
	}

	entry := document.Worktrees[worktree]

	when, ok := parseMoment(entry.At)
	if !ok || now.Sub(when) > activityWindow {
		return "", nil
	}

	return entry.Agent, nil
}

func isFresh(entry activityEntry, agent, session string, now time.Time) bool {
	when, ok := parseMoment(entry.At)
	if !ok || entry.Agent != agent || entry.Session != session {
		return false
	}

	return now.Sub(when) < activityRefresh
}

func parseMoment(text string) (time.Time, bool) {
	when, err := time.Parse(time.RFC3339, text)

	return when, err == nil
}
