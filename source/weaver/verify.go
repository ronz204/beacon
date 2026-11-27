package weaver

import (
	"context"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

func (w Weaver) Verify(
	ctx context.Context,
	base string,
	raise bool,
) (VerifyView, []spider.Signal, error) {
	verifier := action.Verifier{
		Checker: w.checker,
		Differ:  w.differ,
		Manual:  w.attester.Ports,
	}

	result, err := verifier.Verify(ctx, action.VerifyInput{
		Base:  base,
		Raise: raise,
	})

	return verifyView(result), signalViews(result.Check.Signals), err
}

func verifyView(result action.VerifyResult) VerifyView {
	view := VerifyView{
		Revision: result.Revision,
		Risk:     riskView(result.Risk),
		Check:    checkView(result.Check),
	}

	if result.Review != nil {
		view.Review = reviewView(*result.Review)
	}

	return view
}

func riskView(assessment deltas.Assessment) RiskView {
	reasons := make([]string, 0, len(assessment.Reasons))
	for _, reason := range assessment.Reasons {
		reasons = append(reasons, string(reason))
	}

	return RiskView{Level: string(assessment.Level), Reasons: reasons}
}

func reviewView(review action.Review) *ReviewView {
	view := &ReviewView{
		Promises: make([]ReviewPromiseView, 0, len(review.Promises)),
		Changes:  entryViews(review.Changes),
		Rules:    review.Rules,
	}

	for _, promise := range review.Promises {
		view.Promises = append(view.Promises, ReviewPromiseView(promise))
	}

	return view
}
