package handle

import (
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

func Find(printer *output.Printer) *cobra.Command {
	var archived bool

	command := &cobra.Command{
		Use:   "find <text>",
		Short: "Search records by text",
		Args:  cobra.MinimumNArgs(1),
		RunE: inProject(printer, func(
			command *cobra.Command,
			args []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			text := strings.Join(args, " ")
			view, findings := project.Find(command.Context(), text, archived)
			if failure := stateFailure(findings); failure != nil {
				return output.Result{}, failure
			}

			lines := make([]string, 0, len(view.Matches)+1)
			for _, match := range view.Matches {
				lines = append(lines, fmt.Sprintf(
					"%s %s %s",
					match.Type,
					match.ID,
					match.Text,
				))
			}
			if view.Omitted > 0 {
				lines = append(lines, fmt.Sprintf(
					"%d more, narrow the search text",
					view.Omitted,
				))
			}

			return succeed(view, lines...)
		}),
	}

	command.Flags().BoolVar(
		&archived,
		"archived",
		false,
		"also search the archive",
	)

	return command
}
