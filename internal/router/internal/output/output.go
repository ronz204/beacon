package output

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	ExitOK    = 0
	ExitUsage = 1
)

type Printer struct {
	Stdout io.Writer
	Stderr io.Writer
	JSON   bool
}

type Failure struct {
	Code    string
	Message string
	Hint    string
	Details any
	Lines   []string
}

type envelope struct {
	OK      bool           `json:"ok"`
	Data    any            `json:"data"`
	Signals []any          `json:"signals"`
	Hint    *string        `json:"hint"`
	Error   *envelopeError `json:"error,omitempty"`
}

type envelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (p *Printer) Success(data any, lines ...string) int {
	if p.JSON {
		return p.writeEnvelope(envelope{
			OK:      true,
			Data:    data,
			Signals: []any{},
		}, ExitOK)
	}

	writeLines(p.Stdout, lines)

	return ExitOK
}

func (p *Printer) Failure(failure Failure) int {
	if p.JSON {
		return p.writeEnvelope(envelope{
			OK:      false,
			Signals: []any{},
			Hint:    &failure.Hint,
			Error: &envelopeError{
				Code:    failure.Code,
				Message: failure.Message,
				Details: failure.Details,
			},
		}, ExitUsage)
	}

	writeLines(p.Stderr, failure.Lines)
	fmt.Fprintf(p.Stderr, "%s: %s\n", failure.Code, failure.Message)
	fmt.Fprintf(p.Stderr, "hint: %s\n", failure.Hint)

	return ExitUsage
}

func (p *Printer) writeEnvelope(content envelope, code int) int {
	encoder := json.NewEncoder(p.Stdout)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(content); err != nil {
		fmt.Fprintln(p.Stderr, err)
		return ExitUsage
	}

	return code
}

func writeLines(target io.Writer, lines []string) {
	for _, line := range lines {
		fmt.Fprintln(target, line)
	}
}
