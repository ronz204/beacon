package action

import (
	"fmt"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
)

func invalidState(findings []spider.Finding) *Block {
	var items []string

	for _, finding := range findings {
		if finding.Severity == spider.FindingError {
			items = append(items, fmt.Sprintf(
				"%s %s: %s",
				finding.File,
				finding.Code,
				finding.Message,
			))
		}
	}

	if len(items) == 0 {
		return nil
	}

	return &Block{
		Kind:    BlockInvalidState,
		Message: "the project records are not valid",
		Hint:    invalidStateHint,
		Items:   items,
	}
}

func unapproved(pending []weaver.EntryView) *Block {
	if len(pending) == 0 {
		return nil
	}

	items := itemsOf(pending)

	return &Block{
		Kind:    BlockUnapproved,
		Message: "level 3 changes have no recorded yes",
		Hint:    unapprovedHint,
		Items:   items,
	}
}

func failedChecks(failed []weaver.PromiseResultView) *Block {
	if len(failed) == 0 {
		return nil
	}

	items := make([]string, 0, len(failed))
	for _, promise := range failed {
		items = append(items, fmt.Sprintf(
			"%s (%s)",
			promise.ID,
			promise.Capability,
		))
	}

	return &Block{
		Kind:    BlockFailedCheck,
		Message: "related checks are red",
		Hint:    failedCheckHint,
		Items:   items,
	}
}

func driftWarnings(promises []weaver.PromiseResultView) []string {
	var warnings []string

	for _, promise := range promises {
		if promise.Drifted {
			warnings = append(warnings, "promise "+promise.ID+" drifted")
		}
	}

	return warnings
}

func unavailableWarnings(invocations []weaver.InvocationView) []string {
	var warnings []string

	for _, invocation := range invocations {
		if invocation.Status == "unavailable" {
			warnings = append(warnings, "runner "+invocation.Runner+
				" could not start: its promises were not checked")
		}
	}

	return warnings
}

func itemsOf(entries []weaver.EntryView) []string {
	items := make([]string, 0, len(entries))

	for _, entry := range entries {
		items = append(items, fmt.Sprintf(
			"%s %s %s %s",
			entry.Path,
			entry.Target,
			entry.ID,
			entry.Reason,
		))
	}

	return items
}
