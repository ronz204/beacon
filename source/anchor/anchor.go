package anchor

import (
	"context"
	"os/exec"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/agents/claude"
	"github.com/ronz204/beacon/source/anchor/internal/assets"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
	"github.com/ronz204/beacon/source/anchor/internal/driver/files"
)

type Anchor struct {
	files   action.FileStore
	planner action.Planner
	lookup  func(string) (string, error)
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

	return Anchor{
		files: repo,
		planner: action.Planner{
			Files:  repo,
			Shared: shared,
			Agents: []agents.Agent{claude.Agent{}},
		},
		lookup: exec.LookPath,
	}, nil
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
