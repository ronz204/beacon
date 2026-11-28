package record

import (
	"regexp"
)

var (
	packNamePattern  = regexp.MustCompile(`^[a-z]{3,6}$`)
	pieceNamePattern = regexp.MustCompile(`^[a-z]{6}$`)
	stackNamePattern = regexp.MustCompile(`^[a-z]{2,10}$`)
	languagePattern  = regexp.MustCompile(`^[a-z]{2,3}$`)
)

type ConfigLanguage struct {
	Code string
	UI   string
}

type ConfigCommit struct {
	Min int
	Max int
}

type ConfigPacks struct {
	Present  bool
	Enabled  []string
	Skip     []string
	Stacks   []string
	Language ConfigLanguage
	Commit   ConfigCommit
}

func configPacks(f *fields) ConfigPacks {
	section, found := f.mapping("packs")
	if !found {
		return ConfigPacks{}
	}

	packs := ConfigPacks{
		Present: true,
		Enabled: many(section, "enabled", matching(packNamePattern)),
		Skip:    many(section, "skip", matching(pieceNamePattern)),
		Stacks:  many(section, "stacks", matching(stackNamePattern)),
	}

	if language, ok := section.mapping("language"); ok {
		packs.Language = ConfigLanguage{
			Code: one(language, "code", matching(languagePattern)),
			UI:   one(language, "ui", matching(languagePattern)),
		}
	}
	if commit, ok := section.mapping("commit"); ok {
		packs.Commit = configCommit(commit)
	}

	return packs
}

func configCommit(section *fields) ConfigCommit {
	commit := ConfigCommit{
		Min: one(section, "min", parsePositive),
		Max: one(section, "max", parsePositive),
	}

	inverted := commit.Min > 0 && commit.Max > 0
	if inverted && commit.Min > commit.Max {
		line := section.root.Line
		section.fail(line, failAt(RecordPacksInvalid, section.file, line))
	}

	return commit
}

func matching(pattern *regexp.Regexp) func(string) (string, error) {
	return func(raw string) (string, error) {
		if !pattern.MatchString(raw) {
			return "", recordErrors.Fail(
				RecordPacksInvalid,
				map[string]any{"value": raw},
			)
		}

		return raw, nil
	}
}
