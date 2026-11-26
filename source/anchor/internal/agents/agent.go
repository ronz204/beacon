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
	PartImport     Part = "import"
	PartPermission Part = "permission"
	PartActor      Part = "actor"
	PartHooks      Part = "hooks"
	PartReinject   Part = "reinject"
)

type Options struct {
	Hooks    bool
	Reinject bool
}

type BlockFile struct {
	Path  string
	Part  Part
	Block blocks.Block
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
