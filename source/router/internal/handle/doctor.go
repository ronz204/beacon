package handle

import (
	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

const (
	doctorFailedCode    = "doctor_failed"
	doctorFailedMessage = "the project state has problems"
	doctorCleanMessage  = "doctor: no problems found"
)

type doctorDetails struct {
	Findings any `json:"findings"`
}

func Doctor(printer *output.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Validate the project state",
		Args:  cobra.NoArgs,
		RunE: inProject(printer, func(
			command *cobra.Command,
			_ []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			findings := project.Doctor(command.Context())
			details := doctorDetails{Findings: findings}

			if len(errorFindings(findings)) > 0 {
				return output.Result{}, &output.Failure{
					Code:    doctorFailedCode,
					Message: doctorFailedMessage,
					Hint:    findingsHint(findings),
					Details: details,
					Lines:   findingLines(findings),
				}
			}

			lines := findingLines(findings)
			if len(lines) == 0 {
				lines = []string{doctorCleanMessage}
			}

			return succeed(details, lines...)
		}),
	}
}
