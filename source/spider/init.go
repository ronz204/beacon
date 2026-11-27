package spider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ronz204/beacon/source/system"
)

const (
	configName   = "config.yaml"
	configText   = "schema: 1\n"
	maxCapacity  = 10
	goalQuestion = "What is the goal of this project, who is it for and " +
		"what will it not do?"
	folderMode = 0o755
)

type InitRecord struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	File string `json:"file,omitempty"`
	Text string `json:"text"`
}

type InitView struct {
	DryRun   bool         `json:"dryRun"`
	Config   string       `json:"config"`
	Scan     ScanView     `json:"scan"`
	Records  []InitRecord `json:"records"`
	Omitted  int          `json:"omitted"`
	Proposal string       `json:"proposal,omitempty"`
}

type Author func(project Project) (string, error)

func Init(
	ctx context.Context,
	directory string,
	dryRun bool,
	author Author,
) (InitView, error) {
	root, err := filepath.Abs(directory)
	if err != nil {
		return InitView{}, err
	}

	if _, err := Locate(root); err == nil {
		return InitView{}, projectErrors.Fail(ProjectExists, nil)
	}

	found, err := Scan(root)
	if err != nil {
		return InitView{}, err
	}

	planned, omitted := plannedRecords(found)
	view := InitView{
		DryRun:   dryRun,
		Config:   path.Join(stateFolder, configName),
		Scan:     found,
		Omitted:  omitted,
		Proposal: proposalOf(found.Runners),
		Records:  make([]InitRecord, 0, len(planned)),
	}

	if dryRun {
		for _, item := range planned {
			view.Records = append(view.Records, item.view)
		}

		return view, nil
	}

	records, err := create(ctx, root, planned, author)
	view.Records = records

	return view, err
}

type plannedRecord struct {
	view    InitRecord
	payload []byte
}

func plannedRecords(found ScanView) ([]plannedRecord, int) {
	planned := []plannedRecord{questionOf(goalQuestion)}

	shown := found.Candidates
	omitted := 0
	if len(shown) > maxCapacity {
		omitted = len(shown) - maxCapacity
		shown = shown[:maxCapacity]
	}

	for _, candidate := range shown {
		planned = append(planned, capabilityOf(candidate.Path))
		planned = append(planned, questionOf(fmt.Sprintf(
			"What should the code in %s promise? Name the observable "+
				"results worth checking.",
			candidate.Path,
		)))
	}
	if omitted > 0 {
		planned = append(planned, questionOf(restQuestion(found, omitted)))
	}

	return planned, omitted
}

func restQuestion(found ScanView, omitted int) string {
	names := make([]string, 0, omitted)
	for _, candidate := range found.Candidates[maxCapacity:] {
		names = append(names, candidate.Path)
	}

	return fmt.Sprintf(
		"%d more folders could be capabilities (%s). Which of them "+
			"deserve one?",
		omitted,
		strings.Join(names, ", "),
	)
}

func questionOf(text string) plannedRecord {
	return planOf("question", map[string]any{
		"text":   text,
		"status": "open",
	}, text)
}

func capabilityOf(folder string) plannedRecord {
	name := nameOf(folder)

	return planOf("capability", map[string]any{
		"name": name,
		"intent": fmt.Sprintf(
			"Code in %s; its purpose is an open question.",
			folder,
		),
		"paths":    []string{folder + "/**"},
		"promises": []any{},
	}, name)
}

func nameOf(folder string) string {
	base := strings.NewReplacer("-", " ", "_", " ").Replace(path.Base(folder))

	return strings.ToUpper(base[:1]) + base[1:]
}

func planOf(kind string, record map[string]any, text string) plannedRecord {
	payload, _ := json.Marshal(map[string]any{
		"type":   kind,
		"record": record,
	})

	return plannedRecord{
		view:    InitRecord{Type: kind, Text: text},
		payload: payload,
	}
}

func proposalOf(runners []ScanRunner) string {
	if len(runners) == 0 {
		return ""
	}

	section := map[string]any{}
	for _, runner := range runners {
		section[runner.Name] = map[string]any{"run": runner.Run}
	}

	payload, _ := json.Marshal(map[string]any{
		"type":   "config",
		"record": map[string]any{"schema": 1, "runners": section},
	})

	return string(payload)
}

func create(
	ctx context.Context,
	root string,
	planned []plannedRecord,
	author Author,
) ([]InitRecord, error) {
	state := filepath.Join(root, stateFolder)
	if err := os.MkdirAll(state, folderMode); err != nil {
		return nil, err
	}

	records, err := seed(ctx, root, planned, author)
	if err != nil {
		_ = os.RemoveAll(state)
	}

	return records, err
}

func seed(
	ctx context.Context,
	root string,
	planned []plannedRecord,
	author Author,
) ([]InitRecord, error) {
	config := filepath.Join(root, stateFolder, configName)
	if err := system.WriteFileAtomic(config, []byte(configText)); err != nil {
		return nil, err
	}

	project, err := Locate(root)
	if err != nil {
		return nil, err
	}

	actor, err := author(project)
	if err != nil {
		return nil, err
	}

	records := make([]InitRecord, 0, len(planned))
	for _, item := range planned {
		prepared, err := project.Prepare(ctx, item.payload, actor)
		if err != nil {
			return nil, err
		}
		if err := project.Save(ctx, prepared, nil); err != nil {
			return nil, err
		}

		saved := item.view
		saved.ID = prepared.ID
		saved.File = prepared.File
		records = append(records, saved)
	}

	return records, nil
}
