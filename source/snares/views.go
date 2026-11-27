package snares

type BlockView struct {
	Kind    string   `json:"kind"`
	Message string   `json:"message"`
	Hint    string   `json:"hint"`
	Items   []string `json:"items"`
}

type PreCommitView struct {
	Blocks   []BlockView `json:"blocks"`
	Warnings []string    `json:"warnings"`
	Checked  []string    `json:"checked"`
}
