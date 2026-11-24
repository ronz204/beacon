package kernel

import "errors"

type FailureView struct {
	Code    string
	Message string
	File    string
	Line    int
	Cause   string
	Details map[string]any
}

type describedFailure interface {
	failureCode() string
	failureMessage() string
	failureDetails() map[string]any
}

func DescribeFailure(err error) FailureView {
	var view FailureView

	for current := err; current != nil; current = errors.Unwrap(current) {
		described, isDomain := current.(describedFailure)

		switch {
		case isDomain:
			view.absorb(described)
		case view.Cause == "":
			view.Cause = current.Error()
		}
	}

	return view
}

func (v *FailureView) absorb(failure describedFailure) {
	v.Code = failure.failureCode()
	v.Message = failure.failureMessage()
	v.Cause = ""

	for key, value := range failure.failureDetails() {
		switch key {
		case "file":
			v.absorbFile(value)
		case "line":
			v.absorbLine(value)
		default:
			v.absorbDetail(key, value)
		}
	}
}

func (v *FailureView) absorbFile(value any) {
	if file, isText := value.(string); isText && v.File == "" {
		v.File = file
	}
}

func (v *FailureView) absorbLine(value any) {
	if line, isNumber := value.(int); isNumber && v.Line == 0 {
		v.Line = line
	}
}

func (v *FailureView) absorbDetail(key string, value any) {
	if v.Details == nil {
		v.Details = make(map[string]any)
	}

	v.Details[key] = value
}
