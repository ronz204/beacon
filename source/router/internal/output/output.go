package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ronz204/beacon/source/spider"
)

const (
	ExitOK      = 0
	ExitUsage   = 1
	ExitBlocked = 2
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
	Exit    int
	Signals []spider.Signal
	Lines   []string
}

type Result struct {
	Data    any
	Signals []spider.Signal
	Hint    string
	Exit    int
	Lines   []string
}

type envelope struct {
	OK      bool            `json:"ok"`
	Data    any             `json:"data"`
	Signals []spider.Signal `json:"signals"`
	Hint    *string         `json:"hint"`
	Error   *envelopeError  `json:"error,omitempty"`
}

type envelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (p *Printer) Success(data any, lines ...string) int {
	return p.Done(Result{Data: data, Lines: lines})
}

func (p *Printer) Done(result Result) int {
	if p.JSON {
		return p.writeEnvelope(envelope{
			OK:      true,
			Data:    result.Data,
			Signals: signalList(result.Signals),
			Hint:    optionalHint(result.Hint),
		}, result.Exit)
	}

	writeLines(p.Stdout, result.Lines)
	writeLines(p.Stdout, SignalLines(result.Signals))
	if result.Hint != "" {
		fmt.Fprintf(p.Stdout, "hint: %s\n", result.Hint)
	}

	return result.Exit
}

func (p *Printer) Failure(failure Failure) int {
	code := failure.Exit
	if code == ExitOK {
		code = ExitUsage
	}

	if p.JSON {
		return p.writeEnvelope(envelope{
			OK:      false,
			Signals: signalList(failure.Signals),
			Hint:    &failure.Hint,
			Error: &envelopeError{
				Code:    failure.Code,
				Message: failure.Message,
				Details: failure.Details,
			},
		}, code)
	}

	writeLines(p.Stderr, failure.Lines)
	fmt.Fprintf(p.Stderr, "%s: %s\n", failure.Code, failure.Message)
	fmt.Fprintf(p.Stderr, "hint: %s\n", failure.Hint)

	return code
}

func optionalHint(hint string) *string {
	if hint == "" {
		return nil
	}

	return &hint
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
