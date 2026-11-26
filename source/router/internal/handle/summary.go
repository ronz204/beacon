package handle

import (
	"context"
	"fmt"

	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
	"github.com/spf13/cobra"
)

type summaryData struct {
	Lines   []spider.BriefLine `json:"lines"`
	Omitted int                `json:"omitted"`
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

type summaryBuilder func(
	weaver.Weaver,
	context.Context,
) (spider.BriefView, []spider.Finding, error)

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

	view, findings, err := build(weaver.New(project), command.Context())
	if err != nil {
		return printer.Failure(errorFailure(err))
	}
	if failure := stateFailure(findings); failure != nil {
		return printer.Failure(*failure)
	}

	lines := make([]string, 0, len(view.Lines)+1)
	if counted {
		lines = append(lines, countLine(view.Signals))
	}
	for _, line := range view.Lines {
		lines = append(lines, line.Text)
	}

	return printer.Done(output.Result{
		Data:    summaryData{Lines: view.Lines, Omitted: view.Omitted},
		Signals: view.Signals,
		Lines:   lines,
	})
}

func countLine(signals []spider.Signal) string {
	counts := map[string]int{}
	for _, signal := range signals {
		counts[signal.Kind]++
	}

	return fmt.Sprintf(
		"status: %d red, %d drift, %d blocked, %d pending yes",
		counts["red"],
		counts["drift"],
		counts["blocked"],
		counts["pending_yes"],
	)
}
