package handle

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/spf13/cobra"
)

func Get(printer *output.Printer, exit *int) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show one record, active or archived",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			project, failure := locateProject()
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			stored, err := project.Get(command.Context(), args[0])
			if err != nil {
				*exit = printer.Failure(errorFailure(err))
				return nil
			}

			header := fmt.Sprintf("%s %s", stored.Type, stored.File)
			if stored.Archived {
				header += " (archived)"
			}

			var pretty bytes.Buffer
			_ = json.Indent(&pretty, stored.Data, "", "  ")

			*exit = printer.Success(stored, header, pretty.String())

			return nil
		},
	}
}
