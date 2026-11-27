package weaver

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/action"
)

func (w Weaver) Attest(
	ctx context.Context,
	promises []string,
	evidence string,
	base string,
) (AttestView, error) {
	result, err := w.attester.Attest(ctx, action.AttestInput{
		Promises: promises,
		Evidence: evidence,
		Base:     base,
	})

	return attestView(result), err
}

func attestView(result action.AttestResult) AttestView {
	view := AttestView{
		Pending:  make([]ManualView, 0, len(result.Pending)),
		Attested: nonNil(result.Attested),
		Commit:   result.Commit,
	}

	for _, promise := range result.Pending {
		view.Pending = append(view.Pending, ManualView(promise))
	}

	return view
}
