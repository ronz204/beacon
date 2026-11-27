package deltas

import "slices"

func AssessRisk(
	findings []Finding,
	zones []Zone,
	raised bool,
) Assessment {
	var reasons []RiskReason

	if hasWeakening(findings) {
		reasons = append(reasons, RiskWeakening)
	}
	if anyZone(zones, func(zone Zone) bool { return zone.Red }) {
		reasons = append(reasons, RiskRed)
	}
	if anyZone(zones, func(zone Zone) bool { return zone.Drifted }) {
		reasons = append(reasons, RiskDrift)
	}
	if raised {
		reasons = append(reasons, RiskRaised)
	}

	if len(reasons) == 0 {
		return Assessment{Level: RiskLow, Reasons: []RiskReason{}}
	}

	return Assessment{Level: RiskHigh, Reasons: reasons}
}

func hasWeakening(findings []Finding) bool {
	for _, finding := range findings {
		if finding.Level == LevelWeakening {
			return true
		}
	}

	return false
}

func anyZone(zones []Zone, test func(Zone) bool) bool {
	return slices.ContainsFunc(zones, test)
}
