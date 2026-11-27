package spider

import (
	"context"
	"maps"
	"os"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

func (p Project) Runners() ([]RunnerConfig, []Finding) {
	config, failures := p.store.LoadConfig()

	return runnersOf(config), p.findings(failures)
}

func (p Project) RunnersAt(
	ctx context.Context,
	revision string,
) ([]RunnerConfig, error) {
	data, found, err := commit.NewRepository(p.root).ConfigAt(ctx, revision)
	if err != nil || !found {
		return []RunnerConfig{}, err
	}

	config, failures := record.ParseConfig(p.store.ConfigFile(), data)
	if len(failures) > 0 {
		return nil, failures[0]
	}

	return runnersOf(config), nil
}

func runnersOf(config record.Config) []RunnerConfig {
	runners := make([]RunnerConfig, 0, len(config.Runners))

	for _, name := range slices.Sorted(maps.Keys(config.Runners)) {
		runner := config.Runners[name]
		view := RunnerConfig{
			Name:     name,
			Run:      slices.Clone(runner.Run),
			Baseline: make([]string, 0, len(runner.Baseline)),
		}
		for _, glob := range runner.Baseline {
			view.Baseline = append(view.Baseline, glob.String())
		}

		runners = append(runners, view)
	}

	return runners
}

func (p Project) Config() (StoredRecord, error) {
	file := p.store.ConfigFile()

	data, err := os.ReadFile(file)
	if err != nil {
		return StoredRecord{}, kernel.NewRecordFailure(p.relative(file), 0, err)
	}

	document, err := record.ReadDocument(file, data)
	if err != nil {
		return StoredRecord{}, err
	}

	return StoredRecord{
		Type: document.Kind, Data: document.JSON,
		File: p.relative(file),
	}, nil
}
