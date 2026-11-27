package action

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

var reviewRules = []string{
	"You review without the author's context: judge only the promises " +
		"and the changes listed here.",
	"For each promise say whether the change still keeps it, and quote " +
		"the line that shows it.",
	"Point out any level 3 change that weakens a promise without a " +
		"recorded yes.",
	"Do not edit files, commit, push or approve anything: report what " +
		"you find to the person.",
}

type VerifyInput struct {
	Base  string
	Raise bool
}

type ReviewPromise struct {
	ID         string
	Capability string
	Text       string
	State      string
	Drifted    bool
}

type Review struct {
	Promises []ReviewPromise
	Changes  []Entry
	Rules    []string
}

type VerifyResult struct {
	Revision string
	Check    CheckResult
	Risk     deltas.Assessment
	Review   *Review
}

type Verifier struct {
	Checker Checker
	Differ  Differ
	Manual  AttestPorts
}

func (v Verifier) Verify(
	ctx context.Context,
	input VerifyInput,
) (VerifyResult, error) {
	checked, err := v.Checker.Check(ctx, CheckInput{Base: input.Base})
	if err != nil {
		return VerifyResult{Check: checked}, err
	}

	diff, err := v.Differ.Diff(ctx, DiffInput{
		Base:  input.Base,
		Raise: input.Raise,
	})
	if err != nil {
		return VerifyResult{Check: checked}, err
	}

	result := VerifyResult{
		Revision: diff.Revision,
		Check:    checked,
		Risk:     diff.Risk,
	}
	if diff.Risk.Level != deltas.RiskHigh {
		return result, nil
	}

	manual, err := v.Manual.ManualPromises(ctx, diff.Revision)
	if err != nil {
		return result, err
	}

	result.Review = &Review{
		Promises: reviewPromises(checked, manual),
		Changes:  diff.Entries,
		Rules:    reviewRules,
	}

	return result, nil
}

func reviewPromises(
	checked CheckResult,
	manual []ManualPromise,
) []ReviewPromise {
	promises := make([]ReviewPromise, 0, len(checked.Promises)+len(manual))

	for _, report := range checked.Promises {
		promises = append(promises, ReviewPromise{
			ID:         report.ID,
			Capability: report.Capability,
			Text:       report.Text,
			State:      report.State,
			Drifted:    report.Drifted,
		})
	}
	for _, promise := range manual {
		promises = append(promises, ReviewPromise{
			ID:         promise.ID,
			Capability: promise.Capability,
			Text:       promise.Text,
			State:      promise.State,
			Drifted:    promise.Drifted,
		})
	}

	return promises
}
