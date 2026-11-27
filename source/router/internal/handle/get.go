package handle

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

func Get(printer *output.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show one record, active or archived",
		Args:  cobra.ExactArgs(1),
		RunE: inProject(printer, func(
			command *cobra.Command,
			args []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			stored, err := project.Get(command.Context(), args[0])
			if err != nil {
				return fail(err)
			}

			header := fmt.Sprintf("%s %s", stored.Type, stored.File)
			if stored.Archived {
				header += " (archived)"
			}

			var pretty bytes.Buffer
			_ = json.Indent(&pretty, stored.Data, "", "  ")

			return succeed(stored, header, pretty.String())
		}),
	}
}
