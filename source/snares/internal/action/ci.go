package action

import (
	"context"

	"github.com/ronz204/beacon/source/weaver"
)

const (
	ScopeAll     = "all"
	ScopeTouched = "touched"
)

type CIInput struct {
	Base  string
	Scope string
}

type CIResult struct {
	Revision string
	Scope    string
	Diff     weaver.DiffView
	Check    weaver.CheckView
	Blocks   []Block
	Warnings []string
	Comment  string
}

type CI struct {
	Ports CIPorts
}

func (c CI) Run(ctx context.Context, input CIInput) (CIResult, error) {
	result := CIResult{Scope: input.Scope}

	if block := invalidState(c.Ports.Doctor(ctx)); block != nil {
		result.Blocks = append(result.Blocks, *block)
		result.Comment = Comment(result)

		return result, nil
	}

	diff, _, err := c.Ports.Diff(ctx, "", input.Base)
	if err != nil {
		return result, err
	}

	result.Diff = diff
	result.Revision = diff.Revision

	other, pending := splitRunners(diff.Pending)
	runnersAt := ""
	if len(other) > 0 {
		runnersAt = diff.Revision
		result.Blocks = append(result.Blocks, Block{
			Kind:    BlockRunners,
			Message: "the runners changed without a recorded yes",
			Hint:    runnersHint,
			Items:   itemsOf(other),
		})
	}
	if block := unapproved(pending); block != nil {
		result.Blocks = append(result.Blocks, *block)
	}

	check, _, err := c.Ports.CheckAt(
		ctx,
		input.Scope == ScopeAll,
		diff.Revision,
		runnersAt,
	)
	if err != nil {
		return result, err
	}

	result.Check = check
	result.Warnings = append(result.Warnings, driftWarnings(check.Promises)...)
	result.Warnings = append(
		result.Warnings,
		unavailableWarnings(check.Invocations)...,
	)
	if block := failedChecks(check.Failed); block != nil {
		result.Blocks = append(result.Blocks, *block)
	}
	result.Comment = Comment(result)

	return result, nil
}

func splitRunners(
	pending []weaver.EntryView,
) ([]weaver.EntryView, []weaver.EntryView) {
	var runners, rest []weaver.EntryView

	for _, entry := range pending {
		if entry.Reason == runnersChanged {
			runners = append(runners, entry)
			continue
		}

		rest = append(rest, entry)
	}

	return runners, rest
}
