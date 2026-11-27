package handle

import (
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/snares"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

func Snares(printer *output.Printer) *cobra.Command {
	command := &cobra.Command{
		Use:   "snares",
		Short: "Run the checks that do not depend on the agent",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(preCommit(printer))

	return command
}

func preCommit(printer *output.Printer) *cobra.Command {
	return &cobra.Command{
		Use:   "pre-commit",
		Short: "Block a commit that weakens what is promised",
		Args:  cobra.NoArgs,
		RunE: inProject(printer, func(
			command *cobra.Command,
			_ []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			engine := snares.New(project)
			if !printer.JSON {
				engine = engine.Streaming(printer.Stderr)
			}

			view, err := engine.PreCommit(command.Context())
			if err != nil {
				return fail(err)
			}

			return preCommitResult(view), nil
		}),
	}
}

func preCommitResult(view snares.PreCommitView) output.Result {
	result := output.Result{Data: view, Lines: preCommitLines(view)}
	if len(view.Blocks) == 0 {
		return result
	}

	hints := make([]string, 0, len(view.Blocks))
	for _, block := range view.Blocks {
		hints = append(hints, block.Hint)
	}

	result.Exit = output.ExitBlocked
	result.Hint = strings.Join(hints, "; ")

	return result
}

func preCommitLines(view snares.PreCommitView) []string {
	var lines []string

	for _, block := range view.Blocks {
		lines = append(lines, "blocked: "+block.Message)
		for _, item := range block.Items {
			lines = append(lines, "  "+item)
		}
	}
	for _, warning := range view.Warnings {
		lines = append(lines, "warning: "+warning)
	}
	if len(lines) == 0 {
		lines = append(lines, fmt.Sprintf(
			"pre-commit: ok (%d capabilities checked)",
			len(view.Checked),
		))
	}

	return lines
}
