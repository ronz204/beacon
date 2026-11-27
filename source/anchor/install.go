package anchor

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/driver/files"
)

const (
	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"
	binaryName   = "beacon"
	packsRoot    = "packs"
	configPath   = ".beacon/config.yaml"

	missingBinaryNote = "beacon is not on PATH; the agent calls it by " +
		"name, so install it with go install or add it to PATH"
	latestVersion       = "latest"
	unpinnedVersionNote = "this beacon has no release version, so the " +
		"workflow installs the latest one; push a tag or install a tagged " +
		"beacon and run install again to pin it"
	versionedHookNote = "the pre-commit hook is a file of the project, so " +
		"this edits it and the change is committed with the project: "
)

func (a Anchor) Install(
	ctx context.Context,
	options InstallOptions,
) (InstallView, error) {
	pieces, err := a.catalog.Select(options.Packs, options.SkipPieces)
	if err != nil {
		return InstallView{}, err
	}

	plan, err := a.planner.Plan(ctx, agentOptions(options, pieces))
	if err != nil {
		return InstallView{}, err
	}

	view, err := a.finish(ctx, plan, options)
	if err != nil {
		return InstallView{}, err
	}

	view.Skipped = skippedParts(options)
	view.Review = a.reviewNotes(options)

	return view, nil
}

func (a Anchor) Uninstall(
	ctx context.Context,
	options InstallOptions,
) (InstallView, error) {
	everything := InstallOptions{
		Preview:   options.Preview,
		Hooks:     true,
		Reinject:  true,
		Precommit: true,
	}

	pieces, err := a.catalog.Everything()
	if err != nil {
		return InstallView{}, err
	}

	plan, err := a.planner.Withdraw(ctx, agentOptions(everything, pieces))
	if err != nil {
		return InstallView{}, err
	}

	view, err := a.finish(ctx, plan, everything)
	if err != nil {
		return InstallView{}, err
	}

	view.Uninstall = true

	return view, nil
}

func (a Anchor) finish(
	ctx context.Context,
	plan action.Plan,
	options InstallOptions,
) (InstallView, error) {
	if !options.Preview {
		applier := action.Applier{
			Files: a.files,
			Stamp: time.Now().Format(files.StampLayout),
		}
		if err := applier.Apply(ctx, plan); err != nil {
			return InstallView{}, err
		}
	}

	_, initialized, err := a.files.Read(ctx, configPath)
	if err != nil {
		return InstallView{}, err
	}

	view := viewOf(plan, options.Preview)
	view.Initialized = initialized

	return view, nil
}

func agentOptions(
	options InstallOptions,
	pieces []agents.Piece,
) agents.Options {
	return agents.Options{
		Hooks:     options.Hooks,
		Reinject:  options.Reinject,
		Precommit: options.Precommit,
		CI:        options.CI,
		Pieces:    pieces,
	}
}

func skippedParts(options InstallOptions) []string {
	skipped := []string{}

	if !options.Hooks {
		skipped = append(
			skipped,
			string(agents.PartHooks),
			string(agents.PartReinject),
		)
	}
	if options.Hooks && !options.Reinject {
		skipped = append(skipped, string(agents.PartReinject))
	}
	if !options.Precommit {
		skipped = append(skipped, string(agents.PartPrecommit))
	}

	return append(skipped, options.SkipPieces...)
}

func (a Anchor) reviewNotes(options InstallOptions) []string {
	notes := []string{}

	if _, err := a.lookup(binaryName); err != nil {
		notes = append(notes, missingBinaryNote)
	}
	if options.CI != "" && !a.pinned {
		notes = append(notes, unpinnedVersionNote)
	}
	if options.Precommit && versionedHook(a.hookPath) {
		notes = append(notes, versionedHookNote+a.hookPath)
	}

	return notes
}

func versionedHook(path string) bool {
	return !filepath.IsAbs(path) && !strings.HasPrefix(path, ".git/")
}

func viewOf(plan action.Plan, preview bool) InstallView {
	view := InstallView{
		Preview: preview,
		Changes: []ChangeView{},
		Agents:  []AgentView{},
		Skipped: []string{},
		Kept:    []KeptView{},
		Review:  []string{},
	}

	for _, change := range plan.Changes {
		view.Changes = append(view.Changes, changeView(change))
	}
	for _, agent := range plan.Agents {
		view.Agents = append(view.Agents, agentView(agent))
	}
	for _, kept := range plan.Kept {
		view.Kept = append(view.Kept, KeptView{
			Path: kept.Path,
			Part: string(kept.Part),
		})
	}

	return view
}

func changeView(change action.Change) ChangeView {
	parts := make([]string, 0, len(change.Parts))
	for _, part := range change.Parts {
		parts = append(parts, string(part))
	}

	return ChangeView{
		Path:   change.Path,
		Action: actionOf(change),
		Parts:  parts,
		Before: change.Before,
		After:  change.After,
	}
}

func actionOf(change action.Change) string {
	switch {
	case change.Deleted:
		return actionDelete
	case change.Created:
		return actionCreate
	default:
		return actionUpdate
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
