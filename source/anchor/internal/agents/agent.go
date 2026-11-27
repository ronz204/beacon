package agents

import "github.com/ronz204/beacon/source/anchor/internal/domain/blocks"

type Moment string

const (
	MomentStart   Moment = "start"
	MomentCompact Moment = "compact"
	MomentEdit    Moment = "edit"
	MomentStop    Moment = "stop"
)

func Moments() []Moment {
	return []Moment{MomentStart, MomentCompact, MomentEdit, MomentStop}
}

type Part string

const (
	PartGuide      Part = "guide"
	PartAttributes Part = "attributes"
	PartImport     Part = "import"
	PartPermission Part = "permission"
	PartActor      Part = "actor"
	PartHooks      Part = "hooks"
	PartReinject   Part = "reinject"
	PartPrecommit  Part = "precommit"
	PartCI         Part = "ci"
)

type Options struct {
	Hooks     bool
	Reinject  bool
	Precommit bool
	CI        string
}

type BlockFile struct {
	Path     string
	Part     Part
	Block    blocks.Block
	Preamble string
}

type EntryFile struct {
	Path    string
	Part    Part
	Entries []blocks.Entry
}

type Contribution struct {
	Blocks  []BlockFile
	Entries []EntryFile
}

type Agent interface {
	Name() string
	Moments() []Moment
	Contribute(options Options) (Contribution, error)
}
