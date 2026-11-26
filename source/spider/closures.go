package spider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
)

type closedLine struct {
	Type   string `json:"type"`
	Record struct {
		ID       string          `json:"id"`
		Approved json.RawMessage `json:"approved"`
	} `json:"record"`
}

func closuresOf(change commit.StateChange) ([]Closure, error) {
	known := archiveLines(change.Before)
	var closures []Closure

	for _, text := range archiveLines(change.After) {
		if slices.Contains(known, text) {
			continue
		}

		var line closedLine
		if err := json.Unmarshal([]byte(text), &line); err != nil {
			return nil, kernel.NewRecordFailure(change.Path, 0, err)
		}

		closures = append(closures, Closure{
			Type:     line.Type,
			ID:       line.Record.ID,
			Approved: string(line.Record.Approved),
		})
	}

	return closures, nil
}

func archiveLines(data []byte) []string {
	var lines []string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(nil, 1<<24)

	for scanner.Scan() {
		if text := strings.TrimSpace(scanner.Text()); text != "" {
			lines = append(lines, text)
		}
	}

	return lines
}
