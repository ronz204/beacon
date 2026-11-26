package handle

import (
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

const sessionRequiredHint = "pass --session <id> with --hook"

func Context(printer *output.Printer, exit *int) *cobra.Command {
	var hook bool
	var session string

	command := &cobra.Command{
		Use:   "context <paths>",
		Short: "Show what applies to the given files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			project, failure := locateProject()
			if failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			paths := make([]string, 0, len(args))
			for _, arg := range args {
				paths = append(paths, project.RepoPath(arg))
			}

			if hook {
				*exit = hookContext(command, printer, project, session, paths)
				return nil
			}

			view, findings := project.Context(command.Context(), paths)
			if failure := stateFailure(findings); failure != nil {
				*exit = printer.Failure(*failure)
				return nil
			}

			*exit = printer.Done(output.Result{
				Data:    view,
				Signals: view.Signals,
				Lines:   contextLines(view),
			})

			return nil
		},
	}

	command.Flags().BoolVar(
		&hook,
		"hook",
		false,
		"only what is new this session",
	)
	command.Flags().StringVar(&session, "session", "", "session id for --hook")

	return command
}

func hookContext(
	command *cobra.Command,
	printer *output.Printer,
	project spider.Project,
	session string,
	paths []string,
) int {
	if session == "" {
		return printer.Failure(output.Failure{
			Code:    "session_required",
			Message: "--hook needs a session id",
			Hint:    sessionRequiredHint,
		})
	}

	view, findings, err := project.HookContext(
		command.Context(),
		session,
		paths,
	)
	if err != nil {
		return printer.Failure(errorFailure(err))
	}
	if failure := stateFailure(findings); failure != nil {
		return printer.Failure(*failure)
	}

	lines := append([]string{}, view.Lines...)
	if view.Omitted > 0 {
		lines = append(lines, fmt.Sprintf(
			"%d more, run beacon context <paths>",
			view.Omitted,
		))
	}

	return printer.Done(output.Result{
		Data: struct {
			Lines   []string `json:"lines"`
			Omitted int      `json:"omitted"`
		}{Lines: lines, Omitted: view.Omitted},
		Lines: lines,
	})
}

func contextLines(view spider.ContextView) []string {
	var lines []string

	for _, goal := range view.Goals {
		lines = append(lines, fmt.Sprintf("goal %s %s", goal.ID, goal.Text))
	}
	for _, capability := range view.Capabilities {
		lines = append(lines, fmt.Sprintf(
			"capability %s %s",
			capability.ID,
			capability.Name,
		))
		for _, promise := range capability.Promises {
			lines = append(lines, promiseLine(promise))
		}
	}
	for _, decision := range view.Decisions {
		lines = append(lines, summaryLine("decision", decision))
	}
	for _, guide := range view.Guides {
		lines = append(lines, summaryLine("guide", guide))
	}
	for _, question := range view.Questions {
		lines = append(lines, summaryLine("question", question))
	}

	return lines
}

func summaryLine(kind string, summary spider.Summary) string {
	return fmt.Sprintf("%s %s %s", kind, summary.ID, summary.Text)
}

func promiseLine(promise spider.PromiseView) string {
	state := promise.State
	if promise.Drifted {
		state += " (drift)"
	}

	return fmt.Sprintf(
		"  promise %s [%s] %s: %s",
		promise.ID,
		promise.Status,
		state,
		promise.Text,
	)
}
