package facade

import (
	"context"
	"slices"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

const violatedState = "violated"

type Zones struct {
	project spider.Project
}

func NewZones(project spider.Project) Zones {
	return Zones{project: project}
}

func (z Zones) Zones(
	ctx context.Context,
	revision string,
) ([]deltas.Zone, error) {
	touched, err := z.project.TouchedCapabilities(ctx, revision)
	if err != nil {
		return nil, err
	}

	view, findings := z.project.Analyze(ctx)
	if hasErrors(findings) {
		return nil, action.FailStateInvalid(findings)
	}

	zones := make([]deltas.Zone, 0, len(touched))
	for _, capability := range touched {
		zone := deltas.Zone{Capability: capability}

		for _, promise := range view.Promises {
			if promise.Capability != capability {
				continue
			}

			zone.Red = zone.Red || promise.State == violatedState
			zone.Drifted = zone.Drifted || promise.Drifted
		}

		zones = append(zones, zone)
	}

	slices.SortFunc(zones, func(a, b deltas.Zone) int {
		return compare(a.Capability, b.Capability)
	})

	return zones, nil
}

func compare(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
