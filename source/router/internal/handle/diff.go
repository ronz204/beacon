package handle

import (
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

func Diff(printer *output.Printer) *cobra.Command {
	var since, base string

	command := &cobra.Command{
		Use:   "diff",
		Short: "Show the state changes by level",
		Args:  cobra.NoArgs,
		RunE: inProject(printer, func(
			command *cobra.Command,
			_ []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			view, signals, err := weaver.New(project).Diff(
				command.Context(),
				since,
				base,
			)
			if err != nil {
				return fail(err)
			}

			result := output.Result{
				Data:    view,
				Signals: signals,
				Lines:   diffLines(view),
			}
			if len(view.Pending) > 0 {
				result.Hint = pendingHint(view)
			}

			return result, nil
		}),
	}

	command.Flags().StringVar(&since, "since", "", "compare against a revision")
	command.Flags().StringVar(&base, "base", "", "branch to compare against")

	return command
}

func pendingHint(view weaver.DiffView) string {
	return fmt.Sprintf(
		"%d changes wait for a yes: ask the person, and if they say yes, "+
			"send the record again with beacon record --approve -",
		len(view.Pending),
	)
}

func diffLines(view weaver.DiffView) []string {
	lines := []string{riskLine(view.Risk)}

	if len(view.Levels) == 0 {
		return append(lines, "no state changes since "+view.Revision)
	}

	for _, level := range view.Levels {
		lines = append(lines, fmt.Sprintf("level %d", level.Level))
		for _, entry := range level.Changes {
			lines = append(lines, entryLines(entry)...)
		}
	}

	return lines
}

func entryLines(entry weaver.EntryView) []string {
	line := fmt.Sprintf(
		"  %s %s %s %s",
		entry.Kind,
		entry.Path,
		entry.Target,
		entry.ID,
	)
	if entry.Reason != "" {
		line += " " + entry.Reason
	}
	if entry.NeedsYes && entry.Approved {
		line += " (approved)"
	}
	if entry.NeedsYes && !entry.Approved {
		line += " (needs a yes)"
	}

	lines := []string{line}
	if entry.Before != "" {
		lines = append(lines, "    - "+entry.Before)
	}
	if entry.After != "" {
		lines = append(lines, "    + "+entry.After)
	}

	return lines
}
