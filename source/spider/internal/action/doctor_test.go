package action_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

const guideHeader = "---\ntitle: Sample\nkind: domain\n---\n"

func goalText(status string) string {
	return "id: bc-0001\ncreated: 2026-11-22\nauthor: \"human:jane\"\n" +
		"statement: Sample\nsuccess: Sample\nstatus: " + status + "\n"
}

func runDoctor(root string, limit int, failures ...error) []action.Finding {
	return action.Doctor(context.Background(), action.DoctorInput{
		Repositories:   repositories(root),
		Formatter:      record.NewStore(root),
		GuideLimit:     limit,
		ConfigFailures: failures,
	})
}

func findingCodes(findings []action.Finding) []string {
	codes := make([]string, 0, len(findings))

	for _, finding := range findings {
		codes = append(codes, finding.Code)
	}

	return codes
}

func TestDoctorFindsNothingInValidProject(t *testing.T) {
	root := filepath.Join("..", "driver", "record", "testdata", "valid")

	findings := runDoctor(root, 150)

	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestDoctorReportsAnInvalidFileOnce(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"goals/bc-0001-sample.yaml": goalText("paused"),
	})

	findings := runDoctor(root, 150)

	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}

	finding := findings[0]
	if finding.Severity != action.FindingError {
		t.Fatalf("severity = %q", finding.Severity)
	}
	if finding.Code != "goal_status_invalid" || finding.Line != 6 {
		t.Fatalf("finding = %+v", finding)
	}

	file := filepath.ToSlash(finding.File)
	if !strings.HasSuffix(file, "goals/bc-0001-sample.yaml") {
		t.Fatalf("file = %q", finding.File)
	}
}

func TestDoctorSkipsGraphWhenRecordsFailToLoad(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"works/bc-91cd-sample.yaml": "id: bc-91cd\nstatus: paused\n",
		"tasks/bc-91cd.1-sample.yaml": taskText(
			"bc-91cd.1",
			"bc-91cd",
			"open",
			"",
		),
	})

	findings := runDoctor(root, 150)

	for _, code := range findingCodes(findings) {
		if code == "graph_target_missing" {
			t.Fatalf("cascading graph failure: %+v", findings)
		}
	}
}

func TestDoctorReportsGraphFailures(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"tasks/bc-91cd.1-sample.yaml": taskText(
			"bc-91cd.1",
			"bc-91cd",
			"open",
			"",
		),
	})

	findings := runDoctor(root, 150)

	if len(findings) != 1 || findings[0].Code != "graph_target_missing" {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].Details["target"] != "bc-91cd" {
		t.Fatalf("details = %v", findings[0].Details)
	}
}

func TestDoctorReportsRecordsOutOfFormat(t *testing.T) {
	fields := strings.Replace(goalText("x"), "status: x\n", "", 1)
	text := "status: active\n" + fields
	root := writeRecords(t, map[string]string{
		"goals/bc-0001-sample.yaml": text,
	})

	findings := runDoctor(root, 150)

	if len(findings) != 1 || findings[0].Code != action.CodeNotCanonical {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].Severity != action.FindingError {
		t.Fatalf("severity = %q", findings[0].Severity)
	}
}

func TestDoctorWarnsAboutLongGuides(t *testing.T) {
	body := strings.Repeat("line\n", 4)
	root := writeRecords(t, map[string]string{
		"guides/bc-0a9b-sample.md": guideHeader + body,
	})

	if findings := runDoctor(root, 4); len(findings) != 0 {
		t.Fatalf("guide at the limit was reported: %+v", findings)
	}

	findings := runDoctor(root, 3)

	if len(findings) != 1 || findings[0].Code != action.CodeGuideTooLong {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].Severity != action.FindingWarning {
		t.Fatalf("severity = %q", findings[0].Severity)
	}
	if findings[0].Details["lines"] != 4 || findings[0].Details["limit"] != 3 {
		t.Fatalf("details = %v", findings[0].Details)
	}
}

func TestDoctorIncludesConfigFailures(t *testing.T) {
	root := writeRecords(t, nil)
	_, cause := kernel.ParseCalendarDate("nope")
	failure := kernel.NewRecordFailure("config.yaml", 2, cause)

	findings := runDoctor(root, 0, failure)

	if len(findings) != 1 || findings[0].File != "config.yaml" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestDoctorSortsErrorsBeforeWarnings(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"guides/bc-0a9b-sample.md":  guideHeader + "a\nb\n",
		"goals/bc-0001-sample.yaml": goalText("paused"),
	})

	findings := runDoctor(root, 1)

	if len(findings) != 2 {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].Severity != action.FindingError {
		t.Fatalf("first severity = %q", findings[0].Severity)
	}
	if findings[1].Severity != action.FindingWarning {
		t.Fatalf("second severity = %q", findings[1].Severity)
	}
}
