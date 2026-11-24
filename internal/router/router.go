package router

import (
	"context"
	"io"
	"slices"

	"github.com/ronz204/beacon/internal/router/internal/handle"
	"github.com/ronz204/beacon/internal/router/internal/output"
	"github.com/spf13/cobra"
)

const (
	usageCode    = "usage_invalid"
	usageMessage = "the command or its arguments are not valid"
	usageHint    = "run beacon --help"
)

type usageDetails struct {
	Reason string `json:"reason"`
}

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	printer := &output.Printer{
		Stdout: stdout,
		Stderr: stderr,
		JSON:   slices.Contains(args, "--json"),
	}
	exit := output.ExitOK

	root := newRoot(printer, &exit)
	root.SetArgs(args)
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if err := root.ExecuteContext(context.Background()); err != nil {
		return printer.Failure(output.Failure{
			Code:    usageCode,
			Message: usageMessage,
			Hint:    usageHint,
			Details: usageDetails{Reason: err.Error()},
			Lines:   []string{err.Error()},
		})
	}

	return exit
}

func newRoot(printer *output.Printer, exit *int) *cobra.Command {
	root := &cobra.Command{
		Use:           "beacon",
		Short:         "Project memory with checked promises",
		SilenceUsage:  true,
		SilenceErrors: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	root.PersistentFlags().BoolVar(
		&printer.JSON,
		"json",
		printer.JSON,
		"print the result as JSON",
	)
	root.AddCommand(
		handle.Doctor(printer, exit),
		handle.Format(printer, exit),
	)

	return root
}
