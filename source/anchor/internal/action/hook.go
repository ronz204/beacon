package action

import (
	"slices"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/kernel"
)

type HookCode string

const (
	HookAgentUnknown HookCode = "hook_agent_unknown"
	HookEventUnknown HookCode = "hook_event_unknown"
	HookInputInvalid HookCode = "hook_input_invalid"
)

var hookErrors = kernel.ErrorCatalog[HookCode]{
	HookAgentUnknown: "the agent has no hook adapter",
	HookEventUnknown: "the hook event must be start or edit",
	HookInputInvalid: "the agent sent hook input Beacon cannot read",
}

type Hooks struct {
	Agents []agents.Hooker
}

func (h Hooks) Parse(
	agent, event string,
	stdin []byte,
) (agents.HookInput, error) {
	hooker, kind, err := h.resolve(agent, event)
	if err != nil {
		return agents.HookInput{}, err
	}

	input, err := hooker.ParseHook(kind, stdin)
	if err != nil {
		return agents.HookInput{}, hookErrors.Fail(
			HookInputInvalid,
			map[string]any{"agent": agent},
		).WithCause(err)
	}

	return input, nil
}

func (h Hooks) Render(agent, event, text string) ([]byte, error) {
	hooker, kind, err := h.resolve(agent, event)
	if err != nil {
		return nil, err
	}

	return hooker.RenderHook(kind, text)
}

func (h Hooks) resolve(
	agent, event string,
) (agents.Hooker, agents.Event, error) {
	kind := agents.Event(event)
	if !slices.Contains(agents.Events(), kind) {
		details := map[string]any{"event": event}
		return nil, "", hookErrors.Fail(HookEventUnknown, details)
	}

	for _, hooker := range h.Agents {
		if hooker.Name() == agent {
			return hooker, kind, nil
		}
	}

	details := map[string]any{"agent": agent}

	return nil, "", hookErrors.Fail(HookAgentUnknown, details)
}
