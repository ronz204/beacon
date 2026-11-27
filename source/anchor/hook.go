package anchor

import (
	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/agents/claude"
)

func hooks() action.Hooks {
	return action.Hooks{Agents: []agents.Hooker{claude.Agent{}}}
}

func ParseHook(agent, event string, stdin []byte) (HookInput, error) {
	input, err := hooks().Parse(agent, event, stdin)
	if err != nil {
		return HookInput{}, err
	}

	return HookInput(input), nil
}

func RenderHook(agent, event, text string) ([]byte, error) {
	return hooks().Render(agent, event, text)
}
