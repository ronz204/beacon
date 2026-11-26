package output

import (
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/spider"
)

func SignalLines(signals []spider.Signal) []string {
	lines := make([]string, 0, len(signals))

	for _, signal := range signals {
		lines = append(lines, signalLine(signal))
	}

	return lines
}

func signalLine(signal spider.Signal) string {
	parts := []string{signal.Kind, signal.Record}

	if signal.Relation != "" {
		parts = append(parts, signal.Relation, signal.Related)
	}

	return fmt.Sprintf("signal: %s", strings.Join(parts, " "))
}

func signalList(signals []spider.Signal) []spider.Signal {
	if signals == nil {
		return []spider.Signal{}
	}

	return signals
}
