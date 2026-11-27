package benchmark_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/benchmark"
	"github.com/ronz204/beacon/testing/utilities"
)

func TestAnUntouchedRepositoryDoesNotResolveTheTask(t *testing.T) {
	beacon := beaconIn(t)

	for _, name := range []string{"add-total", "tempt-weaken"} {
		directory := t.TempDir()
		chosen := task(t, name)

		err := benchmark.Prepare(
			directory, chosen, benchmark.ConditionFull, beacon,
		)
		if err != nil {
			t.Fatal(err)
		}

		outcome, err := benchmark.Inspect(directory, chosen, beacon)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if outcome.Resolved || outcome.Pending != 0 {
			t.Errorf("%s: outcome = %+v", name, outcome)
		}
	}
}

func TestInspectSeesAWeakenedPromiseWaitingForAYes(t *testing.T) {
	beacon := beaconIn(t)
	chosen := task(t, "tempt-weaken")
	directory := t.TempDir()

	err := benchmark.Prepare(directory, chosen, benchmark.ConditionFull, beacon)
	if err != nil {
		t.Fatal(err)
	}

	record := filepath.Join(directory, filepath.FromSlash(relativeRecord))
	text := utilities.ReadFile(t, record)
	weaker := strings.Replace(text, "18000 cents", "19000 cents", 1)
	utilities.WriteFile(t, directory, relativeRecord, weaker)

	outcome, err := benchmark.Inspect(directory, chosen, beacon)
	if err != nil {
		t.Fatal(err)
	}

	if outcome.Pending == 0 {
		t.Errorf("the weakened promise should be pending: %+v", outcome)
	}
}
