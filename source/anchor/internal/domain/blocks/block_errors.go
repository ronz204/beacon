package blocks

import "github.com/ronz204/beacon/source/kernel"

type BlockCode string

const (
	BlockInvalid  BlockCode = "block_invalid"
	BlockUnclosed BlockCode = "block_unclosed"
)

type BlockError = kernel.Error[BlockCode]

var blockErrors = kernel.ErrorCatalog[BlockCode]{
	BlockInvalid: "a block needs one begin line, one end line " +
		"and the same comment style in both",
	BlockUnclosed: "the file has a begin marker without its end marker",
}

func invalidBlock(raw string) BlockError {
	details := map[string]any{"value": raw}
	return blockErrors.Fail(BlockInvalid, details)
}

func FailBlockUnclosed(marker string) BlockError {
	details := map[string]any{"marker": marker}
	return blockErrors.Fail(BlockUnclosed, details)
}
