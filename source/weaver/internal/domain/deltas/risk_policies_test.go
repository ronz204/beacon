package deltas_test

import (
	"slices"
	"testing"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

func findingAt(level deltas.ChangeLevel) deltas.Finding {
	return deltas.Finding{Level: level}
}

func TestRiskIsLowForMaintenanceAndAdditions(t *testing.T) {
	findings := []deltas.Finding{
		findingAt(deltas.LevelMaintenance),
		findingAt(deltas.LevelAddition),
	}
	zones := []deltas.Zone{{Capability: "bc-7f3a"}}

	got := deltas.AssessRisk(findings, zones, false)

	if got.Level != deltas.RiskLow || len(got.Reasons) != 0 {
		t.Errorf("got %+v", got)
	}
}

func TestRiskIsHighWithAWeakeningChange(t *testing.T) {
	findings := []deltas.Finding{findingAt(deltas.LevelWeakening)}

	got := deltas.AssessRisk(findings, nil, false)

	want := []deltas.RiskReason{deltas.RiskWeakening}
	if got.Level != deltas.RiskHigh || !slices.Equal(got.Reasons, want) {
		t.Errorf("got %+v", got)
	}
}

func TestRiskIsHighInAZoneWithRedOrDriftedPromises(t *testing.T) {
	cases := map[string]struct {
		zone deltas.Zone
		want deltas.RiskReason
	}{
		"red":   {deltas.Zone{Red: true}, deltas.RiskRed},
		"drift": {deltas.Zone{Drifted: true}, deltas.RiskDrift},
	}

	for name, c := range cases {
		got := deltas.AssessRisk(nil, []deltas.Zone{c.zone}, false)

		if got.Level != deltas.RiskHigh || got.Reasons[0] != c.want {
			t.Errorf("%s: got %+v", name, got)
		}
	}
}

func TestRiskListsEveryReasonOnce(t *testing.T) {
	findings := []deltas.Finding{
		findingAt(deltas.LevelWeakening),
		findingAt(deltas.LevelWeakening),
	}
	zones := []deltas.Zone{{Red: true}, {Red: true, Drifted: true}}

	got := deltas.AssessRisk(findings, zones, false)

	want := []deltas.RiskReason{
		deltas.RiskWeakening, deltas.RiskRed, deltas.RiskDrift,
	}
	if !slices.Equal(got.Reasons, want) {
		t.Errorf("reasons = %v, want %v", got.Reasons, want)
	}
}

func TestTheAgentCanRaiseTheRiskButNeverLowerIt(t *testing.T) {
	raised := deltas.AssessRisk(nil, nil, true)
	if raised.Level != deltas.RiskHigh {
		t.Errorf("raised on request = %+v", raised)
	}

	findings := []deltas.Finding{findingAt(deltas.LevelWeakening)}
	kept := deltas.AssessRisk(findings, nil, false)
	if kept.Level != deltas.RiskHigh {
		t.Errorf("without raising = %+v", kept)
	}
}
