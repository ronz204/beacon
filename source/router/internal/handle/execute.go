package handle

import (
	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
	"github.com/spf13/cobra"
)

type outcome func(
	command *cobra.Command,
	args []string,
) (output.Result, *output.Failure)

type projectOutcome func(
	command *cobra.Command,
	args []string,
	project spider.Project,
) (output.Result, *output.Failure)

func execute(
	printer *output.Printer,
	run outcome,
) func(*cobra.Command, []string) error {
	return func(command *cobra.Command, args []string) error {
		result, failure := run(command, args)
		printer.Report(result, failure)

		return nil
	}
}

func inProject(
	printer *output.Printer,
	run projectOutcome,
) func(*cobra.Command, []string) error {
	return execute(printer, func(
		command *cobra.Command,
		args []string,
	) (output.Result, *output.Failure) {
		project, failure := locateProject()
		if failure != nil {
			return output.Result{}, failure
		}

		return run(command, args, project)
	})
}

func succeed(data any, lines ...string) (output.Result, *output.Failure) {
	return output.Result{Data: data, Lines: lines}, nil
}

func fail(err error) (output.Result, *output.Failure) {
	return output.Result{}, errorFailure(err)
}
