package handle

import (
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

const sessionRequiredHint = "pass --session <id> with --hook"

func Context(printer *output.Printer) *cobra.Command {
	var hook bool
	var session string

	command := &cobra.Command{
		Use:   "context <paths>",
		Short: "Show what applies to the given files",
		Args:  cobra.MinimumNArgs(1),
		RunE: inProject(printer, func(
			command *cobra.Command,
			args []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			paths := make([]string, 0, len(args))
			for _, arg := range args {
				paths = append(paths, project.RepoPath(arg))
			}

			if hook {
				return hookContext(command, project, session, paths)
			}

			view, findings := project.Context(command.Context(), paths)
			if failure := stateFailure(findings); failure != nil {
				return output.Result{}, failure
			}

			return output.Result{
				Data:    view,
				Signals: view.Signals,
				Lines:   contextLines(view),
			}, nil
		}),
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
	project spider.Project,
	session string,
	paths []string,
) (output.Result, *output.Failure) {
	if session == "" {
		return output.Result{}, &output.Failure{
			Code:    "session_required",
			Message: "--hook needs a session id",
			Hint:    sessionRequiredHint,
		}
	}

	view, findings, err := project.HookContext(
		command.Context(),
		session,
		paths,
	)
	if err != nil {
		return fail(err)
	}
	if failure := stateFailure(findings); failure != nil {
		return output.Result{}, failure
	}

	lines := hookLines(view)

	return succeed(struct {
		Lines   []string `json:"lines"`
		Omitted int      `json:"omitted"`
	}{Lines: lines, Omitted: view.Omitted}, lines...)
}

func hookLines(view spider.HookView) []string {
	lines := append([]string{}, view.Lines...)
	if view.Omitted > 0 {
		lines = append(lines, fmt.Sprintf(
			"%d more, run beacon context <paths>",
			view.Omitted,
		))
	}

	return lines
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
