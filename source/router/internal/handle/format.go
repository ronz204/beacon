package handle

import (
	"fmt"
	"slices"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

const (
	formatFailedCode    = "format_failed"
	formatFailedMessage = "some records could not be formatted"
)

type formatDetails struct {
	Changed  []string `json:"changed"`
	Findings any      `json:"findings,omitempty"`
}

func Format(printer *output.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "fmt",
		Short: "Rewrite records in canonical format",
		Args:  cobra.NoArgs,
		RunE: inProject(printer, func(
			command *cobra.Command,
			_ []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			result := project.Format(command.Context())
			changed := slices.Concat([]string{}, result.Changed)

			if len(result.Findings) > 0 {
				return output.Result{}, &output.Failure{
					Code:    formatFailedCode,
					Message: formatFailedMessage,
					Hint:    fixRecordsHint,
					Details: formatDetails{
						Changed:  changed,
						Findings: result.Findings,
					},
					Lines: findingLines(result.Findings),
				}
			}

			summary := fmt.Sprintf("fmt: %d files formatted", len(changed))
			lines := slices.Concat(changed, []string{summary})

			return succeed(formatDetails{Changed: changed}, lines...)
		}),
	}
}
