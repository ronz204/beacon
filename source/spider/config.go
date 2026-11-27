package spider

import (
	"maps"
	"os"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

func (p Project) Runners() ([]RunnerConfig, []Finding) {
	config, failures := p.store.LoadConfig()
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

	return runners, p.findings(failures)
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
		Record: Record{Type: document.Kind, Data: document.JSON},
		File:   p.relative(file),
	}, nil
}
