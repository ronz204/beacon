package facade

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

type Attests struct {
	project spider.Project
}

func NewAttests(project spider.Project) Attests {
	return Attests{project: project}
}

func (a Attests) ManualPromises(
	ctx context.Context,
	revision string,
) ([]action.ManualPromise, error) {
	promises, findings, err := a.project.ManualPromises(ctx, revision)
	if err != nil {
		return nil, err
	}
	if hasErrors(findings) {
		return nil, action.FailStateInvalid(findings)
	}

	manual := make([]action.ManualPromise, 0, len(promises))
	for _, promise := range promises {
		manual = append(manual, action.ManualPromise(promise))
	}

	return manual, nil
}

func (a Attests) Attest(
	ctx context.Context,
	promise string,
	stamp kernel.ApprovalInput,
	commit string,
	evidence string,
) error {
	return a.project.Attest(ctx, promise, stamp, commit, evidence)
}

func (a Attests) Head(ctx context.Context) (string, error) {
	return a.project.Head(ctx)
}
