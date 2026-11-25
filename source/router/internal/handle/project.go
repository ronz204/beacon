package handle

import (
	"fmt"
	"os"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
)

const (
	createProjectHint  = "create .beacon/config.yaml with schema: 1"
	existingFolderHint = "run beacon from an existing folder"
	fixRecordsHint     = "fix the reported records, then run beacon doctor"
	formatHint         = "run beacon fmt"
	failureFallback    = "command_failed"
)

func locateProject() (spider.Project, *output.Failure) {
	directory, err := os.Getwd()
	if err != nil {
		return spider.Project{}, projectFailure(err)
	}

	project, err := spider.Locate(directory)
	if err != nil {
		return spider.Project{}, projectFailure(err)
	}

	return project, nil
}

func projectFailure(err error) *output.Failure {
	view := kernel.DescribeFailure(err)

	if view.Code == "" {
		return &output.Failure{
			Code:    failureFallback,
			Message: view.Cause,
			Hint:    existingFolderHint,
		}
	}

	return &output.Failure{
		Code:    view.Code,
		Message: view.Message,
		Hint:    createProjectHint,
	}
}

func errorFindings(findings []spider.Finding) []spider.Finding {
	failed := make([]spider.Finding, 0, len(findings))

	for _, finding := range findings {
		if finding.Severity == spider.FindingError {
			failed = append(failed, finding)
		}
	}

	return failed
}

func findingsHint(findings []spider.Finding) string {
	for _, finding := range errorFindings(findings) {
		if finding.Code != spider.CodeNotCanonical {
			return fixRecordsHint
		}
	}

	return formatHint
}

func findingLines(findings []spider.Finding) []string {
	lines := make([]string, 0, len(findings))

	for _, finding := range findings {
		lines = append(lines, findingLine(finding))
	}

	return lines
}

func findingLine(finding spider.Finding) string {
	line := fmt.Sprintf(
		"%s %s %s: %s",
		finding.Severity,
		findingLocation(finding),
		finding.Code,
		finding.Message,
	)

	if finding.Cause != "" {
		line += " (" + finding.Cause + ")"
	}

	return line
}

func findingLocation(finding spider.Finding) string {
	details := finding.Details

	switch {
	case finding.File != "" && finding.Line > 0:
		return fmt.Sprintf("%s:%d", finding.File, finding.Line)
	case finding.File != "":
		return finding.File
	case details["guide"] != nil:
		return fmt.Sprint(details["guide"])
	case details["relation"] != nil:
		return fmt.Sprintf(
			"%v %v %v",
			details["source"],
			details["relation"],
			details["target"],
		)
	case details["source"] != nil:
		return fmt.Sprintf("%v -> %v", details["source"], details["target"])
	}

	return "project"
}
