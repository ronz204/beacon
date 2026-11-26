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

func Close(printer *output.Printer, exit *int) *cobra.Command {
	var reason string
	var approve bool

	command := &cobra.Command{
		Use:   "close <id>",
		Short: "Close a record with a reason and archive it",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if strings.TrimSpace(reason) == "" {
				*exit = printer.Failure(output.Failure{
					Code:    reasonRequiredCode,
					Message: reasonRequiredMessage,
					Hint:    reasonRequiredHint,
				})
				return nil
			}

			project, failure := locateProject()
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
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
				*exit = printer.Failure(failure)

				return nil
			}

			lines := append(
				[]string{fmt.Sprintf("closed %s %s", view.ID, view.File)},
				changeLines(view.Changes)...,
			)
			*exit = printer.Success(view, lines...)

			return nil
		},
	}

	command.Flags().StringVar(&reason, "reason", "", "why it no longer holds")
	command.Flags().BoolVar(&approve, "approve", false, "pass the person's yes")

	return command
}
