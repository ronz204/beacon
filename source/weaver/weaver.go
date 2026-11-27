package weaver

import (
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
	"github.com/ronz204/beacon/source/weaver/internal/driver/branch"
	"github.com/ronz204/beacon/source/weaver/internal/driver/facade"
	"github.com/ronz204/beacon/source/weaver/internal/driver/identity"
	"github.com/ronz204/beacon/source/weaver/internal/driver/runner"
)

type Weaver struct {
	project  spider.Project
	changes  deltas.ChangeSource
	records  action.Recorder
	closer   action.Closer
	claimer  action.Claimer
	checker  action.Checker
	differ   action.Differ
	attester action.Attester
}

func New(project spider.Project) Weaver {
	today, _ := kernel.CalendarDateOf(time.Now())
	records := facade.NewRecords(project)
	who := identity.Git{ActiveAgent: project.ActiveAgent}

	differ := action.Differ{
		Changes: facade.NewChanges(project),
		Base:    branch.Git{},
		Zones:   facade.NewZones(project),
	}

	return Weaver{
		project: project,
		changes: facade.NewChanges(project),
		records: action.Recorder{
			Store:    records,
			Identity: who,
			Today:    today,
			Pending:  differ,
		},
		closer: action.Closer{
			Store:    facade.NewCloses(project),
			Identity: who,
			Today:    today,
		},
		claimer: action.Claimer{
			Ports:    facade.NewTasks(project),
			Store:    records,
			Identity: who,
		},
		differ: differ,
		attester: action.Attester{
			Ports:    facade.NewAttests(project),
			Base:     branch.Git{},
			Identity: who,
			Today:    today,
		},
		checker: action.Checker{
			Source:   facade.NewChecks(project),
			Base:     branch.Git{},
			Executor: runner.New(nil),
		},
	}
}
