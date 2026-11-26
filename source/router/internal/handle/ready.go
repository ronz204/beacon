package handle

import (
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

const detachedHint = "check out a branch before claiming tasks"

type readyData struct {
	Branch spider.BranchView `json:"branch"`
	Works  []spider.Summary  `json:"works"`
	Tasks  []spider.Summary  `json:"tasks"`
}

func Ready(printer *output.Printer, exit *int) *cobra.Command {
	return &cobra.Command{
		Use:   "ready",
		Short: "Show the work ready on the current branch",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			project, failure := locateProject()
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			branch, err := project.Branch(command.Context())
			if err != nil {
				*exit = printer.Failure(errorFailure(err))
				return nil
			}

			view, findings, err := project.Ready(
				command.Context(),
				branch.Name,
			)
			if err != nil {
				*exit = printer.Failure(errorFailure(err))
				return nil
			}
			if failure := stateFailure(findings); failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			result := output.Result{
				Data: readyData{
					Branch: branch,
					Works:  view.Works,
					Tasks:  view.Tasks,
				},
				Lines: readyLines(branch, view),
			}
			if branch.Detached {
				result.Hint = detachedHint
			}

			*exit = printer.Done(result)

			return nil
		},
	}
}

func readyLines(branch spider.BranchView, view spider.ReadyView) []string {
	var lines []string

	if branch.Detached {
		lines = append(lines, "HEAD is detached, using commit "+branch.Name)
	}
	for _, work := range view.Works {
		lines = append(lines, summaryLine("work", work))
	}
	for _, task := range view.Tasks {
		lines = append(lines, summaryLine("task", task))
	}
	if len(view.Works) == 0 && len(view.Tasks) == 0 {
		lines = append(lines, fmt.Sprintf("nothing ready on %s", branch.Name))
	}

	return lines
}
