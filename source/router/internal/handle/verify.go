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
	riskHigh      = "high"
	riskCode      = "risk_invalid"
	riskMessage   = "the agent can raise the risk, never lower it"
	riskHint      = "use --risk high or leave the flag out"
	reviewHeading = "review packet, for a reviewer with a clean context:"
)

const reviewHint = "the risk is high: give the review packet to an agent " +
	"with a clean context or to the person; the reviewer never commits " +
	"or pushes"

func Verify(printer *output.Printer) *cobra.Command {
	var risk, base string

	command := &cobra.Command{
		Use:   "verify",
		Short: "Run the related checks, rate the risk and prepare the review",
		Args:  cobra.NoArgs,
		RunE: inProject(printer, func(
			command *cobra.Command,
			_ []string,
			project spider.Project,
		) (output.Result, *output.Failure) {
			if risk != "" && risk != riskHigh {
				return output.Result{}, &output.Failure{
					Code:    riskCode,
					Message: riskMessage,
					Hint:    riskHint,
				}
			}

			engine := weaver.New(project)
			if !printer.JSON {
				engine = engine.Streaming(printer.Stderr)
			}

			view, signals, err := engine.Verify(
				command.Context(),
				base,
				risk == riskHigh,
			)
			if err != nil {
				return fail(err)
			}

			return verifyResult(view, signals)
		}),
	}

	command.Flags().StringVar(&risk, "risk", "", "raise the risk to high")
	command.Flags().StringVar(&base, "base", "", "branch to compare against")

	return command
}

func verifyResult(
	view weaver.VerifyView,
	signals []spider.Signal,
) (output.Result, *output.Failure) {
	checked, failure := checkResult(view.Check, signals)
	if failure != nil {
		return output.Result{}, failure
	}

	lines := []string{riskLine(view.Risk)}
	lines = append(lines, checkLines(view.Check)...)
	lines = append(lines, reviewLines(view.Review)...)

	result := output.Result{
		Data:    view,
		Signals: signals,
		Lines:   lines,
		Exit:    checked.Exit,
		Hint:    checked.Hint,
	}
	if result.Hint == "" && view.Risk.Level == riskHigh {
		result.Hint = reviewHint
	}

	return result, nil
}

func riskLine(risk weaver.RiskView) string {
	if len(risk.Reasons) == 0 {
		return "risk: " + risk.Level
	}

	return fmt.Sprintf(
		"risk: %s (%s)",
		risk.Level,
		strings.Join(risk.Reasons, ", "),
	)
}

func reviewLines(review *weaver.ReviewView) []string {
	if review == nil {
		return nil
	}

	lines := []string{reviewHeading}

	for _, promise := range review.Promises {
		state := promise.State
		if promise.Drifted {
			state += " (drift)"
		}

		lines = append(lines, fmt.Sprintf(
			"  promise %s (%s) [%s]: %s",
			promise.ID,
			promise.Capability,
			state,
			promise.Text,
		))
	}
	for _, change := range review.Changes {
		lines = append(lines, entryLines(change)...)
	}
	for _, rule := range review.Rules {
		lines = append(lines, "  rule: "+rule)
	}

	return lines
}
