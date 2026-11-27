package handle

import (
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const (
	reasonRequiredCode    = "reason_required"
	reasonRequiredMessage = "closing a record needs a reason"
	reasonRequiredHint    = "pass --reason with why it no longer holds"
)

func Close(printer *output.Printer) *cobra.Command {
	var reason string
	var approve bool

	command := &cobra.Command{
		Use:   "close <id>",
		Short: "Close a record with a reason and archive it",
		Args:  cobra.ExactArgs(1),
		RunE: execute(printer, func(
			command *cobra.Command,
			args []string,
		) (output.Result, *output.Failure) {
			if strings.TrimSpace(reason) == "" {
				return output.Result{}, &output.Failure{
					Code:    reasonRequiredCode,
					Message: reasonRequiredMessage,
					Hint:    reasonRequiredHint,
				}
			}

			project, failure := locateProject()
			if failure != nil {
				return output.Result{}, failure
			}

			view, err := weaver.New(project).Close(
				command.Context(),
				args[0],
				reason,
				approve,
			)
			if err != nil {
				failure := errorFailure(err)
				failure.Lines = changeLines(view.Changes)

				return output.Result{}, failure
			}

			lines := append(
				[]string{fmt.Sprintf("closed %s %s", view.ID, view.File)},
				changeLines(view.Changes)...,
			)

			return succeed(view, lines...)
		}),
	}

	command.Flags().StringVar(&reason, "reason", "", "why it no longer holds")
	command.Flags().BoolVar(&approve, "approve", false, "pass the person's yes")

	return command
}
