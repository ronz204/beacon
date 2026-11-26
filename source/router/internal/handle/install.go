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
	skipHooks    = "hooks"
	skipReinject = "reinject"
	skipCode     = "skip_invalid"
	skipMessage  = "--skip accepts only hooks or reinject"
	skipHint     = "use --skip hooks or --skip reinject"
)

func Install(printer *output.Printer, exit *int) *cobra.Command {
	var dryRun, uninstall bool
	var skip []string

	command := &cobra.Command{
		Use:   "install",
		Short: "Install Beacon in the coding agents of this repository",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			options, failure := installOptions(dryRun, skip)
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			directory, err := os.Getwd()
			if err != nil {
				*exit = printer.Failure(*projectFailure(err))
				return nil
			}

			tool, err := anchor.New(command.Context(), directory)
			if err != nil {
				*exit = printer.Failure(errorFailure(err))
				return nil
			}

			run := tool.Install
			if uninstall {
				run = tool.Uninstall
			}

			view, err := run(command.Context(), options)
			if err != nil {
				*exit = printer.Failure(errorFailure(err))
				return nil
			}

			*exit = printer.Done(output.Result{
				Data:  view,
				Hint:  installHint(view),
				Lines: installLines(view),
			})

			return nil
		},
	}

	flags := command.Flags()
	flags.BoolVar(&dryRun, "dry-run", false, "show the changes without writing")
	flags.BoolVar(&uninstall, "uninstall", false, "remove what install added")
	flags.StringSliceVar(&skip, "skip", nil, "leave out hooks or reinject")

	return command
}

func installOptions(
	dryRun bool,
	skip []string,
) (anchor.InstallOptions, *output.Failure) {
	for _, part := range skip {
		if part != skipHooks && part != skipReinject {
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
		DryRun:   dryRun,
		Hooks:    hooks,
		Reinject: hooks && !slices.Contains(skip, skipReinject),
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

	if view.DryRun {
		lines = append(lines, "dry run: nothing was written")
	}

	for _, change := range view.Changes {
		lines = append(lines, changeLine(view.DryRun, change))

		if view.DryRun {
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

func changeLine(dryRun bool, change anchor.ChangeView) string {
	verb := change.Action
	if dryRun {
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
	for _, line := range strings.Split(change.Before, "\n") {
		known[line] = true
	}

	var added []string
	for _, line := range strings.Split(change.After, "\n") {
		if !known[line] && line != "" {
			added = append(added, "  + "+strings.TrimSuffix(line, "\r"))
		}
	}

	return added
}
