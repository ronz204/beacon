package handle

import (
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

func Claim(printer *output.Printer) *cobra.Command {
	var force bool

	command := &cobra.Command{
		Use:   "claim <task>",
		Short: "Claim a task for the current branch",
		Args:  cobra.ExactArgs(1),
		RunE: inProject(printer, func(
			command *cobra.Command,
			args []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			view, err := weaver.New(project).Claim(
				command.Context(),
				args[0],
				force,
			)
			if err != nil {
				return fail(err)
			}

			lines := []string{fmt.Sprintf(
				"claimed %s on %s",
				view.Task,
				view.Branch,
			)}
			if view.Replaced != nil {
				lines = append(lines, fmt.Sprintf(
					"took it from branch %s",
					view.Replaced.Branch,
				))
			}

			return succeed(view, lines...)
		}),
	}

	command.Flags().BoolVar(
		&force,
		"force",
		false,
		"take a claimed task, only when the person asked",
	)

	return command
}
