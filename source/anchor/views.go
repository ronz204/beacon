package anchor

type PackBinding struct {
	Language  string
	UI        string
	CommitMin int
	CommitMax int
	Stacks    []string
}

type InstallOptions struct {
	Preview    bool
	Hooks      bool
	Reinject   bool
	Precommit  bool
	CI         string
	Packs      []string
	SkipPieces []string
	Binding    PackBinding
	Reconcile  bool
}

type ChangeView struct {
	Path   string   `json:"path"`
	Action string   `json:"action"`
	Parts  []string `json:"parts"`
	Before string   `json:"before,omitempty"`
	After  string   `json:"after"`
}

type AgentView struct {
	Name    string   `json:"name"`
	Covers  []string `json:"covers"`
	Missing []string `json:"missing"`
}

type KeptView struct {
	Path string `json:"path"`
	Part string `json:"part"`
}

type InstallView struct {
	Preview     bool         `json:"preview"`
	Uninstall   bool         `json:"uninstall"`
	Initialized bool         `json:"initialized"`
	Changes     []ChangeView `json:"changes"`
	Agents      []AgentView  `json:"agents"`
	Skipped     []string     `json:"skipped"`
	Kept        []KeptView   `json:"kept"`
	Review      []string     `json:"review"`
}

type HookInput struct {
	Session string
	Cwd     string
	Path    string
	Skip    bool
}
