package bench

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

var (
	shellTools = []string{"Bash", "PowerShell"}
	editTools  = []string{"Edit", "Write", "MultiEdit"}

	beaconCall = regexp.MustCompile(
		`(?:^|[\s;&|(])beacon\s+([a-z]+)([^;&|]*)`,
	)
)

const (
	blockedByBeacon     = "approval_required"
	blockedByBeaconText = "needs a yes"
	blockedByPermission = "requires approval"
	startHookPrefix     = "SessionStart"
	editHookPrefix      = "PreToolUse:"
	stateFolder         = "/.beacon/"
	maxLine             = 4 << 20
)

type streamLine struct {
	Type       string  `json:"type"`
	Subtype    string  `json:"subtype"`
	HookName   string  `json:"hook_name"`
	Stdout     string  `json:"stdout"`
	NumTurns   int     `json:"num_turns"`
	DurationMS int     `json:"duration_ms"`
	CostUSD    float64 `json:"total_cost_usd"`
	Message    struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type contentBlock struct {
	Type    string          `json:"type"`
	Name    string          `json:"name"`
	Input   json.RawMessage `json:"input"`
	Content json.RawMessage `json:"content"`
}

type toolInput struct {
	Command  string `json:"command"`
	FilePath string `json:"file_path"`
}

type reader struct {
	session     Session
	contextArgs []string
	toolSeen    bool
	lastEdit    map[string]int
}

func ParseTranscript(source io.Reader) (Session, error) {
	state := reader{
		session:  Session{Commands: map[string]int{}},
		lastEdit: map[string]int{},
	}

	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLine)

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		var line streamLine
		if err := json.Unmarshal([]byte(text), &line); err != nil {
			return Session{}, err
		}

		state.read(line)
	}

	return state.session, scanner.Err()
}

func (r *reader) read(line streamLine) {
	switch {
	case line.Type == "system" && line.Subtype == "hook_response":
		r.readHook(line)
	case line.Type == "assistant":
		r.readAssistant(line)
	case line.Type == "user":
		r.readResults(line)
	case line.Type == "result":
		r.session.Turns = line.NumTurns
		r.session.DurationMS = line.DurationMS
		r.session.CostUSD = line.CostUSD
	}
}

func (r *reader) readHook(line streamLine) {
	context := additionalContext(line.Stdout)
	if context == "" {
		return
	}

	lines := strings.Count(context, "\n") + 1
	r.session.Injections++
	r.session.InjectedLines += lines

	if strings.HasPrefix(line.HookName, startHookPrefix) && !r.toolSeen {
		r.session.StartContext = true
		r.session.StartLines += lines
	}

	tool, isTool := strings.CutPrefix(line.HookName, editHookPrefix)
	if index, waiting := r.lastEdit[tool]; isTool && waiting {
		r.session.Edits[index].HookContext = true
		delete(r.lastEdit, tool)
	}
}

func additionalContext(stdout string) string {
	var reply struct {
		Specific struct {
			Context string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}

	err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &reply)
	if err != nil {
		return ""
	}

	return reply.Specific.Context
}

func (r *reader) readAssistant(line streamLine) {
	for _, block := range blocksOf(line.Message.Content) {
		if block.Type != "tool_use" {
			continue
		}

		var input toolInput
		_ = json.Unmarshal(block.Input, &input)

		r.toolSeen = true

		switch {
		case contains(shellTools, block.Name):
			r.readCommand(input.Command)
		case contains(editTools, block.Name):
			r.readEdit(block.Name, input.FilePath)
		}
	}
}

func (r *reader) readCommand(command string) {
	for _, match := range beaconCall.FindAllStringSubmatch(command, -1) {
		name := match[1]
		r.session.Commands[name]++

		switch name {
		case "brief":
			if len(r.session.Edits) == 0 {
				r.session.AgentBrief = true
			}
		case "context":
			r.contextArgs = append(r.contextArgs, strings.Fields(match[2])...)
		}
	}
}

func (r *reader) readEdit(tool, file string) {
	slashed := strings.ReplaceAll(file, "\\", "/")

	if strings.Contains(slashed, stateFolder) {
		r.session.Tampering++
	}

	for _, edit := range r.session.Edits {
		if edit.Path == slashed {
			return
		}
	}

	r.session.Edits = append(r.session.Edits, Edit{
		Path:         slashed,
		AgentContext: r.mentioned(slashed),
	})
	r.lastEdit[tool] = len(r.session.Edits) - 1
}

func (r *reader) mentioned(file string) bool {
	for _, argument := range r.contextArgs {
		token := strings.Trim(argument, "\"'")
		token = strings.TrimPrefix(strings.ReplaceAll(token, "\\", "/"), "./")

		if token == "" || strings.HasPrefix(token, "-") {
			continue
		}
		if file == token || strings.HasSuffix(file, "/"+token) {
			return true
		}
	}

	return false
}

func (r *reader) readResults(line streamLine) {
	for _, block := range blocksOf(line.Message.Content) {
		if block.Type != "tool_result" {
			continue
		}

		text := resultText(block.Content)

		switch {
		case strings.Contains(text, blockedByBeacon),
			strings.Contains(text, blockedByBeaconText):
			r.session.BlockedByBeacon++
		case strings.Contains(text, blockedByPermission):
			r.session.BlockedByPermission++
		}
	}
}

func blocksOf(raw json.RawMessage) []contentBlock {
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil
	}

	return blocks
}

func resultText(raw json.RawMessage) string {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}

	var parts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return string(raw)
	}

	var joined strings.Builder
	for _, part := range parts {
		joined.WriteString(part.Text)
	}

	return joined.String()
}

func contains(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}

	return false
}
