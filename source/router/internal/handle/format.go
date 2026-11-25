package handle

import (
	"fmt"
	"slices"

	"github.com/ronz204/beacon/source/router/internal/output"
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

func Format(printer *output.Printer, exit *int) *cobra.Command {
	return &cobra.Command{
		Use:   "fmt",
		Short: "Rewrite records in canonical format",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			project, failure := locateProject()
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			result := project.Format(command.Context())
			changed := slices.Concat([]string{}, result.Changed)

			if len(result.Findings) > 0 {
				*exit = printer.Failure(output.Failure{
					Code:    formatFailedCode,
					Message: formatFailedMessage,
					Hint:    fixRecordsHint,
					Details: formatDetails{
						Changed:  changed,
						Findings: result.Findings,
					},
					Lines: findingLines(result.Findings),
				})
				return nil
			}

			summary := fmt.Sprintf("fmt: %d files formatted", len(changed))
			lines := slices.Concat(changed, []string{summary})
			*exit = printer.Success(formatDetails{Changed: changed}, lines...)

			return nil
		},
	}
}
