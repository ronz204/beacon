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
	Exit   int
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

func (p *Printer) Report(result Result, failure *Failure) {
	if failure != nil {
		p.fail(*failure)
		return
	}

	p.done(result)
}

func (p *Printer) done(result Result) {
	if p.JSON {
		p.writeEnvelope(envelope{
			OK:      true,
			Data:    result.Data,
			Signals: signalList(result.Signals),
			Hint:    optionalHint(result.Hint),
		}, result.Exit)
		return
	}

	writeLines(p.Stdout, result.Lines)
	writeLines(p.Stdout, SignalLines(result.Signals))
	if result.Hint != "" {
		fmt.Fprintf(p.Stdout, "hint: %s\n", result.Hint)
	}

	p.Exit = result.Exit
}

func (p *Printer) fail(failure Failure) {
	code := failure.Exit
	if code == ExitOK {
		code = ExitUsage
	}

	if p.JSON {
		p.writeEnvelope(envelope{
			OK:      false,
			Signals: signalList(failure.Signals),
			Hint:    &failure.Hint,
			Error: &envelopeError{
				Code:    failure.Code,
				Message: failure.Message,
				Details: failure.Details,
			},
		}, code)
		return
	}

	writeLines(p.Stderr, failure.Lines)
	fmt.Fprintf(p.Stderr, "%s: %s\n", failure.Code, failure.Message)
	fmt.Fprintf(p.Stderr, "hint: %s\n", failure.Hint)

	p.Exit = code
}

func optionalHint(hint string) *string {
	if hint == "" {
		return nil
	}

	return &hint
}

func (p *Printer) writeEnvelope(content envelope, code int) {
	encoder := json.NewEncoder(p.Stdout)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(content); err != nil {
		fmt.Fprintln(p.Stderr, err)
		p.Exit = ExitUsage
		return
	}

	p.Exit = code
}

func writeLines(target io.Writer, lines []string) {
	for _, line := range lines {
		fmt.Fprintln(target, line)
	}
}
