package weaver

import (
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
	"github.com/ronz204/beacon/source/weaver/internal/driver/facade"
)

type Weaver struct {
	changes deltas.ChangeSource
}

func New(project spider.Project) Weaver {
	return Weaver{changes: facade.NewChanges(project)}
}
