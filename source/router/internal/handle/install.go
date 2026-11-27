package handle

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/anchor"
	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/spf13/cobra"
)

const (
	skipHooks     = "hooks"
	skipReinject  = "reinject"
	skipPrecommit = "precommit"
	skipCode      = "skip_invalid"
	skipMessage   = "--skip accepts only hooks, reinject or precommit"
	skipHint      = "use --skip hooks, --skip reinject or --skip precommit"
	ciGithub      = "github"
	ciCode        = "ci_invalid"
	ciMessage     = "--ci accepts only github"
	ciHint        = "use --ci github or leave the flag out"
)

var skipParts = []string{skipHooks, skipReinject, skipPrecommit}

func Install(printer *output.Printer) *cobra.Command {
	var preview, uninstall bool
	var skip []string
	var ci string

	command := &cobra.Command{
		Use:   "install",
		Short: "Install Beacon in the coding agents of this repository",
		Args:  cobra.NoArgs,
		RunE: execute(printer, func(
			command *cobra.Command,
			_ []string,
		) (output.Result, *output.Failure) {
			options, failure := installOptions(preview, skip, ci)
			if failure != nil {
				return output.Result{}, failure
			}

			directory, err := os.Getwd()
			if err != nil {
				return output.Result{}, projectFailure(err)
			}

			tool, err := anchor.New(command.Context(), directory)
			if err != nil {
				return fail(err)
			}

			run := tool.Install
			if uninstall {
				run = tool.Uninstall
			}

			view, err := run(command.Context(), options)
			if err != nil {
				return fail(err)
			}

			return output.Result{
				Data:  view,
				Hint:  installHint(view),
				Lines: installLines(view),
			}, nil
		}),
	}

	flags := command.Flags()
	flags.BoolVar(&preview, "preview", false, "show the changes without writing")
	flags.BoolVar(&uninstall, "uninstall", false, "remove what install added")
	flags.StringVar(&ci, "ci", "", "add the CI workflow of a platform: github")
	flags.StringSliceVar(&skip, "skip", nil, "leave out hooks, reinject or precommit")

	return command
}

func installOptions(
	preview bool,
	skip []string,
	ci string,
) (anchor.InstallOptions, *output.Failure) {
	if ci != "" && ci != ciGithub {
		return anchor.InstallOptions{}, &output.Failure{
			Code:    ciCode,
			Message: ciMessage,
			Hint:    ciHint,
		}
	}

	for _, part := range skip {
		if !slices.Contains(skipParts, part) {
			return anchor.InstallOptions{}, &output.Failure{
				Code:    skipCode,
				Message: skipMessage,
				Hint:    skipHint,
				Lines:   []string{"unknown part: " + part},
			}
		}
	}

	hooks := !slices.Contains(skip, skipHooks)

	return anchor.InstallOptions{
		Preview:   preview,
		Hooks:     hooks,
		Reinject:  hooks && !slices.Contains(skip, skipReinject),
		Precommit: !slices.Contains(skip, skipPrecommit),
		CI:        ci,
	}, nil
}

func installHint(view anchor.InstallView) string {
	if view.Initialized || view.Uninstall {
		return ""
	}

	return createProjectHint
}

func installLines(view anchor.InstallView) []string {
	var lines []string

	if view.Preview {
		lines = append(lines, "preview: nothing was written")
	}

	for _, change := range view.Changes {
		lines = append(lines, changeLine(view.Preview, change))

		if view.Preview {
			lines = append(lines, addedLines(change)...)
		}
	}
	if len(view.Changes) == 0 {
		lines = append(lines, "nothing to change")
	}

	for _, kept := range view.Kept {
		lines = append(lines, fmt.Sprintf(
			"kept %s (%s): edited by hand, remove it yourself",
			kept.Path,
			kept.Part,
		))
	}
	if len(view.Skipped) > 0 {
		lines = append(lines, "skipped: "+strings.Join(view.Skipped, ", "))
	}
	for _, note := range view.Review {
		lines = append(lines, "review: "+note)
	}

	return lines
}

func changeLine(preview bool, change anchor.ChangeView) string {
	verb := change.Action
	if preview {
		verb = "would " + verb
	}

	return fmt.Sprintf(
		"%s %s (%s)",
		verb,
		change.Path,
		strings.Join(change.Parts, ", "),
	)
}

func addedLines(change anchor.ChangeView) []string {
	if change.Action == "delete" {
		return nil
	}

	known := map[string]bool{}
	for line := range strings.SplitSeq(change.Before, "\n") {
		known[line] = true
	}

	var added []string
	for line := range strings.SplitSeq(change.After, "\n") {
		if !known[line] && line != "" {
			added = append(added, "  + "+strings.TrimSuffix(line, "\r"))
		}
	}

	return added
}
