package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

type HookInput struct {
	Session kernel.SessionID
	Paths   []string
	Limit   int
	Log     kernel.InjectionLog
}

type HookResult struct {
	Lines   []string
	Omitted int
}

type hookItem struct {
	key   string
	lines []string
}

func HookContext(
	ctx context.Context,
	sources Sources,
	input HookInput,
) (HookResult, []error) {
	light := sources
	light.Changes = nil
	light.History = nil

	result, failures := Context(ctx, light, input.Paths)

	injected, err := input.Log.InjectedKeys(ctx, input.Session)
	if err != nil {
		return HookResult{}, append(failures, err)
	}

	fresh := freshItems(hookItems(result), injected)
	hook, emitted := fitHook(fresh, input.Limit, input.Paths)

	if len(emitted) > 0 {
		err := input.Log.RecordInjected(ctx, input.Session, emitted)
		if err != nil {
			failures = append(failures, err)
		}
	}

	return hook, failures
}

func hookItems(result ContextResult) []hookItem {
	items := signalItems(result.Signals)

	for _, capability := range result.Capabilities {
		lines := []string{fmt.Sprintf(
			"Capability %s: %s",
			capability.Capability.ID(),
			firstLine(capability.Capability.Name().String()),
		)}
		for _, nongoal := range capability.Capability.Nongoals() {
			lines = append(lines, "Not doing: "+firstLine(nongoal.String()))
		}
		items = append(items, item(capability.Capability.ID(), lines...))

		for _, promise := range capability.Promises {
			items = append(items, item(promise.Promise.ID(), fmt.Sprintf(
				"Promise %s (%s): %s",
				promise.Promise.ID(),
				promise.Verification.Status(),
				firstLine(promise.Promise.Text().String()),
			)))
		}
	}

	for _, decision := range result.Decisions {
		lines := []string{fmt.Sprintf(
			"Decision %s: %s",
			decision.ID(),
			firstLine(decision.What().String()),
		)}
		for _, rejected := range decision.Rejected() {
			lines = append(lines, "Rejected: "+firstLine(rejected.String()))
		}
		items = append(items, item(decision.ID(), lines...))
	}

	for _, guide := range result.Guides {
		items = append(items, item(guide.ID(), fmt.Sprintf(
			"Guide %s (%s): %s",
			guide.ID(),
			guide.Kind(),
			firstLine(guide.Title().String()),
		)))
	}

	for _, question := range result.Questions {
		items = append(items, item(question.ID(), fmt.Sprintf(
			"Open question %s: %s",
			question.ID(),
			firstLine(question.Text().String()),
		)))
	}

	return items
}

func signalItems(signals []thread.Signal) []hookItem {
	var items []hookItem

	for _, signal := range signals {
		if signal.Kind == thread.SignalTension {
			continue
		}

		key := fmt.Sprintf("signal:%s:%s", signal.Kind, signal.Subject)
		text := fmt.Sprintf("Signal %s: %s", signal.Kind, signal.Subject)
		items = append(items, hookItem{key: key, lines: []string{text}})
	}

	return items
}

func item(id fmt.Stringer, lines ...string) hookItem {
	return hookItem{key: "record:" + id.String(), lines: lines}
}

func freshItems(items []hookItem, injected map[string]bool) []hookItem {
	var fresh []hookItem

	for _, candidate := range items {
		if !injected[candidate.key] {
			fresh = append(fresh, candidate)
		}
	}

	return fresh
}

func fitHook(
	items []hookItem,
	limit int,
	paths []string,
) (HookResult, []string) {
	var result HookResult
	var emitted []string

	for index, candidate := range items {
		reserve := min(len(items)-index-1, 1)
		room := limit - len(result.Lines) - reserve
		first := len(result.Lines) == 0

		if first {
			keep := min(len(candidate.lines), max(room, 1))
			candidate.lines = candidate.lines[:keep]
		}
		if !first && len(candidate.lines) > room {
			result.Omitted = len(items) - index
			result.Lines = appendSummary(result, limit, paths)
			break
		}

		result.Lines = append(result.Lines, candidate.lines...)
		emitted = append(emitted, candidate.key)
	}

	return result, emitted
}

func appendSummary(result HookResult, limit int, paths []string) []string {
	if len(result.Lines) >= limit {
		return result.Lines
	}

	summary := fmt.Sprintf(
		"%d more, run beacon context %s",
		result.Omitted,
		strings.Join(paths, " "),
	)

	return append(result.Lines, summary)
}
