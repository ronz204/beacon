package memory

import "github.com/ronz204/beacon/internal/kernel"

type DecisionCode string

const (
	DecisionIDInvalid     DecisionCode = "decision_id_invalid"
	DecisionStatusInvalid DecisionCode = "decision_status_invalid"
	DecisionFieldMissing  DecisionCode = "decision_field_missing"
	ReviewSpanInvalid     DecisionCode = "review_span_invalid"
)

type DecisionError = kernel.Error[DecisionCode]

var decisionErrors = kernel.ErrorCatalog[DecisionCode]{
	DecisionIDInvalid:     "decision id must be a root record id",
	DecisionStatusInvalid: "decision status must be active or superseded",
	DecisionFieldMissing:  "decision is missing a required field",
	ReviewSpanInvalid:     "review must look like 6mo or 30d",
}

func invalidDecision(code DecisionCode, raw string) DecisionError {
	details := map[string]any{"value": raw}
	return decisionErrors.Fail(code, details)
}
