package spider_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/spider"
)

func crlfCapability(t *testing.T, root string) {
	t.Helper()

	text := readFile(t, root, capabilityFile)
	writeFile(t, root, capabilityFile, strings.ReplaceAll(text, "\n", "\r\n"))
}

func TestDoctorNamesCRLFLineEndings(t *testing.T) {
	project, root := newProject(t)
	crlfCapability(t, root)

	findings := project.Doctor(context.Background())

	var codes []string
	for _, finding := range findings {
		codes = append(codes, finding.Code)
	}
	if len(codes) != 1 || codes[0] != spider.CodeLineEndings {
		t.Fatalf("codes = %v", codes)
	}
}

func TestFormatFixesCRLFLineEndings(t *testing.T) {
	project, root := newProject(t)
	crlfCapability(t, root)

	result := project.Format(context.Background())

	if len(result.Changed) != 1 || len(result.Findings) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if strings.Contains(readFile(t, root, capabilityFile), "\r") {
		t.Fatal("fmt must write LF")
	}
	if findings := project.Doctor(context.Background()); len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestDoctorKeepsTheNotCanonicalCodeForOtherDifferences(t *testing.T) {
	project, root := newProject(t)
	text := readFile(t, root, capabilityFile)
	writeFile(t, root, capabilityFile, strings.Replace(
		text,
		"paths: [src/auth/**, tests/auth/**]",
		"paths: [ src/auth/**,   tests/auth/** ]",
		1,
	))

	findings := project.Doctor(context.Background())

	if len(findings) != 1 || findings[0].Code != spider.CodeNotCanonical {
		t.Fatalf("findings = %+v", findings)
	}
}
