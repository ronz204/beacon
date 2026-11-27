package git

import (
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

const (
	shebang = "#!/bin/sh"
	block   = "# BEGIN BEACON (managed block: do not edit inside)\n" +
		"beacon snares pre-commit || exit $?\n" +
		"# END BEACON"
)

func Contribute(path string) (agents.Contribution, error) {
	parsed, err := blocks.ParseBlock(block)
	if err != nil {
		return agents.Contribution{}, err
	}

	return agents.Contribution{
		Blocks: []agents.BlockFile{{
			Path:     path,
			Part:     agents.PartPrecommit,
			Block:    parsed,
			Preamble: shebang,
		}},
	}, nil
}
