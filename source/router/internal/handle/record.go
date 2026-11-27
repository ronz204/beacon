package handle

import (
	"fmt"
	"io"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const (
	stdinMarker       = "-"
	stdinMarkerHint   = "run beacon record - and send the record on stdin"
	stdinMarkerCode   = "record_source_invalid"
	stdinMarkerReason = "the record is read from stdin, pass - as the argument"
)

func Record(printer *output.Printer) *cobra.Command {
	var approve bool

	command := &cobra.Command{
		Use:   "record -",
		Short: "Create or replace a record read from stdin",
		Args:  cobra.ExactArgs(1),
		RunE: execute(printer, func(
			command *cobra.Command,
			args []string,
		) (output.Result, *output.Failure) {
			if args[0] != stdinMarker {
				return output.Result{}, &output.Failure{
					Code:    stdinMarkerCode,
					Message: stdinMarkerReason,
					Hint:    stdinMarkerHint,
				}
			}

			project, failure := locateProject()
			if failure != nil {
				return output.Result{}, failure
			}

			payload, err := io.ReadAll(command.InOrStdin())
			if err != nil {
				return fail(err)
			}

			view, err := weaver.New(project).Record(
				command.Context(),
				payload,
				approve,
			)
			if err != nil {
				failure := errorFailure(err)
				failure.Lines = changeLines(view.Changes)

				return output.Result{}, failure
			}

			return succeed(view, recordLines(view)...)
		}),
	}

	command.Flags().BoolVar(&approve, "approve", false, "pass the person's yes")

	return command
}

func recordLines(view weaver.RecordView) []string {
	lines := []string{fmt.Sprintf(
		"%s %s %s",
		recordVerb(view),
		view.ID,
		view.File,
	)}

	for _, ignored := range view.Ignored {
		lines = append(lines, "ignored field: "+ignored)
	}

	lines = append(lines, changeLines(view.Changes)...)

	return append(lines, guidanceLines(view.Guidance)...)
}

func guidanceLines(guidance []string) []string {
	if len(guidance) == 0 {
		return nil
	}

	lines := []string{"writing a promise:"}
	for _, line := range guidance {
		lines = append(lines, "  - "+line)
	}

	return lines
}

func quotedChange(change weaver.ChangeView) []string {
	var lines []string

	if change.Before != "" {
		lines = append(lines, "  - "+change.Before)
	}
	if change.After != "" {
		lines = append(lines, "  + "+change.After)
	}

	return lines
}

func recordVerb(view weaver.RecordView) string {
	if view.Saved {
		return "recorded"
	}

	return "unchanged"
}

func changeLines(changes []weaver.ChangeView) []string {
	lines := make([]string, 0, len(changes))

	for _, change := range changes {
		line := fmt.Sprintf(
			"level %d %s %s",
			change.Level,
			change.Target,
			change.ID,
		)
		if change.NeedsYes {
			line += " needs a yes: " + change.Reason
		}

		lines = append(lines, line)
		lines = append(lines, quotedChange(change)...)
	}

	return lines
}
