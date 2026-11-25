package action_test

import (
	"context"
	"slices"
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/action"
)

func TestCloseRulesFindOwnersAndOpenSubtasks(t *testing.T) {
	snapshot := action.Load(context.Background(), repositories(validRoot))

	owner, found := action.PromiseOwner(snapshot, "bc-c4a1")
	if !found || owner != "bc-7f3a" {
		t.Fatalf("owner = %q, %v", owner, found)
	}

	open := action.OpenSubtasks(snapshot, "bc-91cd.1")
	if !slices.Equal(open, []string{"bc-91cd.1.1"}) {
		t.Fatalf("open = %v", open)
	}
}
