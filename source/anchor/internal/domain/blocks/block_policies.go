package blocks

import "strings"

type RemovalStatus string

const (
	RemovalRemoved RemovalStatus = "removed"
	RemovalAbsent  RemovalStatus = "absent"
	RemovalEdited  RemovalStatus = "edited"
)

type region struct {
	start int
	stop  int
	end   int
}

func Upsert(text string, block Block) (string, error) {
	found, ok, err := locate(text, block)
	if err != nil {
		return "", err
	}

	newline := lineEnding(text)
	rendered := strings.ReplaceAll(block.value, "\n", newline)

	if ok {
		trailing := text[found.stop:found.end]
		return text[:found.start] + rendered + trailing + text[found.end:], nil
	}

	return appendBlock(text, rendered, newline), nil
}

func appendBlock(text, rendered, newline string) string {
	switch {
	case text == "":
		return rendered + newline
	case strings.HasSuffix(text, "\n"):
		return text + newline + rendered + newline
	default:
		return text + newline + newline + rendered + newline
	}
}

func Remove(text string, block Block) (string, RemovalStatus, error) {
	found, ok, err := locate(text, block)
	if err != nil {
		return "", RemovalAbsent, err
	}
	if !ok {
		return text, RemovalAbsent, nil
	}

	current := strings.ReplaceAll(text[found.start:found.stop], "\r\n", "\n")
	if current != block.value {
		return text, RemovalEdited, nil
	}

	before := text[:found.start]
	if found.end == len(text) {
		before = trimSeparator(before)
	}

	return before + text[found.end:], RemovalRemoved, nil
}

func Covers(text string, block Block) bool {
	if _, found, err := locate(text, block); found || err != nil {
		return false
	}

	inner := strings.Split(block.value, "\n")
	inner = inner[1 : len(inner)-1]

	present := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		present[strings.TrimSuffix(line, "\r")] = true
	}

	required := 0
	for _, line := range inner {
		if line == "" {
			continue
		}

		required++
		if !present[line] {
			return false
		}
	}

	return required > 0
}

func trimSeparator(before string) string {
	for _, newline := range []string{"\r\n", "\n"} {
		if strings.HasSuffix(before, newline+newline) {
			return strings.TrimSuffix(before, newline)
		}
	}

	return before
}

func locate(text string, block Block) (region, bool, error) {
	start := -1
	position := 0

	for position <= len(text) {
		line, next := lineAt(text, position)

		if start < 0 && strings.HasPrefix(line, block.style.begin) {
			start = position
		}
		if start >= 0 && line == block.style.end {
			stop := position + len(line)
			return region{start, stop, stop + eolLength(text, stop)}, true, nil
		}
		if next < 0 {
			break
		}

		position = next
	}

	if start >= 0 {
		return region{}, false, FailBlockUnclosed(block.style.end)
	}

	return region{}, false, nil
}

func lineAt(text string, position int) (string, int) {
	rest := text[position:]
	cut := strings.IndexByte(rest, '\n')

	if cut < 0 {
		return strings.TrimSuffix(rest, "\r"), -1
	}

	return strings.TrimSuffix(rest[:cut], "\r"), position + cut + 1
}

func eolLength(text string, position int) int {
	rest := text[position:]

	switch {
	case strings.HasPrefix(rest, "\r\n"):
		return 2
	case strings.HasPrefix(rest, "\n"):
		return 1
	default:
		return 0
	}
}

func lineEnding(text string) string {
	if strings.Contains(text, "\r\n") {
		return "\r\n"
	}

	return "\n"
}
