package benchmark_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/benchmark"
	"github.com/ronz204/beacon/testing/utilities"
)

func TestConditionsInstallWhatTheProtocolSays(t *testing.T) {
	beacon := beaconIn(t)
	task := task(t, "add-total")

	cases := map[benchmark.Condition]struct {
		settings bool
		starts   int
		edits    int
	}{
		benchmark.ConditionWithout:    {false, 0, 0},
		benchmark.ConditionGuide:      {true, 0, 0},
		benchmark.ConditionFull:       {true, 2, 1},
		benchmark.ConditionNoReinject: {true, 1, 0},
	}

	for condition, want := range cases {
		directory := t.TempDir()
		err := benchmark.Prepare(directory, task, condition, beacon)
		if err != nil {
			t.Fatalf("%s: %v", condition, err)
		}

		path := filepath.Join(directory, ".claude", "settings.json")
		_, statErr := os.Stat(path)
		if (statErr == nil) != want.settings {
			t.Errorf("%s: settings present = %v", condition, statErr == nil)
		}
		if !want.settings {
			continue
		}

		text := utilities.ReadFile(t, path)
		gotStarts := strings.Count(text, "beacon hook start")
		gotEdits := strings.Count(text, "beacon hook edit")
		if gotStarts != want.starts || gotEdits != want.edits {
			t.Errorf("%s: start hooks %d, edit hooks %d",
				condition, gotStarts, gotEdits)
		}
	}
}

func TestThePreparedRepositoryIsHealthy(t *testing.T) {
	beacon := beaconIn(t)
	directory := t.TempDir()

	err := benchmark.Prepare(
		directory, task(t, "add-total"), benchmark.ConditionFull, beacon,
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := beacon(directory, "doctor"); err != nil {
		t.Errorf("doctor: %v", err)
	}
}
