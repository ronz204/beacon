package claude

import (
	"bytes"
	"encoding/json"
	"slices"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
)

const (
	name         = "claude-code"
	startEvent   = "SessionStart"
	editEvent    = "PreToolUse"
	toolEdit     = "Edit"
	toolWrite    = "Write"
	sessionBlank = "the hook input has no session id"
)

var editTools = []string{toolEdit, toolWrite}

type Agent struct{}

type hookPayload struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

type hookOutput struct {
	Specific hookSpecific `json:"hookSpecificOutput"`
}

type hookSpecific struct {
	EventName string `json:"hookEventName"`
	Context   string `json:"additionalContext"`
}

func (Agent) Name() string {
	return name
}

func (Agent) ParseHook(
	event agents.Event,
	stdin []byte,
) (agents.HookInput, error) {
	var payload hookPayload
	if err := json.Unmarshal(stdin, &payload); err != nil {
		return agents.HookInput{}, err
	}
	if payload.SessionID == "" {
		return agents.HookInput{}, errBlankSession
	}

	input := agents.HookInput{
		Session: payload.SessionID,
		Cwd:     payload.Cwd,
		Path:    payload.ToolInput.FilePath,
	}
	if event == agents.EventEdit {
		input.Skip = !isEdit(payload)
	}

	return input, nil
}

func isEdit(payload hookPayload) bool {
	if payload.ToolInput.FilePath == "" {
		return false
	}

	return slices.Contains(editTools, payload.ToolName)
}

func (Agent) RenderHook(event agents.Event, text string) ([]byte, error) {
	if text == "" {
		return nil, nil
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(hookOutput{Specific: hookSpecific{
		EventName: eventName(event),
		Context:   text,
	}})

	return buffer.Bytes(), err
}

func eventName(event agents.Event) string {
	if event == agents.EventEdit {
		return editEvent
	}

	return startEvent
}
