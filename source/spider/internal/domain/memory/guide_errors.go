package memory

import "github.com/ronz204/beacon/source/kernel"

type GuideCode string

const (
	GuideIDInvalid       GuideCode = "guide_id_invalid"
	GuideKindInvalid     GuideCode = "guide_kind_invalid"
	GuideAudienceInvalid GuideCode = "guide_audience_invalid"
	GuideStatusInvalid   GuideCode = "guide_status_invalid"
	GuideTagInvalid      GuideCode = "guide_tag_invalid"
	GuideReferInvalid    GuideCode = "guide_refer_invalid"
	GuideFieldMissing    GuideCode = "guide_field_missing"
)

type GuideError = kernel.Error[GuideCode]

var guideErrors = kernel.ErrorCatalog[GuideCode]{
	GuideIDInvalid:       "guide id must be a root record id",
	GuideKindInvalid:     "guide kind is not in the closed set",
	GuideAudienceInvalid: "guide audience must be agent, human or both",
	GuideStatusInvalid:   "guide status must be draft or current",
	GuideTagInvalid:      "guide tag must be one word",
	GuideReferInvalid:    "guide refers must be root capability ids",
	GuideFieldMissing:    "guide is missing a required field",
}

func invalidGuide(code GuideCode, raw string) GuideError {
	details := map[string]any{"value": raw}
	return guideErrors.Fail(code, details)
}
