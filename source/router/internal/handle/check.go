package handle

import (
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const (
	unavailableCode    = "runner_unavailable"
	unavailableMessage = "a runner program could not be started"
)

func Check(printer *output.Printer) *cobra.Command {
	var all bool
	var base string

	command := &cobra.Command{
		Use:   "check [capability]",
		Short: "Run the checks of the promises",
		Args:  cobra.MaximumNArgs(1),
		RunE: inProject(printer, func(
			command *cobra.Command,
			args []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			capability := ""
			if len(args) == 1 {
				capability = args[0]
			}

			engine := weaver.New(project)
			if !printer.JSON {
				engine = engine.Streaming(printer.Stderr)
			}

			view, signals, err := engine.Check(
				command.Context(),
				capability,
				all,
				base,
			)
			if err != nil {
				return fail(err)
			}

			return checkResult(view, signals)
		}),
	}

	command.Flags().BoolVar(&all, "all", false, "check every capability")
	command.Flags().StringVar(&base, "base", "", "branch to compare against")

	return command
}

func checkResult(
	view weaver.CheckView,
	signals []spider.Signal,
) (output.Result, *output.Failure) {
	result := output.Result{
		Data:    view,
		Signals: signals,
		Lines:   checkLines(view),
	}

	if len(view.Failed) > 0 {
		result.Exit = output.ExitBlocked
		result.Hint = failedHint(view)

		return result, nil
	}

	if unavailable := unavailableRuns(view); len(unavailable) > 0 {
		return output.Result{}, &output.Failure{
			Code:    unavailableCode,
			Message: unavailableMessage,
			Hint:    unavailableHint(unavailable[0]),
			Details: map[string]any{"invocations": unavailable},
			Signals: signals,
			Lines:   checkLines(view),
		}
	}

	return result, nil
}

func failedHint(view weaver.CheckView) string {
	failed := view.Failed[0]

	return fmt.Sprintf(
		"promise %s failed: fix the code until beacon check %s passes; "+
			"do not edit the promise",
		failed.ID,
		failed.Capability,
	)
}

func unavailableRuns(view weaver.CheckView) []weaver.InvocationView {
	var runs []weaver.InvocationView

	for _, invocation := range view.Invocations {
		if invocation.Status == "unavailable" {
			runs = append(runs, invocation)
		}
	}

	return runs
}

func unavailableHint(run weaver.InvocationView) string {
	return fmt.Sprintf(
		"install %s or fix runner %s in .beacon/config.yaml, then rerun "+
			"beacon check",
		run.Argv[0],
		run.Runner,
	)
}

func checkLines(view weaver.CheckView) []string {
	var lines []string

	if len(view.Capabilities) == 0 {
		lines = append(lines, noCapabilitiesLine(view))
	}
	for _, promise := range view.Promises {
		lines = append(lines, fmt.Sprintf(
			"%s %s (%s)%s",
			promise.Status,
			promise.ID,
			promise.Capability,
			reasonSuffix(promise.Reason),
		))
	}
	for _, run := range unavailableRuns(view) {
		lines = append(lines, "unavailable "+strings.Join(run.Promises, " "))
	}
	if len(view.Manual) > 0 {
		lines = append(lines, fmt.Sprintf(
			"%d manual promises wait for attest: %s",
			len(view.Manual),
			strings.Join(view.Manual, " "),
		))
	}

	return lines
}

func noCapabilitiesLine(view weaver.CheckView) string {
	if view.Base != "" {
		return "no capability touched since " + view.Base
	}

	return "no capability to check"
}

func reasonSuffix(reason string) string {
	if reason == "" {
		return ""
	}

	return " " + reason
}
