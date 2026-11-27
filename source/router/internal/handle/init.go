package handle

import (
	"fmt"
	"os"
	"strings"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const (
	noStackHint = "no known stack was found: the runners stay empty; write " +
		"them in .beacon/config.yaml when the project has a test command"
	runnersHint = "the proposed runners run code in the CI, so they need a " +
		"yes: show them to the person, and if they say yes, save the " +
		"proposal from data.proposal to a file and run " +
		"beacon record --approve --file <path>"
)

func Init(printer *output.Printer) *cobra.Command {
	var preview bool

	command := &cobra.Command{
		Use:   "init",
		Short: "Start Beacon in an existing repository",
		Args:  cobra.NoArgs,
		RunE: execute(printer, func(
			command *cobra.Command,
			_ []string,
		) (output.Result, *output.Failure) {
			directory, err := os.Getwd()
			if err != nil {
				return output.Result{}, projectFailure(err)
			}

			view, err := spider.Init(
				command.Context(),
				directory,
				preview,
				func(project spider.Project) (string, error) {
					return weaver.New(project).Author(command.Context())
				},
			)
			if err != nil {
				return fail(err)
			}

			return output.Result{
				Data:  view,
				Hint:  initHint(view),
				Lines: initLines(view),
			}, nil
		}),
	}

	command.Flags().BoolVar(&preview, "preview", false, "show the plan only")

	return command
}

func initHint(view spider.InitView) string {
	if view.Proposal == "" {
		return noStackHint
	}

	return runnersHint
}

func initLines(view spider.InitView) []string {
	var lines []string

	if view.Preview {
		lines = append(lines, "preview: nothing was written")
	}

	lines = append(lines, "stacks: "+stacksText(view.Scan.Stacks))
	for _, runner := range view.Scan.Runners {
		lines = append(lines, fmt.Sprintf(
			"proposed runner %s: %s",
			runner.Name,
			strings.Join(runner.Run, " "),
		))
	}
	for _, record := range view.Records {
		lines = append(lines, initRecordLine(view.Preview, record))
	}
	if view.Omitted > 0 {
		lines = append(lines, fmt.Sprintf(
			"%d more folders are left in one question",
			view.Omitted,
		))
	}

	return lines
}

func stacksText(stacks []string) string {
	if len(stacks) == 0 {
		return "none found"
	}

	return strings.Join(stacks, ", ")
}

func initRecordLine(preview bool, record spider.InitRecord) string {
	if preview {
		return fmt.Sprintf("would create %s: %s", record.Type, record.Text)
	}

	return fmt.Sprintf(
		"created %s %s %s: %s",
		record.Type,
		record.ID,
		record.File,
		record.Text,
	)
}
