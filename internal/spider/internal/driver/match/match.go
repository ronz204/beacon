package match

import (
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/ronz204/beacon/internal/kernel"
)

const (
	anySegments    = `(?:.*/)?`
	anyDescendants = `(?:/.*)?`
	anyPath        = `.*`
	anyCharacters  = `[^/]*`
	oneCharacter   = `[^/]`
)

type Matcher struct {
	mutex    sync.Mutex
	patterns map[kernel.PathGlob]*regexp.Regexp
}

func NewMatcher() *Matcher {
	return &Matcher{patterns: make(map[kernel.PathGlob]*regexp.Regexp)}
}

func (m *Matcher) MatchPath(glob kernel.PathGlob, path string) (bool, error) {
	pattern, err := m.compiled(glob)
	if err != nil {
		return false, kernel.FailPathGlob(glob, err)
	}

	normalized := strings.ReplaceAll(path, `\`, "/")

	return pattern.MatchString(normalized), nil
}

func (m *Matcher) compiled(glob kernel.PathGlob) (*regexp.Regexp, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if pattern, found := m.patterns[glob]; found {
		return pattern, nil
	}

	pattern, err := regexp.Compile(globExpression(glob.String()))
	if err != nil {
		return nil, err
	}

	m.patterns[glob] = pattern

	return pattern, nil
}

func globExpression(glob string) string {
	alternatives := expandBraces(glob)
	expressions := make([]string, 0, len(alternatives))

	for _, alternative := range alternatives {
		expressions = append(expressions, plainExpression(alternative))
	}

	return `(?s)\A(?:` + strings.Join(expressions, "|") + `)\z`
}

func expandBraces(glob string) []string {
	opening := strings.IndexByte(glob, '{')
	if opening < 0 {
		return []string{glob}
	}

	closing := opening + strings.IndexByte(glob[opening:], '}')
	prefix := glob[:opening]
	tails := expandBraces(glob[closing+1:])

	var expanded []string
	for _, choice := range strings.Split(glob[opening+1:closing], ",") {
		for _, tail := range tails {
			expanded = append(expanded, prefix+choice+tail)
		}
	}

	return expanded
}

func plainExpression(glob string) string {
	var expression strings.Builder

	for index := 0; index < len(glob); {
		token, width := nextToken(glob, index)
		expression.WriteString(token)
		index += width
	}

	return expression.String()
}

func nextToken(glob string, index int) (string, int) {
	rest := glob[index:]
	startsSegment := index == 0 || glob[index-1] == '/'

	switch {
	case rest == "/**":
		return anyDescendants, 3
	case startsSegment && strings.HasPrefix(rest, "**/"):
		return anySegments, 3
	case startsSegment && rest == "**":
		return anyPath, 2
	case strings.HasPrefix(rest, "**"):
		return anyCharacters, 2
	case rest[0] == '*':
		return anyCharacters, 1
	case rest[0] == '?':
		return oneCharacter, 1
	}

	_, width := utf8.DecodeRuneInString(rest)

	return regexp.QuoteMeta(rest[:width]), width
}
