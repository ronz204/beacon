package spider

import (
	"context"
	"math"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

func (p Project) Brief(
	ctx context.Context,
	pending []string,
) (BriefView, []Finding) {
	return p.brief(ctx, pending, false)
}

func (p Project) Status(
	ctx context.Context,
	pending []string,
) (BriefView, []Finding) {
	return p.brief(ctx, pending, true)
}

func (p Project) brief(
	ctx context.Context,
	pending []string,
	uncapped bool,
) (BriefView, []Finding) {
	sources, config, failures := p.sources(ctx)

	limit := config.Limits.Brief
	if uncapped {
		limit = math.MaxInt32
	}

	result, briefFailures := action.Brief(ctx, sources, limit, pending)

	return briefView(result), p.findings(failures, briefFailures)
}

func (p Project) Context(
	ctx context.Context,
	paths []string,
) (ContextView, []Finding) {
	sources, _, failures := p.sources(ctx)
	result, contextFailures := action.Context(ctx, sources, paths)

	return contextView(result), p.findings(failures, contextFailures)
}

func (p Project) HookContext(
	ctx context.Context,
	session string,
	paths []string,
) (HookView, []Finding, error) {
	id, err := kernel.ParseSessionID(session)
	if err != nil {
		return HookView{}, nil, err
	}

	log, err := states.Open(ctx, p.root)
	if err != nil {
		return HookView{}, nil, err
	}

	sources, config, failures := p.sources(ctx)
	result, hookFailures := action.HookContext(ctx, sources, action.HookInput{
		Session: id,
		Paths:   paths,
		Limit:   config.Limits.Hook,
		Log:     log,
	})
	view := HookView{Lines: result.Lines, Omitted: result.Omitted}

	return view, p.findings(failures, hookFailures), nil
}

func (p Project) Ready(
	ctx context.Context,
	branch string,
) (ReadyView, []Finding, error) {
	parsed, err := agenda.ParseTaskBranch(branch)
	if err != nil {
		return ReadyView{}, nil, err
	}

	claims, err := states.Open(ctx, p.root)
	if err != nil {
		return ReadyView{}, nil, err
	}

	result, failures := action.Ready(ctx, p.repositories(), claims, parsed)

	return readyView(result), p.findings(failures), nil
}

func (p Project) Analyze(ctx context.Context) (AnalysisView, []Finding) {
	sources, _, failures := p.sources(ctx)
	analysis, analysisFailures := action.Analyze(ctx, sources)

	return analysisView(analysis), p.findings(failures, analysisFailures)
}
