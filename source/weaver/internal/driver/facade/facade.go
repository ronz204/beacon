package facade

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"slices"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

var unclassifiedKeys = []string{
	"id", "created", "author", "approved", "promises", "body",
}

type Changes struct {
	project spider.Project
}

func NewChanges(project spider.Project) Changes {
	return Changes{project: project}
}

func (c Changes) ListChanges(
	ctx context.Context,
	revision string,
) ([]deltas.StateChange, error) {
	raw, err := c.project.StateChanges(ctx, revision)
	if err != nil {
		return nil, deltas.FailChangeSource(err)
	}

	changes := make([]deltas.StateChange, 0, len(raw))
	for _, change := range raw {
		converted, err := convertChange(change)
		if err != nil {
			return nil, err
		}

		changes = append(changes, converted)
	}

	return changes, nil
}

func convertChange(change spider.StateChange) (deltas.StateChange, error) {
	before, err := convertRecord(change.Before)
	if err != nil {
		return deltas.StateChange{}, err
	}

	after, err := convertRecord(change.After)
	if err != nil {
		return deltas.StateChange{}, err
	}

	closed := make([]deltas.ClosureInput, 0, len(change.Closed))
	for _, entry := range change.Closed {
		closed = append(closed, deltas.ClosureInput{
			ID:       entry.ID,
			Approved: entry.Approved,
		})
	}

	return deltas.ParseStateChange(deltas.StateChangeInput{
		Path:     change.Path,
		Kind:     change.Kind,
		Baseline: change.Baseline,
		Runners:  change.Runners,
		Closed:   closed,
		Before:   before,
		After:    after,
	})
}

func convertRecord(record *spider.Record) (*deltas.ChangedRecordInput, error) {
	if record == nil {
		return nil, nil
	}

	values, err := decodeObject(record.Data)
	if err != nil {
		return nil, deltas.FailChangeSource(err)
	}

	input := &deltas.ChangedRecordInput{
		Kind:     record.Type,
		ID:       textOf(values["id"]),
		Fields:   fieldsOf(values),
		Approved: canonical(values["approved"]),
	}

	promises, _ := values["promises"].([]any)
	for _, raw := range promises {
		item, _ := raw.(map[string]any)
		input.Promises = append(input.Promises, convertPromise(item))
	}

	return input, nil
}

func convertPromise(item map[string]any) deltas.ChangedPromiseInput {
	check, _ := item["check"].(map[string]any)

	return deltas.ChangedPromiseInput{
		ID:       textOf(item["id"]),
		Text:     textOf(item["text"]),
		Check:    canonical(item["check"]),
		Runner:   textOf(check["runner"]),
		Status:   textOf(item["status"]),
		Approved: canonical(item["approved"]),
	}
}

func fieldsOf(values map[string]any) map[string]string {
	fields := map[string]string{}

	for key, value := range maps.All(values) {
		if !slices.Contains(unclassifiedKeys, key) {
			fields[key] = canonical(value)
		}
	}

	return fields
}

func decodeObject(data []byte) (map[string]any, error) {
	var values map[string]any

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&values); err != nil {
		return nil, err
	}

	return values, nil
}

func textOf(value any) string {
	text, _ := value.(string)
	return text
}

func canonical(value any) string {
	if value == nil {
		return ""
	}
	if text, isText := value.(string); isText {
		return text
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(encoded)
}
