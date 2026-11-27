package handle

import (
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/snares"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/system"
	"github.com/spf13/cobra"
)

const (
	scopeCode      = "scope_invalid"
	scopeMessage   = "--scope accepts only all or touched"
	scopeHint      = "use --scope all or --scope touched"
	commentCode    = "comment_unwritable"
	commentMessage = "the comment file could not be written"
	commentHint    = "pass a --comment path in a folder that exists and " +
		"that you can write to"
)

func Snares(printer *output.Printer) *cobra.Command {
	command := &cobra.Command{
		Use:   "snares",
		Short: "Run the checks that do not depend on the agent",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(preCommit(printer), ci(printer))

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
	summary := fmt.Sprintf(
		"pre-commit: ok (%d capabilities checked)",
		len(view.Checked),
	)

	return snareResult(view, view.Blocks, view.Warnings, summary)
}

func snareResult(
	data any,
	blocks []snares.BlockView,
	warnings []string,
	summary string,
) output.Result {
	result := output.Result{
		Data:  data,
		Lines: blockLines(blocks, warnings, summary),
	}
	if len(blocks) == 0 {
		return result
	}

	hints := make([]string, 0, len(blocks))
	for _, block := range blocks {
		hints = append(hints, block.Hint)
	}

	result.Exit = output.ExitBlocked
	result.Hint = strings.Join(hints, "; ")

	return result
}

func blockLines(
	blocks []snares.BlockView,
	warnings []string,
	summary string,
) []string {
	var lines []string

	for _, block := range blocks {
		lines = append(lines, "blocked: "+block.Message)
		for _, item := range block.Items {
			lines = append(lines, "  "+item)
		}
	}
	for _, warning := range warnings {
		lines = append(lines, "warning: "+warning)
	}
	if len(lines) == 0 {
		lines = append(lines, summary)
	}

	return lines
}

func ci(printer *output.Printer) *cobra.Command {
	var base, scope, comment string

	command := &cobra.Command{
		Use:   "ci",
		Short: "Recalculate the checks and verify the yes of a pull request",
		Args:  cobra.NoArgs,
		RunE: inProject(printer, func(
			command *cobra.Command,
			_ []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			if scope != snares.ScopeAll && scope != snares.ScopeTouched {
				return output.Result{}, &output.Failure{
					Code:    scopeCode,
					Message: scopeMessage,
					Hint:    scopeHint,
				}
			}

			engine := snares.New(project)
			if !printer.JSON {
				engine = engine.Streaming(printer.Stderr)
			}

			view, err := engine.CI(command.Context(), snares.CIInput{
				Base:  base,
				Scope: scope,
			})
			if err != nil {
				return fail(err)
			}
			if failure := writeComment(comment, view.Comment); failure != nil {
				return output.Result{}, failure
			}

			summary := fmt.Sprintf(
				"ci: ok (%d capabilities checked)",
				len(view.Checked),
			)

			return snareResult(view, view.Blocks, view.Warnings, summary), nil
		}),
	}

	flags := command.Flags()
	flags.StringVar(&base, "base", "", "branch or commit to compare against")
	flags.StringVar(&scope, "scope", snares.ScopeAll, "all or touched")
	flags.StringVar(&comment, "comment", "", "write the pull request comment")

	return command
}

func writeComment(path, text string) *output.Failure {
	if path == "" {
		return nil
	}

	if err := system.WriteFileAtomic(path, []byte(text)); err != nil {
		return &output.Failure{
			Code:    commentCode,
			Message: commentMessage,
			Hint:    commentHint,
			Lines:   []string{err.Error()},
		}
	}

	return nil
}
