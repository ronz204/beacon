package action_test

import (
	"context"
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

func findIDs(t *testing.T, text string) []string {
	t.Helper()

	result, failures := action.Find(
		context.Background(),
		repositories(validRoot),
		text,
		nil,
	)
	requireClean(t, failures)

	ids := make([]string, 0, len(result.Matches))
	for _, match := range result.Matches {
		ids = append(ids, string(match.Kind)+":"+match.ID)
	}

	return ids
}

func TestFindIgnoresCaseAndAccents(t *testing.T) {
	ids := findIDs(t, "SÍGN-IN")

	requireLines(t, ids, []string{
		string(thread.NodeCapability) + ":bc-7f3a",
		string(thread.NodePromise) + ":bc-c4a1",
	})
}

func TestFindWithBlankTextMatchesNothing(t *testing.T) {
	if ids := findIDs(t, "  "); len(ids) != 0 {
		t.Fatalf("ids = %v", ids)
	}
}
