package anchor

type InstallOptions struct {
	DryRun   bool
	Hooks    bool
	Reinject bool
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

type InstallView struct {
	DryRun  bool         `json:"dryRun"`
	Changes []ChangeView `json:"changes"`
	Agents  []AgentView  `json:"agents"`
}
