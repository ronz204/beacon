package spider

import (
	"bytes"
	"context"
	"os"

	"github.com/ronz204/beacon/source/spider/internal/action"
)

type FormatResult struct {
	Changed  []string
	Findings []Finding
}

func (p Project) Format(ctx context.Context) FormatResult {
	changed, failures := p.store.Format(ctx, true)

	for index, file := range changed {
		changed[index] = p.relative(file)
	}

	return FormatResult{
		Changed:  changed,
		Findings: p.relativeFindings(action.FindingsFrom(failures)),
	}
}

const lineEndingsMessage = "the file uses CRLF line endings"

func refineLineEndings(findings []Finding) []Finding {
	for index, finding := range findings {
		if finding.Code != CodeNotCanonical {
			continue
		}

		data, err := os.ReadFile(finding.File)
		if err == nil && bytes.Contains(data, []byte("\r\n")) {
			findings[index].Code = CodeLineEndings
			findings[index].Message = lineEndingsMessage
		}
	}

	return findings
}
