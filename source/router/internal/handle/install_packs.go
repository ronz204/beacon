package handle

import (
	"context"
	"os"
	"slices"

	"github.com/ronz204/beacon/source/anchor"
	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
)

const (
	defaultPack     = "core"
	defaultLanguage = "en"
	packsFile       = ".beacon/config.yaml"
)

type installResult struct {
	anchor.InstallView
	Packs    *spider.PacksConfig `json:"packs,omitempty"`
	Recorded bool                `json:"recorded"`
}

type packsOutcome struct {
	config   spider.PacksConfig
	changed  bool
	recorded bool
}

func foundProject() (spider.Project, bool) {
	directory, err := os.Getwd()
	if err != nil {
		return spider.Project{}, false
	}

	project, err := spider.Locate(directory)

	return project, err == nil
}

func resolvePacks(
	ctx context.Context,
	project spider.Project,
	tool anchor.Anchor,
	options anchor.InstallOptions,
	approve bool,
) (anchor.InstallOptions, packsOutcome, *output.Failure) {
	current, findings := project.Packs()
	if failure := stateFailure(findings); failure != nil {
		return options, packsOutcome{}, failure
	}

	scanned, err := project.Scan()
	if err != nil {
		return options, packsOutcome{}, errorFailure(err)
	}

	desired := desiredPacks(current, scanned.Stacks, options)
	resolved := optionsFor(options, desired)

	if err := tool.Validate(resolved); err != nil {
		return options, packsOutcome{}, errorFailure(err)
	}

	outcome := packsOutcome{
		config:  desired,
		changed: !samePacks(current, desired),
	}
	if !outcome.changed || options.Preview {
		return resolved, outcome, nil
	}

	failure := recordPacks(ctx, project, desired, approve)
	outcome.recorded = failure == nil

	return resolved, outcome, failure
}

func recordPacks(
	ctx context.Context,
	project spider.Project,
	desired spider.PacksConfig,
	approve bool,
) *output.Failure {
	payload, err := project.PacksPayload(desired)
	if err != nil {
		return errorFailure(err)
	}

	tool := weaver.New(project)
	yes := approve || tool.PersonPresent(ctx)

	view, err := tool.Record(ctx, payload, yes)
	if err != nil {
		failure := errorFailure(err)
		failure.Lines = changeLines(view.Changes)

		return failure
	}

	return nil
}

func desiredPacks(
	current spider.PacksConfig,
	stacks []string,
	options anchor.InstallOptions,
) spider.PacksConfig {
	desired := current
	desired.Present = true
	desired.Enabled = union(orDefault(current.Enabled), options.Packs)
	desired.Skip = union(current.Skip, options.SkipPieces)

	if len(desired.Stacks) == 0 && len(stacks) > 0 {
		desired.Stacks = stacks
	}
	if desired.Language.Code == "" {
		desired.Language.Code = defaultLanguage
	}

	return desired
}

func orDefault(packs []string) []string {
	if len(packs) == 0 {
		return []string{defaultPack}
	}

	return packs
}

func union(base, extra []string) []string {
	merged := slices.Clone(base)

	for _, name := range extra {
		if !slices.Contains(merged, name) {
			merged = append(merged, name)
		}
	}

	return merged
}

func samePacks(a, b spider.PacksConfig) bool {
	if a.Present != b.Present || a.Language != b.Language {
		return false
	}
	if a.Commit != b.Commit {
		return false
	}
	if !slices.Equal(a.Enabled, b.Enabled) {
		return false
	}

	return slices.Equal(a.Skip, b.Skip) && slices.Equal(a.Stacks, b.Stacks)
}

func optionsFor(
	options anchor.InstallOptions,
	packs spider.PacksConfig,
) anchor.InstallOptions {
	options.Packs = packs.Enabled
	options.SkipPieces = packs.Skip
	options.Reconcile = true
	options.Binding = anchor.PackBinding{
		Language:  packs.Language.Code,
		UI:        packs.Language.UI,
		CommitMin: packs.Commit.Min,
		CommitMax: packs.Commit.Max,
		Stacks:    packs.Stacks,
	}

	return options
}

func packsLines(preview bool, outcome packsOutcome) []string {
	switch {
	case outcome.recorded:
		return []string{"recorded packs in " + packsFile}
	case preview && outcome.changed:
		return []string{"would record packs in " + packsFile}
	}

	return nil
}
