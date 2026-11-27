package snares

import (
	"github.com/ronz204/beacon/source/snares/internal/action"
	"github.com/ronz204/beacon/source/weaver"
)

type BlockView struct {
	Kind    string   `json:"kind"`
	Message string   `json:"message"`
	Hint    string   `json:"hint"`
	Items   []string `json:"items"`
}

const (
	ScopeAll     = action.ScopeAll
	ScopeTouched = action.ScopeTouched
)

type CIInput struct {
	Base  string
	Scope string
}

type CIView struct {
	Revision string          `json:"revision"`
	Scope    string          `json:"scope"`
	Risk     weaver.RiskView `json:"risk"`
	Blocks   []BlockView     `json:"blocks"`
	Warnings []string        `json:"warnings"`
	Checked  []string        `json:"checked"`
	Skipped  []string        `json:"skipped"`
	Comment  string          `json:"-"`
}

type PreCommitView struct {
	Blocks   []BlockView `json:"blocks"`
	Warnings []string    `json:"warnings"`
	Checked  []string    `json:"checked"`
}
