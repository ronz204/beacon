package anchor

import (
	"context"
	"time"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/assets"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
	"github.com/ronz204/beacon/source/anchor/internal/driver/files"
)

const (
	actionCreate = "create"
	actionUpdate = "update"
	stampLayout  = "20060102T150405"
)

type Anchor struct {
	files   action.FileStore
	planner action.Planner
}

func New(ctx context.Context, directory string) (Anchor, error) {
	method, err := blocks.ParseBlock(assets.Method)
	if err != nil {
		return Anchor{}, err
	}

	repo, err := files.Open(ctx, directory)
	if err != nil {
		return Anchor{}, err
	}

	return Anchor{
		files:   repo,
		planner: action.Planner{Files: repo, Method: method},
	}, nil
}

func (a Anchor) Install(
	ctx context.Context,
	options InstallOptions,
) (InstallView, error) {
	plan, err := a.planner.Plan(ctx, agents.Options{
		Hooks:    options.Hooks,
		Reinject: options.Reinject,
	})
	if err != nil {
		return InstallView{}, err
	}

	if !options.DryRun {
		applier := action.Applier{
			Files: a.files,
			Stamp: time.Now().Format(stampLayout),
		}
		if err := applier.Apply(ctx, plan); err != nil {
			return InstallView{}, err
		}
	}

	return viewOf(plan, options.DryRun), nil
}

func viewOf(plan action.Plan, dryRun bool) InstallView {
	view := InstallView{
		DryRun:  dryRun,
		Changes: []ChangeView{},
		Agents:  []AgentView{},
	}

	for _, change := range plan.Changes {
		view.Changes = append(view.Changes, changeView(change))
	}
	for _, agent := range plan.Agents {
		view.Agents = append(view.Agents, agentView(agent))
	}

	return view
}

func changeView(change action.Change) ChangeView {
	verb := actionUpdate
	if change.Created {
		verb = actionCreate
	}

	parts := make([]string, 0, len(change.Parts))
	for _, part := range change.Parts {
		parts = append(parts, string(part))
	}

	return ChangeView{
		Path:   change.Path,
		Action: verb,
		Parts:  parts,
		Before: change.Before,
		After:  change.After,
	}
}

func agentView(agent action.AgentPlan) AgentView {
	return AgentView{
		Name:    agent.Name,
		Covers:  momentNames(agent.Covers),
		Missing: momentNames(agent.Missing),
	}
}

func momentNames(moments []agents.Moment) []string {
	names := make([]string, 0, len(moments))
	for _, moment := range moments {
		names = append(names, string(moment))
	}

	return names
}
