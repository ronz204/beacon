package handle

import (
	"github.com/ronz204/beacon/internal/router/internal/output"
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

func Doctor(printer *output.Printer, exit *int) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Validate the project state",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			project, failure := locateProject()
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			findings := project.Doctor(command.Context())
			details := doctorDetails{Findings: findings}

			if len(errorFindings(findings)) > 0 {
				*exit = printer.Failure(output.Failure{
					Code:    doctorFailedCode,
					Message: doctorFailedMessage,
					Hint:    findingsHint(findings),
					Details: details,
					Lines:   findingLines(findings),
				})
				return nil
			}

			lines := findingLines(findings)
			if len(lines) == 0 {
				lines = []string{doctorCleanMessage}
			}

			*exit = printer.Success(details, lines...)

			return nil
		},
	}
}
