package anchor

import (
	"context"
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/agents/claude"
	"github.com/ronz204/beacon/source/anchor/internal/agents/git"
	"github.com/ronz204/beacon/source/anchor/internal/agents/github"
	"github.com/ronz204/beacon/source/anchor/internal/assets"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
	"github.com/ronz204/beacon/source/anchor/internal/driver/files"
)

type Anchor struct {
	files    action.FileStore
	planner  action.Planner
	lookup   func(string) (string, error)
	hookPath string
	pinned   bool
}

func New(ctx context.Context, directory string) (Anchor, error) {
	shared, err := sharedContribution()
	if err != nil {
		return Anchor{}, err
	}

	repo, err := files.Open(ctx, directory)
	if err != nil {
		return Anchor{}, err
	}

	hook, err := git.Contribute(repo.HookPath())
	if err != nil {
		return Anchor{}, err
	}

	version, pinned := buildVersion()
	workflow, err := github.Contribute(version)
	if err != nil {
		return Anchor{}, err
	}

	return Anchor{
		files: repo,
		planner: action.Planner{
			Files:  repo,
			Shared: shared,
			Agents: []agents.Agent{claude.Agent{}},
			Git:    hook,
			CI:     map[string]agents.Contribution{github.Name: workflow},
		},
		lookup:   exec.LookPath,
		hookPath: repo.HookPath(),
		pinned:   pinned,
	}, nil
}

func buildVersion() (string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return latestVersion, false
	}

	version := info.Main.Version
	unusable := version == "" || version == "(devel)"
	if unusable || strings.HasSuffix(version, "+dirty") {
		return latestVersion, false
	}

	return version, true
}

func sharedContribution() (agents.Contribution, error) {
	method, err := blocks.ParseBlock(assets.Method)
	if err != nil {
		return agents.Contribution{}, err
	}

	attributes, err := blocks.ParseBlock(assets.Attributes)
	if err != nil {
		return agents.Contribution{}, err
	}

	return agents.Contribution{
		Blocks: []agents.BlockFile{
			{Path: "AGENTS.md", Part: agents.PartGuide, Block: method},
			{
				Path:  ".gitattributes",
				Part:  agents.PartAttributes,
				Block: attributes,
			},
		},
	}, nil
}
