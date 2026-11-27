package bench

type Edit struct {
	Path         string
	AgentContext bool
	HookContext  bool
}

func (e Edit) HadContext() bool {
	return e.AgentContext || e.HookContext
}

type Session struct {
	StartContext        bool
	StartLines          int
	AgentBrief          bool
	Edits               []Edit
	Commands            map[string]int
	Injections          int
	InjectedLines       int
	Tampering           int
	BlockedByBeacon     int
	BlockedByPermission int
	Turns               int
	DurationMS          int
	CostUSD             float64
	Outcome             *Outcome
}

type Outcome struct {
	Resolved bool
	Pending  int
}

func (s Session) BriefAtStart() bool {
	return s.StartContext || s.AgentBrief
}

func (s Session) EditsWithContext() int {
	count := 0
	for _, edit := range s.Edits {
		if edit.HadContext() {
			count++
		}
	}

	return count
}
