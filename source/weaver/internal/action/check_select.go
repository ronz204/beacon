package action

import (
	"context"
	"slices"
)

func scopeOf(input CheckInput) string {
	switch {
	case input.Capability != "":
		return "capability"
	case input.All:
		return "all"
	}

	return "changed"
}

func (c Checker) choose(
	ctx context.Context,
	input CheckInput,
	targets []TargetDef,
	result *CheckResult,
) ([]TargetDef, error) {
	switch {
	case input.Capability != "":
		return oneTarget(targets, input.Capability)
	case input.All:
		return targets, nil
	}

	base, err := c.Base.Base(ctx, input.Base)
	if err != nil {
		return nil, err
	}

	result.Base = base
	touched, err := c.Source.Touched(ctx, base)
	if err != nil {
		return nil, err
	}

	return filterTargets(targets, touched), nil
}

func oneTarget(targets []TargetDef, id string) ([]TargetDef, error) {
	for _, target := range targets {
		if target.ID == id {
			return []TargetDef{target}, nil
		}
	}

	return nil, checkErrors.Fail(CapabilityMissing, map[string]any{"id": id})
}

func filterTargets(targets []TargetDef, ids []string) []TargetDef {
	var selected []TargetDef

	for _, target := range targets {
		if slices.Contains(ids, target.ID) {
			selected = append(selected, target)
		}
	}

	return selected
}

func idsOf(targets []TargetDef) []string {
	ids := make([]string, 0, len(targets))

	for _, target := range targets {
		ids = append(ids, target.ID)
	}

	return ids
}
