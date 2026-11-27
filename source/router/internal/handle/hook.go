package handle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/ronz204/beacon/source/anchor"
	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const (
	hookInputLimit = 1 << 20
	hookStart      = "start"
	hookEdit       = "edit"
)

func Hook(printer *output.Printer) *cobra.Command {
	var agent string

	command := &cobra.Command{
		Use:   "hook <event>",
		Short: "Answer an agent hook: start or edit",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			err := runHook(command, printer, agent, args[0])
			if err != nil {
				fmt.Fprintf(printer.Stderr, "beacon hook: %v\n", err)
			}

			return nil
		},
	}

	command.Flags().StringVar(&agent, "agent", "", "agent that calls the hook")

	return command
}

func runHook(
	command *cobra.Command,
	printer *output.Printer,
	agent, event string,
) error {
	reader := io.LimitReader(command.InOrStdin(), hookInputLimit)

	stdin, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	input, err := anchor.ParseHook(agent, event, stdin)
	if err != nil {
		return err
	}
	if input.Skip {
		return nil
	}

	project, err := spider.Locate(workingDirectory(input))
	if err != nil {
		return nil
	}

	ctx := command.Context()

	text, err := hookText(ctx, project, input, event)
	if err != nil {
		return err
	}

	rendered, err := anchor.RenderHook(agent, event, text)
	if err != nil {
		return err
	}

	if _, err = printer.Stdout.Write(rendered); err != nil {
		return err
	}

	lines := injectedLines(text)
	counted := project.RecordInjection(ctx, input.Session, event, lines)
	noted := project.NoteAgent(ctx, agent, input.Session)

	return errors.Join(counted, noted)
}

func injectedLines(text string) int {
	if text == "" {
		return 0
	}

	return strings.Count(text, "\n") + 1
}

func workingDirectory(input anchor.HookInput) string {
	if input.Cwd == "" {
		return "."
	}

	return input.Cwd
}

func hookText(
	ctx context.Context,
	project spider.Project,
	input anchor.HookInput,
	event string,
) (string, error) {
	if event == hookEdit {
		return editText(ctx, project, input)
	}

	return startText(ctx, project)
}

func startText(ctx context.Context, project spider.Project) (string, error) {
	summary, err := weaver.New(project).Brief(ctx)
	if err != nil {
		return "", err
	}
	if failure := stateFailure(summary.Findings); failure != nil {
		return "", fmt.Errorf("%s: %s", failure.Code, failure.Message)
	}

	lines := make([]string, 0, len(summary.View.Lines))
	for _, line := range summary.View.Lines {
		lines = append(lines, line.Text)
	}

	return strings.Join(lines, "\n"), nil
}

func editText(
	ctx context.Context,
	project spider.Project,
	input anchor.HookInput,
) (string, error) {
	relative := project.RepoPath(input.Path)
	if outsideRepository(relative) {
		return "", nil
	}

	view, findings, err := project.HookContext(
		ctx,
		input.Session,
		[]string{relative},
	)
	if err != nil {
		return "", err
	}
	if failure := stateFailure(findings); failure != nil {
		return "", fmt.Errorf("%s: %s", failure.Code, failure.Message)
	}

	return strings.Join(hookLines(view), "\n"), nil
}

func outsideRepository(relative string) bool {
	return relative == ".." ||
		strings.HasPrefix(relative, "../") ||
		path.IsAbs(relative) ||
		strings.Contains(relative, ":")
}
