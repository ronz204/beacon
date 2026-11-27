package handle

import (
	"fmt"
	"io"
	"os"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const (
	stdinMarker  = "-"
	sourceCode   = "record_source_invalid"
	sourceReason = "the record is read from stdin with - " +
		"or from a file with --file"
	sourceBothReason = "pass the record on stdin or in a file, not both"
	sourceHint       = "run beacon record - with the record on stdin, " +
		"or beacon record --file <path>"
	fileCode   = "record_file_unreadable"
	fileReason = "the file passed to --file could not be read"
	fileHint   = "check the path passed to --file, then run it again"
)

func Record(printer *output.Printer) *cobra.Command {
	var approve bool
	var file string

	command := &cobra.Command{
		Use:   "record [-]",
		Short: "Create or replace a record read from stdin or a file",
		Args:  cobra.MaximumNArgs(1),
		RunE: execute(printer, func(
			command *cobra.Command,
			args []string,
		) (output.Result, *output.Failure) {
			if failure := sourceFailure(args, file); failure != nil {
				return output.Result{}, failure
			}

			project, failure := locateProject()
			if failure != nil {
				return output.Result{}, failure
			}

			payload, failure := readPayload(command, file)
			if failure != nil {
				return output.Result{}, failure
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
	command.Flags().StringVar(&file, "file", "", "read the record from a file")

	return command
}

func sourceFailure(args []string, file string) *output.Failure {
	fromStdin := len(args) == 1 && args[0] == stdinMarker

	switch {
	case fromStdin && file != "":
		return sourceError(sourceBothReason)
	case !fromStdin && file == "":
		return sourceError(sourceReason)
	case len(args) == 1 && !fromStdin:
		return sourceError(sourceReason)
	}

	return nil
}

func sourceError(reason string) *output.Failure {
	return &output.Failure{Code: sourceCode, Message: reason, Hint: sourceHint}
}

func readPayload(
	command *cobra.Command,
	file string,
) ([]byte, *output.Failure) {
	if file == "" {
		payload, err := io.ReadAll(command.InOrStdin())
		if err != nil {
			return nil, errorFailure(err)
		}

		return payload, nil
	}

	payload, err := os.ReadFile(file)
	if err != nil {
		return nil, &output.Failure{
			Code:    fileCode,
			Message: fileReason,
			Hint:    fileHint,
			Details: map[string]string{"file": file},
			Lines:   []string{err.Error()},
		}
	}

	return payload, nil
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
