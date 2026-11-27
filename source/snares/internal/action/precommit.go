package action

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
)

const baseMissingCode = "base_missing"

type PreCommitResult struct {
	Blocks   []Block
	Warnings []string
	Checked  []string
}

type PreCommit struct {
	Ports Ports
}

func (p PreCommit) Run(ctx context.Context) (PreCommitResult, error) {
	result := PreCommitResult{}

	if block := invalidState(p.Ports.Doctor(ctx)); block != nil {
		result.Blocks = append(result.Blocks, *block)

		return result, nil
	}

	if !p.Ports.HasHead(ctx) {
		result.Warnings = append(result.Warnings, "no commit yet: nothing "+
			"to compare against")

		return result, nil
	}

	diff, _, err := p.Ports.Diff(ctx, "HEAD", "")
	if err != nil {
		return result, err
	}
	if block := unapproved(diff.Pending); block != nil {
		result.Blocks = append(result.Blocks, *block)
	}

	return p.check(ctx, result)
}

func (p PreCommit) check(
	ctx context.Context,
	result PreCommitResult,
) (PreCommitResult, error) {
	view, _, err := p.Ports.Check(ctx, "", false, "")
	if kernel.DescribeFailure(err).Code == baseMissingCode {
		result.Warnings = append(result.Warnings, "no base branch found: "+
			"the related checks were skipped")

		return result, nil
	}
	if err != nil {
		return result, err
	}

	result.Checked = view.Capabilities
	result.Warnings = append(result.Warnings, driftWarnings(view.Promises)...)
	result.Warnings = append(
		result.Warnings,
		unavailableWarnings(view.Invocations)...,
	)
	if block := failedChecks(view.Failed); block != nil {
		result.Blocks = append(result.Blocks, *block)
	}

	return result, nil
}
