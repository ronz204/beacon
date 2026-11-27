package handle

import (
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const (
	attestAskHint = "ask the person, one promise at a time, whether it " +
		"holds; for each yes run beacon attest <promise> after the code " +
		"is committed, adding --evidence <ref> if they point to one"
	nothingToAttest = "no manual promise waits for attest"
)

func Attest(printer *output.Printer) *cobra.Command {
	var evidence, base string

	command := &cobra.Command{
		Use:   "attest [promise...]",
		Short: "List the manual promises to attest, or attest them",
		RunE: inProject(printer, func(
			command *cobra.Command,
			args []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			view, err := weaver.New(project).Attest(
				command.Context(),
				args,
				evidence,
				base,
			)
			if err != nil {
				return fail(err)
			}

			return attestResult(view, len(args) > 0), nil
		}),
	}

	command.Flags().StringVar(
		&evidence,
		"evidence",
		"",
		"path or link to what the person reviewed",
	)
	command.Flags().StringVar(&base, "base", "", "branch to compare against")

	return command
}

func attestResult(view weaver.AttestView, attested bool) output.Result {
	if attested {
		return output.Result{Data: view, Lines: attestedLines(view)}
	}

	result := output.Result{Data: view, Lines: pendingLines(view)}
	if len(view.Pending) > 0 {
		result.Hint = attestAskHint
	}

	return result
}

func attestedLines(view weaver.AttestView) []string {
	lines := make([]string, 0, len(view.Attested))

	for _, promise := range view.Attested {
		lines = append(lines, fmt.Sprintf(
			"attested %s at %s",
			promise,
			view.Commit,
		))
	}

	return lines
}

func pendingLines(view weaver.AttestView) []string {
	if len(view.Pending) == 0 {
		return []string{nothingToAttest}
	}

	var lines []string
	for _, promise := range view.Pending {
		state := promise.State
		if promise.Drifted {
			state += " (drift)"
		}

		lines = append(
			lines,
			fmt.Sprintf(
				"manual %s (%s) %s: %s",
				promise.ID,
				promise.Capability,
				state,
				promise.Text,
			),
			"  steps: "+promise.Manual,
		)
	}

	return lines
}
