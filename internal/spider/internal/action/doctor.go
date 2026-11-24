package action

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
)

const (
	CodeNotCanonical = "record_not_canonical"
	CodeGuideTooLong = "guide_too_long"

	notCanonicalMessage = "run beacon fmt to format this record"
	guideTooLongMessage = "split this guide into smaller guides"
)

type FindingSeverity string

const (
	FindingError   FindingSeverity = "error"
	FindingWarning FindingSeverity = "warning"
)

type Finding struct {
	Severity FindingSeverity `json:"severity"`
	Code     string          `json:"code"`
	Message  string          `json:"message"`
	File     string          `json:"file,omitempty"`
	Line     int             `json:"line,omitempty"`
	Cause    string          `json:"cause,omitempty"`
	Details  map[string]any  `json:"details,omitempty"`
}

type DoctorInput struct {
	Repositories   Repositories
	Formatter      kernel.RecordFormatter
	GuideLimit     int
	ConfigFailures []error
}

func Doctor(ctx context.Context, input DoctorInput) []Finding {
	snapshot := Load(ctx, input.Repositories)
	failures := slices.Concat(input.ConfigFailures, snapshot.Failures)

	if len(snapshot.Failures) == 0 {
		_, graphFailures := Validate(snapshot)
		failures = append(failures, graphFailures...)
	}

	changed, formatFailures := input.Formatter.Format(ctx, false)
	failures = append(failures, formatFailures...)

	findings := FindingsFrom(failures)
	findings = append(findings, canonicalFindings(changed)...)
	findings = append(
		findings,
		guideFindings(snapshot.Guides, input.GuideLimit)...,
	)

	return sortedFindings(uniqueFindings(findings))
}

func FindingsFrom(failures []error) []Finding {
	findings := make([]Finding, 0, len(failures))

	for _, failure := range failures {
		view := kernel.DescribeFailure(failure)
		findings = append(findings, Finding{
			Severity: FindingError,
			Code:     view.Code,
			Message:  view.Message,
			File:     view.File,
			Line:     view.Line,
			Cause:    view.Cause,
			Details:  view.Details,
		})
	}

	return findings
}

func canonicalFindings(files []string) []Finding {
	findings := make([]Finding, 0, len(files))

	for _, file := range files {
		findings = append(findings, Finding{
			Severity: FindingError,
			Code:     CodeNotCanonical,
			Message:  notCanonicalMessage,
			File:     file,
		})
	}

	return findings
}

func guideFindings(guides []memory.Guide, limit int) []Finding {
	if limit <= 0 {
		return nil
	}

	var findings []Finding
	for _, guide := range guides {
		if guide.Lines() <= limit {
			continue
		}

		findings = append(findings, Finding{
			Severity: FindingWarning,
			Code:     CodeGuideTooLong,
			Message:  guideTooLongMessage,
			Details: map[string]any{
				"guide": guide.ID().String(),
				"lines": guide.Lines(),
				"limit": limit,
			},
		})
	}

	return findings
}

func uniqueFindings(findings []Finding) []Finding {
	seen := make(map[string]bool, len(findings))
	unique := make([]Finding, 0, len(findings))

	for _, finding := range findings {
		key := fmt.Sprint(
			finding.File,
			finding.Line,
			finding.Code,
			finding.Details,
		)

		if !seen[key] {
			seen[key] = true
			unique = append(unique, finding)
		}
	}

	return unique
}

func sortedFindings(findings []Finding) []Finding {
	slices.SortStableFunc(findings, func(a, b Finding) int {
		return cmp.Or(
			cmp.Compare(a.Severity, b.Severity),
			cmp.Compare(a.File, b.File),
			cmp.Compare(a.Line, b.Line),
			cmp.Compare(a.Code, b.Code),
		)
	})

	return findings
}
