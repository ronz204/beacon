package deltas_test

import (
	"testing"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

func promise(id, text, status string) deltas.ChangedPromiseInput {
	return deltas.ChangedPromiseInput{
		ID:     id,
		Text:   text,
		Check:  `{"runner":"test-go","target":"a"}`,
		Runner: "test-go",
		Status: status,
	}
}

func caps(
	promises ...deltas.ChangedPromiseInput,
) *deltas.ChangedRecordInput {
	return &deltas.ChangedRecordInput{
		Kind:     "capability",
		ID:       "bc-7f3a",
		Fields:   map[string]string{"name": "Login", "paths": `["src/**"]`},
		Promises: promises,
	}
}

func recordOf(
	kind, id string,
	fields map[string]string,
) *deltas.ChangedRecordInput {
	return &deltas.ChangedRecordInput{Kind: kind, ID: id, Fields: fields}
}

func changeOf(
	kind string,
	before, after *deltas.ChangedRecordInput,
) deltas.StateChange {
	input := deltas.StateChangeInput{
		Path:   ".beacon/records/sample.yaml",
		Kind:   kind,
		Before: before,
		After:  after,
	}

	parsed, err := deltas.ParseStateChange(input)
	if err != nil {
		panic(err)
	}

	return parsed
}

func levelOf(t *testing.T, change deltas.StateChange) (int, bool) {
	t.Helper()

	level, needsYes := deltas.ClassifyChange(change)

	return int(level), needsYes
}
