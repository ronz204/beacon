package handle

import (
	"context"
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

const uncheckedHint = "no base branch was found, so pending yeses were " +
	"not checked; run beacon diff --base <branch> to check them"

type summaryData struct {
	Lines   []spider.BriefLine `json:"lines"`
	Omitted int                `json:"omitted"`
	Counts  map[string]int     `json:"counts,omitempty"`
}

func Brief(printer *output.Printer, exit *int) *cobra.Command {
	return &cobra.Command{
		Use:   "brief",
		Short: "Show the short state of the project",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			*exit = runSummary(command, printer, weaver.Weaver.Brief, false)
			return nil
		},
	}
}

func Status(printer *output.Printer, exit *int) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the whole state: reds, drift, blocks and pending yeses",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			*exit = runSummary(command, printer, weaver.Weaver.Status, true)
			return nil
		},
	}
}

type summaryBuilder func(weaver.Weaver, context.Context) (weaver.Summary, error)

func runSummary(
	command *cobra.Command,
	printer *output.Printer,
	build summaryBuilder,
	counted bool,
) int {
	project, failure := locateProject()
	if failure != nil {
		return printer.Failure(*failure)
	}

	summary, err := build(weaver.New(project), command.Context())
	if err != nil {
		return printer.Failure(errorFailure(err))
	}
	if failure := stateFailure(summary.Findings); failure != nil {
		return printer.Failure(*failure)
	}

	view := summary.View
	data := summaryData{Lines: view.Lines, Omitted: view.Omitted}
	lines := make([]string, 0, len(view.Lines)+1)
	if counted {
		data.Counts = countsOf(view.Signals)
		lines = append(lines, countLine(data.Counts))
	}
	for _, line := range view.Lines {
		lines = append(lines, line.Text)
	}

	result := output.Result{Data: data, Signals: view.Signals, Lines: lines}
	if !summary.PendingChecked {
		result.Hint = uncheckedHint
	}

	return printer.Done(result)
}

func countsOf(signals []spider.Signal) map[string]int {
	counts := map[string]int{
		"red": 0, "drift": 0, "blocked": 0, "pending_yes": 0,
	}

	for _, signal := range signals {
		if _, counted := counts[signal.Kind]; counted {
			counts[signal.Kind]++
		}
	}

	return counts
}

func countLine(counts map[string]int) string {
	return fmt.Sprintf(
		"status: %d red, %d drift, %d blocked, %d pending yes",
		counts["red"],
		counts["drift"],
		counts["blocked"],
		counts["pending_yes"],
	)
}
