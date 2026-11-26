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

func Record(printer *output.Printer, exit *int) *cobra.Command {
	var approve bool

	command := &cobra.Command{
		Use:   "record -",
		Short: "Create or replace a record read from stdin",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if args[0] != stdinMarker {
				*exit = printer.Failure(output.Failure{
					Code:    stdinMarkerCode,
					Message: stdinMarkerReason,
					Hint:    stdinMarkerHint,
				})
				return nil
			}

			project, failure := locateProject()
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			payload, err := io.ReadAll(command.InOrStdin())
			if err != nil {
				*exit = printer.Failure(errorFailure(err))
				return nil
			}

			view, err := weaver.New(project).Record(
				command.Context(),
				payload,
				approve,
			)
			if err != nil {
				failure := errorFailure(err)
				failure.Lines = changeLines(view)
				*exit = printer.Failure(failure)

				return nil
			}

			*exit = printer.Success(view, recordLines(view)...)

			return nil
		},
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

	return append(lines, changeLines(view)...)
}

func recordVerb(view weaver.RecordView) string {
	if view.Saved {
		return "recorded"
	}

	return "unchanged"
}

func changeLines(view weaver.RecordView) []string {
	lines := make([]string, 0, len(view.Changes))

	for _, change := range view.Changes {
		line := fmt.Sprintf(
			"level %d %s %s",
			change.Level,
			change.Target,
			change.ID,
		)
		if change.NeedsYes {
			line += " needs a yes"
		}

		lines = append(lines, line)
	}

	return lines
}
