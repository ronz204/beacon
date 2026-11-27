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
	ciGithub      = "github"
	ciCode        = "ci_invalid"
	ciMessage     = "--ci accepts only github"
	ciHint        = "use --ci github or leave the flag out"
)

var skipParts = []string{skipHooks, skipReinject, skipPrecommit}

func Install(printer *output.Printer) *cobra.Command {
	var preview, uninstall, approve bool
	var skip, packs []string
	var ci string

	command := &cobra.Command{
		Use:   "install",
		Short: "Install Beacon in the coding agents of this repository",
		Args:  cobra.NoArgs,
		RunE: execute(printer, func(
			command *cobra.Command,
			_ []string,
		) (output.Result, *output.Failure) {
			options, failure := installOptions(preview, skip, packs, ci)
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

			project, inProject := foundProject()
			saving := inProject && !uninstall
			outcome := packsOutcome{}

			if saving {
				options, outcome, failure = resolvePacks(
					command.Context(),
					project,
					tool,
					options,
					approve,
				)
			} else {
				options.Packs = orDefault(options.Packs)
			}
			if failure != nil {
				return output.Result{}, failure
			}

			run := tool.Install
			if uninstall {
				run = tool.Uninstall
			}

			view, err := run(command.Context(), options)
			if err != nil {
				return fail(err)
			}

			result := installResult{
				InstallView: view,
				Recorded:    outcome.recorded,
			}
			if saving {
				result.Packs = &outcome.config
			}

			return output.Result{
				Data:  result,
				Hint:  installHint(view),
				Lines: installLines(view, outcome),
			}, nil
		}),
	}

	flags := command.Flags()
	flags.BoolVar(&preview, "preview", false, "show the changes without writing")
	flags.BoolVar(&uninstall, "uninstall", false, "remove what install added")
	flags.StringVar(&ci, "ci", "", "add the CI workflow of a platform: github")
	flags.StringSliceVar(
		&skip,
		"skip",
		nil,
		"leave out hooks, reinject, precommit or a piece by name",
	)
	flags.BoolVar(
		&approve,
		"approve",
		false,
		"pass the person's yes to a change of the packs",
	)
	flags.StringSliceVar(
		&packs,
		"pack",
		nil,
		"packs of rules and skills to install (default core)",
	)

	return command
}

func installOptions(
	preview bool,
	skip, packs []string,
	ci string,
) (anchor.InstallOptions, *output.Failure) {
	if ci != "" && ci != ciGithub {
		return anchor.InstallOptions{}, &output.Failure{
			Code:    ciCode,
			Message: ciMessage,
			Hint:    ciHint,
		}
	}

	hooks := !slices.Contains(skip, skipHooks)

	return anchor.InstallOptions{
		Preview:    preview,
		Hooks:      hooks,
		Reinject:   hooks && !slices.Contains(skip, skipReinject),
		Precommit:  !slices.Contains(skip, skipPrecommit),
		CI:         ci,
		Packs:      packs,
		SkipPieces: pieceNames(skip),
	}, nil
}

func pieceNames(skip []string) []string {
	return slices.DeleteFunc(slices.Clone(skip), func(name string) bool {
		return slices.Contains(skipParts, name)
	})
}

func installHint(view anchor.InstallView) string {
	if view.Initialized || view.Uninstall {
		return ""
	}

	return createProjectHint
}

func installLines(
	view anchor.InstallView,
	outcome packsOutcome,
) []string {
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

	lines = append(lines, packsLines(view.Preview, outcome)...)

	for _, kept := range view.Kept {
		lines = append(lines, fmt.Sprintf(
			"kept %s (%s): edited by hand, left as it is",
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
