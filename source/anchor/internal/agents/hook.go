package agents

type Event string

const (
	EventStart Event = "start"
	EventEdit  Event = "edit"
)

func Events() []Event {
	return []Event{EventStart, EventEdit}
}

type HookInput struct {
	Session string
	Cwd     string
	Path    string
	Skip    bool
}

type Hooker interface {
	Name() string
	ParseHook(event Event, stdin []byte) (HookInput, error)
	RenderHook(event Event, text string) ([]byte, error)
}
