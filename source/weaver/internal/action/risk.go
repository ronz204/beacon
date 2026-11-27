package action

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type ZoneSource interface {
	Zones(ctx context.Context, revision string) ([]deltas.Zone, error)
}

func findingsOf(entries []Entry) []deltas.Finding {
	findings := make([]deltas.Finding, 0, len(entries))

	for _, entry := range entries {
		findings = append(findings, entry.Finding)
	}

	return findings
}
