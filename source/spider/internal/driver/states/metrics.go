package states

import (
	"context"
	"path/filepath"
	"time"

	"github.com/ronz204/beacon/source/kernel"
)

const metricsFile = "metrics.json"

type metricsDocument struct {
	Version  int                       `json:"version"`
	Sessions map[string]metricsSession `json:"sessions"`
}

type metricsSession struct {
	Day    string                `json:"day"`
	Events map[string]eventCount `json:"events"`
}

type eventCount struct {
	Calls    int `json:"calls"`
	Injected int `json:"injected"`
	Lines    int `json:"lines"`
}

func (s Store) RecordInjection(
	_ context.Context,
	session kernel.SessionID,
	event string,
	lines int,
) error {
	path := filepath.Join(s.directory, metricsFile)
	today := time.Now().Format(time.DateOnly)
	oldest := time.Now().AddDate(0, 0, -sessionRetention).Format(time.DateOnly)

	return s.locked(path, func() error {
		var document metricsDocument
		if err := s.read(metricsFile, &document); err != nil {
			return err
		}
		if document.Sessions == nil {
			document.Sessions = make(map[string]metricsSession)
		}

		for name, entry := range document.Sessions {
			if entry.Day < oldest {
				delete(document.Sessions, name)
			}
		}

		entry := document.Sessions[session.String()]
		entry.Day = today
		entry.Events = countInjection(entry.Events, event, lines)
		document.Sessions[session.String()] = entry
		document.Version = formatVersion

		return s.write(metricsFile, document)
	})
}

func countInjection(
	events map[string]eventCount,
	event string,
	lines int,
) map[string]eventCount {
	if events == nil {
		events = make(map[string]eventCount)
	}

	count := events[event]
	count.Calls++
	count.Lines += lines

	if lines > 0 {
		count.Injected++
	}

	events[event] = count

	return events
}
