package kernel

import (
	"strings"
	"unicode/utf8"
)

type PathGlob struct {
	value string
}

func ParsePathGlob(raw string) (PathGlob, error) {
	value := strings.TrimSpace(raw)

	if !isPathGlob(value) {
		return PathGlob{}, invalidValue(KernelGlobInvalid, raw)
	}

	return PathGlob{value: value}, nil
}

func (g PathGlob) String() string {
	return g.value
}

func isPathGlob(value string) bool {
	switch {
	case value == "":
		return false
	case !utf8.ValidString(value):
		return false
	case hasControlCharacter(value):
		return false
	case strings.HasPrefix(value, "!"):
		return false
	case strings.ContainsAny(value, `\[]`):
		return false
	case !hasRepoRelativeShape(value):
		return false
	case !hasFlatBraces(value):
		return false
	}

	return hasValidSegments(value)
}

func hasRepoRelativeShape(value string) bool {
	if strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") {
		return false
	}
	if strings.Contains(value, "//") {
		return false
	}

	return len(value) < 3 || value[1:3] != ":/"
}

func hasFlatBraces(value string) bool {
	depth := 0

	for _, char := range value {
		switch char {
		case '{':
			depth++
		case '}':
			depth--
		}

		if depth < 0 || depth > 1 {
			return false
		}
	}

	return depth == 0
}

func hasValidSegments(value string) bool {
	segments := strings.FieldsFunc(value, isSegmentDelimiter)

	for _, segment := range segments {
		if !isValidSegment(segment) {
			return false
		}
	}

	return true
}

func isSegmentDelimiter(char rune) bool {
	return strings.ContainsRune("/{},", char)
}

func isValidSegment(segment string) bool {
	if segment == "." || segment == ".." {
		return false
	}

	return segment == "**" || !strings.Contains(segment, "**")
}
