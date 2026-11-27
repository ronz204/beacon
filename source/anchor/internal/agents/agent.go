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
	PartRule       Part = "rule"
	PartSkill      Part = "skill"
)

type PieceKind string

const (
	PieceRule  PieceKind = "rule"
	PieceSkill PieceKind = "skill"
)

type Piece struct {
	Pack    string
	Kind    PieceKind
	Name    string
	Content string
}

type Options struct {
	Hooks     bool
	Reinject  bool
	Precommit bool
	CI        string
	Pieces    []Piece
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

type ManagedFile struct {
	Path    string
	Part    Part
	Content string
}

type Contribution struct {
	Blocks  []BlockFile
	Entries []EntryFile
	Files   []ManagedFile
}

type Agent interface {
	Name() string
	Moments() []Moment
	Contribute(options Options) (Contribution, error)
}
