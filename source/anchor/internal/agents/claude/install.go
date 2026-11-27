package claude

import (
	"fmt"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

const (
	importPath   = "CLAUDE.md"
	importBegin  = "<!-- BEGIN BEACON (managed block: do not edit inside) -->"
	importBlock  = importBegin + "\n@AGENTS.md\n<!-- END BEACON -->"
	settingsPath = ".claude/settings.json"

	startMatcher    = "startup|resume"
	reinjectMatcher = "clear|compact"
	editMatcher     = "Edit|Write"
	startTimeout    = 30
	editTimeout     = 10
)

var shellTools = []string{"Bash", "PowerShell"}

func (Agent) Moments() []agents.Moment {
	return []agents.Moment{
		agents.MomentStart,
		agents.MomentCompact,
		agents.MomentEdit,
	}
}

func (Agent) Contribute(
	options agents.Options,
) (agents.Contribution, error) {
	block, err := blocks.ParseBlock(importBlock)
	if err != nil {
		return agents.Contribution{}, err
	}

	allow, err := permissionEntries()
	if err != nil {
		return agents.Contribution{}, err
	}

	contribution := agents.Contribution{
		Blocks: []agents.BlockFile{{
			Path:  importPath,
			Part:  agents.PartImport,
			Block: block,
		}},
		Entries: []agents.EntryFile{{
			Path:    settingsPath,
			Part:    agents.PartPermission,
			Entries: allow,
		}},
	}

	files, err := hookFiles(options)
	if err != nil {
		return agents.Contribution{}, err
	}

	contribution.Entries = append(contribution.Entries, files...)
	contribution.Files, err = pieceFiles(options.Pieces)

	return contribution, err
}

func permissionEntries() ([]blocks.Entry, error) {
	var entries []blocks.Entry

	for _, tool := range shellTools {
		rule := tool + "(beacon *)"

		entry, err := blocks.ParseEntry(
			blocks.EntryAppend, rule, "permissions", "allow",
		)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func hookFiles(options agents.Options) ([]agents.EntryFile, error) {
	if !options.Hooks {
		return nil, nil
	}

	start, err := hookEntry(
		"SessionStart", startMatcher, agents.EventStart, startTimeout,
	)
	if err != nil {
		return nil, err
	}

	files := []agents.EntryFile{{
		Path:    settingsPath,
		Part:    agents.PartHooks,
		Entries: []blocks.Entry{start},
	}}
	if !options.Reinject {
		return files, nil
	}

	reinject, err := hookEntry(
		"SessionStart", reinjectMatcher, agents.EventStart, startTimeout,
	)
	if err != nil {
		return nil, err
	}

	edit, err := hookEntry(
		"PreToolUse", editMatcher, agents.EventEdit, editTimeout,
	)
	if err != nil {
		return nil, err
	}

	return append(files, agents.EntryFile{
		Path:    settingsPath,
		Part:    agents.PartReinject,
		Entries: []blocks.Entry{reinject, edit},
	}), nil
}

func hookEntry(
	claudeEvent, matcher string,
	event agents.Event,
	timeout int,
) (blocks.Entry, error) {
	command := fmt.Sprintf("beacon hook %s --agent %s", event, name)

	return blocks.ParseEntry(
		blocks.EntryAppend,
		map[string]any{
			"matcher": matcher,
			"hooks": []any{map[string]any{
				"type":    "command",
				"command": command,
				"timeout": timeout,
			}},
		},
		"hooks",
		claudeEvent,
	)
}
