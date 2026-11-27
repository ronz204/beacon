package claude

import (
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

const (
	piecePrefix = "beacon-"
	rulesFolder = ".claude/rules/"
	skillFolder = ".claude/skills/"
	skillFile   = "/SKILL.md"
)

func pieceFiles(pieces []agents.Piece) ([]agents.ManagedFile, error) {
	files := make([]agents.ManagedFile, 0, len(pieces))

	for _, piece := range pieces {
		content, err := blocks.StampManaged(piece.Content)
		if err != nil {
			return nil, err
		}

		path, part := pieceTarget(piece)
		files = append(files, agents.ManagedFile{
			Path:    path,
			Part:    part,
			Content: content,
		})
	}

	return files, nil
}

func pieceTarget(piece agents.Piece) (string, agents.Part) {
	name := piecePrefix + piece.Name

	if piece.Kind == agents.PieceSkill {
		return skillFolder + name + skillFile, agents.PartSkill
	}

	return rulesFolder + name + ".md", agents.PartRule
}
