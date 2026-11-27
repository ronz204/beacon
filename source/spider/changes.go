package spider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

func (p Project) StateChanges(
	ctx context.Context,
	revision string,
) ([]StateChange, error) {
	config, failures := p.store.LoadConfig()
	if len(failures) > 0 {
		return nil, failures[0]
	}

	repository := commit.NewRepository(p.root)
	changes, err := repository.StateChanges(
		ctx,
		revision,
		baselinesOf(config),
		match.NewMatcher(),
	)
	if err != nil {
		return nil, err
	}

	views := make([]StateChange, 0, len(changes))
	for _, change := range changes {
		view, err := p.stateChange(change, config)
		if err != nil {
			return nil, err
		}

		views = append(views, view)
	}

	return views, nil
}

func (p Project) stateChange(
	change commit.StateChange,
	config record.Config,
) (StateChange, error) {
	view := StateChange{
		Path:     change.Path,
		Kind:     string(change.Kind),
		Baseline: change.Baseline,
	}
	if change.Baseline {
		runners, err := baselineRunners(config, change.Path)
		view.Runners = runners

		return view, err
	}
	if change.Archive {
		var err error
		view.Closed, err = closuresOf(change)

		return view, err
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

func baselineRunners(config record.Config, path string) ([]string, error) {
	matcher := match.NewMatcher()
	var runners []string

	for _, name := range slices.Sorted(maps.Keys(config.Runners)) {
		for _, glob := range config.Runners[name].Baseline {
			matched, err := matcher.MatchPath(glob, path)
			if err != nil {
				return nil, err
			}
			if matched {
				runners = append(runners, name)
				break
			}
		}
	}

	return runners, nil
}

func baselinesOf(config record.Config) []kernel.PathGlob {
	var globs []kernel.PathGlob

	for _, name := range slices.Sorted(maps.Keys(config.Runners)) {
		globs = append(globs, config.Runners[name].Baseline...)
	}

	return globs
}

type closedLine struct {
	Type   string `json:"type"`
	Record struct {
		ID       string          `json:"id"`
		Approved json.RawMessage `json:"approved"`
		Promises []struct {
			ID string `json:"id"`
		} `json:"promises"`
	} `json:"record"`
}

func closuresOf(change commit.StateChange) ([]Closure, error) {
	known := archiveLines(change.Before)
	var closures []Closure

	for _, text := range archiveLines(change.After) {
		if slices.Contains(known, text) {
			continue
		}

		var line closedLine
		if err := json.Unmarshal([]byte(text), &line); err != nil {
			return nil, kernel.NewRecordFailure(change.Path, 0, err)
		}

		closure := Closure{
			Type:     line.Type,
			ID:       line.Record.ID,
			Approved: string(line.Record.Approved),
		}
		for _, promise := range line.Record.Promises {
			closure.Promises = append(closure.Promises, promise.ID)
		}

		closures = append(closures, closure)
	}

	return closures, nil
}

func archiveLines(data []byte) []string {
	var lines []string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(nil, 1<<24)

	for scanner.Scan() {
		if text := strings.TrimSpace(scanner.Text()); text != "" {
			lines = append(lines, text)
		}
	}

	return lines
}
