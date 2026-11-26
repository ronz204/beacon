package action_test

import (
	"context"
	"testing"

	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
	"github.com/ronz204/beacon/source/weaver/internal/driver/double"
)

func capabilityWith(
	text, approved string,
) *deltas.ChangedRecordInput {
	return &deltas.ChangedRecordInput{
		Kind:   "capability",
		ID:     "bc-7f3a",
		Fields: map[string]string{"name": "Sign-in"},
		Promises: []deltas.ChangedPromiseInput{{
			ID:       "bc-c3d4",
			Text:     text,
			Check:    `{"runner":"test-go","target":"auth"}`,
			Runner:   "test-go",
			Status:   "active",
			Approved: approved,
		}},
	}
}

func changeOf(
	before, after *deltas.ChangedRecordInput,
) deltas.StateChange {
	change, err := deltas.ParseStateChange(deltas.StateChangeInput{
		Path:   ".beacon/records/capabilities/bc-7f3a-sign-in.yaml",
		Kind:   "modified",
		Before: before,
		After:  after,
	})
	if err != nil {
		panic(err)
	}

	return change
}

func diffOf(t *testing.T, changes ...deltas.StateChange) action.DiffResult {
	t.Helper()

	result, err := action.Differ{
		Changes: double.Changes{Items: changes},
		Base:    fakeBase{},
	}.Diff(context.Background(), action.DiffInput{})
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func TestDiffListsAHandEditedPromiseAsPending(t *testing.T) {
	edited := changeOf(
		capabilityWith("Locks after 5", "old"),
		capabilityWith("Locks after 10", "old"),
	)

	result := diffOf(t, edited)

	if result.Revision != "abc1234" || len(result.Pending) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if result.Pending[0].Finding.Target.Value() != "bc-c3d4" {
		t.Fatalf("pending = %+v", result.Pending)
	}
}

func TestDiffStopsAskingOnceTheApprovedStampChanged(t *testing.T) {
	approved := changeOf(
		capabilityWith("Locks after 5", "old"),
		capabilityWith("Locks after 10", "new"),
	)

	result := diffOf(t, approved)

	if len(result.Pending) != 0 {
		t.Fatalf("pending = %+v", result.Pending)
	}
	if len(result.Entries) != 1 || !result.Entries[0].Approved {
		t.Fatalf("entries = %+v", result.Entries)
	}
}

func TestDiffUsesTheSinceRevisionInsteadOfTheBase(t *testing.T) {
	result, err := action.Differ{
		Changes: double.Changes{},
		Base:    fakeBase{},
	}.Diff(context.Background(), action.DiffInput{Since: "def5678"})

	if err != nil {
		t.Fatal(err)
	}
	if result.Revision != "def5678" {
		t.Fatalf("revision = %q", result.Revision)
	}
}

func TestDiffOrdersByLevelFromTheHighest(t *testing.T) {
	typo := changeOf(
		capabilityWith("Locks after 5", "old"),
		capabilityWith("Locks after 5!", "old"),
	)
	renamed := capabilityWith("Locks after 5", "old")
	renamed.Fields["name"] = "Sign in"
	weakened := changeOf(
		capabilityWith("Locks after 5", "old"),
		capabilityWith("Locks after 10", "old"),
	)

	rename := changeOf(capabilityWith("x", "o"), renamed)

	result := diffOf(t, typo, rename, weakened)

	if result.Entries[0].Finding.Level != deltas.LevelWeakening {
		t.Fatalf("entries = %+v", result.Entries)
	}
}
