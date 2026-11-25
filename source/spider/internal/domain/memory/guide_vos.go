package memory

import (
	"slices"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/source/kernel"
)

type GuideID struct {
	value kernel.RecordID
}

func ParseGuideID(raw string) (GuideID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return GuideID{}, invalidGuide(GuideIDInvalid, raw)
	}

	return GuideID{value: id}, nil
}

func (id GuideID) String() string {
	return id.value.String()
}

type GuideCapabilityID struct {
	value kernel.RecordID
}

func ParseGuideCapabilityID(raw string) (GuideCapabilityID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return GuideCapabilityID{}, invalidGuide(GuideReferInvalid, raw)
	}

	return GuideCapabilityID{value: id}, nil
}

func (id GuideCapabilityID) String() string {
	return id.value.String()
}

type GuideKind string

const (
	GuideProduct      GuideKind = "product"
	GuideDomain       GuideKind = "domain"
	GuideArchitecture GuideKind = "architecture"
	GuideConvention   GuideKind = "convention"
	GuideRunbook      GuideKind = "runbook"
)

func GuideKinds() []GuideKind {
	return []GuideKind{
		GuideProduct,
		GuideDomain,
		GuideArchitecture,
		GuideConvention,
		GuideRunbook,
	}
}

func ParseGuideKind(raw string) (GuideKind, error) {
	kind := GuideKind(strings.TrimSpace(raw))

	if !slices.Contains(GuideKinds(), kind) {
		return "", invalidGuide(GuideKindInvalid, raw)
	}

	return kind, nil
}

type GuideAudience string

const (
	GuideForAgent GuideAudience = "agent"
	GuideForHuman GuideAudience = "human"
	GuideForBoth  GuideAudience = "both"
)

func GuideAudiences() []GuideAudience {
	return []GuideAudience{
		GuideForAgent,
		GuideForHuman,
		GuideForBoth,
	}
}

func ParseGuideAudience(raw string) (GuideAudience, error) {
	audience := GuideAudience(strings.TrimSpace(raw))

	if !slices.Contains(GuideAudiences(), audience) {
		return "", invalidGuide(GuideAudienceInvalid, raw)
	}

	return audience, nil
}

type GuideStatus string

const (
	GuideDraft   GuideStatus = "draft"
	GuideCurrent GuideStatus = "current"
)

func GuideStatuses() []GuideStatus {
	return []GuideStatus{
		GuideDraft,
		GuideCurrent,
	}
}

func ParseGuideStatus(raw string) (GuideStatus, error) {
	status := GuideStatus(strings.TrimSpace(raw))

	if !slices.Contains(GuideStatuses(), status) {
		return "", invalidGuide(GuideStatusInvalid, raw)
	}

	return status, nil
}

type GuideTag struct {
	value string
}

func ParseGuideTag(raw string) (GuideTag, error) {
	value := strings.TrimSpace(raw)

	if value == "" || strings.ContainsFunc(value, unicode.IsSpace) {
		return GuideTag{}, invalidGuide(GuideTagInvalid, raw)
	}

	return GuideTag{value: value}, nil
}

func (t GuideTag) String() string {
	return t.value
}
