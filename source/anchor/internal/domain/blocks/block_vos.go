package blocks

import (
	"slices"
	"strings"
)

type markerStyle struct {
	begin string
	end   string
}

var markerStyles = []markerStyle{
	{begin: "<!-- BEGIN BEACON", end: "<!-- END BEACON -->"},
	{begin: "# BEGIN BEACON", end: "# END BEACON"},
}

type Block struct {
	value string
	style markerStyle
}

func ParseBlock(raw string) (Block, error) {
	value := strings.ReplaceAll(raw, "\r\n", "\n")
	value = strings.TrimSpace(value)
	lines := strings.Split(value, "\n")

	for _, style := range markerStyles {
		if hasMarkers(lines, style) {
			return Block{value: value, style: style}, nil
		}
	}

	return Block{}, invalidBlock(raw)
}

func hasMarkers(lines []string, style markerStyle) bool {
	if len(lines) < 2 {
		return false
	}
	if !strings.HasPrefix(lines[0], style.begin) {
		return false
	}
	if lines[len(lines)-1] != style.end {
		return false
	}

	for _, line := range lines[1:] {
		if strings.HasPrefix(line, style.begin) {
			return false
		}
	}
	return !slices.Contains(lines[:len(lines)-1], style.end)
}

func (b Block) String() string {
	return b.value
}
