package github

import (
	"strings"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/assets"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

const (
	Name         = "github"
	workflowPath = ".github/workflows/beacon.yml"
	placeholder  = "{{version}}"
)

func Contribute(version string) (agents.Contribution, error) {
	text := strings.ReplaceAll(assets.Workflow, placeholder, version)

	block, err := blocks.ParseBlock(text)
	if err != nil {
		return agents.Contribution{}, err
	}

	return agents.Contribution{
		Blocks: []agents.BlockFile{{
			Path:  workflowPath,
			Part:  agents.PartCI,
			Block: block,
		}},
	}, nil
}
