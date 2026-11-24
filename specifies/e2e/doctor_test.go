package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/specifies/utils"
)

const brokenGoal = ".beacon/records/goals/bc-0001-pagos-sin-friccion.yaml"

type finding struct {
	Severity string         `json:"severity"`
	Code     string         `json:"code"`
	File     string         `json:"file"`
	Line     int            `json:"line"`
	Details  map[string]any `json:"details"`
}

type failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details struct {
		Findings []finding `json:"findings"`
		Changed  []string  `json:"changed"`
	} `json:"details"`
}

type envelope struct {
	OK      bool            `json:"ok"`
	Data    json.RawMessage `json:"data"`
	Signals []any           `json:"signals"`
	Hint    *string         `json:"hint"`
	Error   *failure        `json:"error"`
}

func TestMain(m *testing.M) {
	os.Exit(utils.RunSuite(m))
}

func parseEnvelope(t *testing.T, text string) envelope {
	t.Helper()

	var parsed envelope
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid json %q: %v", text, err)
	}
	if parsed.Signals == nil {
		t.Fatalf("signals must be a list: %s", text)
	}

	return parsed
}

func breakGoal(t *testing.T, root string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(brokenGoal))
	text := strings.Replace(
		utils.ReadFile(t, path),
		"status: active",
		"status: paused",
		1,
	)
	utils.WriteFile(t, root, brokenGoal, text)
}

func TestDoctorPassesOnValidProject(t *testing.T) {
	root := utils.NewWorkspace(t)

	result := utils.Beacon(t, root, "doctor")

	if result.Code != 0 {
		t.Fatalf("code = %d, stderr = %s", result.Code, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "no problems") {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}

func TestDoctorJSONUsesTheEnvelope(t *testing.T) {
	root := utils.NewWorkspace(t)

	result := utils.Beacon(t, root, "doctor", "--json")
	parsed := parseEnvelope(t, result.Stdout)

	if result.Code != 0 || !parsed.OK {
		t.Fatalf("result = %+v", result)
	}
	if parsed.Hint != nil || parsed.Error != nil {
		t.Fatalf("envelope = %s", result.Stdout)
	}
}

func TestDoctorFindsTheProjectFromASubfolder(t *testing.T) {
	root := utils.NewWorkspace(t)
	nested := filepath.Join(root, "src", "auth")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	result := utils.Beacon(t, nested, "doctor")

	if result.Code != 0 {
		t.Fatalf("code = %d, stderr = %s", result.Code, result.Stderr)
	}
}

func TestDoctorReportsInvalidRecord(t *testing.T) {
	root := utils.NewWorkspace(t)
	breakGoal(t, root)

	result := utils.Beacon(t, root, "doctor", "--json")
	parsed := parseEnvelope(t, result.Stdout)

	if result.Code != 1 || parsed.OK {
		t.Fatalf("result = %+v", result)
	}
	if string(parsed.Data) != "null" {
		t.Fatalf("data = %s", parsed.Data)
	}
	if parsed.Hint == nil {
		t.Fatal("a failure must carry a hint")
	}

	findings := parsed.Error.Details.Findings
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].Code != "goal_status_invalid" {
		t.Fatalf("code = %q", findings[0].Code)
	}
	if findings[0].File != brokenGoal || findings[0].Line != 8 {
		t.Fatalf("location = %s:%d", findings[0].File, findings[0].Line)
	}
}

func TestDoctorTextNamesFileAndLine(t *testing.T) {
	root := utils.NewWorkspace(t)
	breakGoal(t, root)

	result := utils.Beacon(t, root, "doctor")

	if result.Code != 1 {
		t.Fatalf("code = %d", result.Code)
	}
	if !strings.Contains(result.Stderr, brokenGoal+":8") {
		t.Fatalf("stderr = %q", result.Stderr)
	}
}

func TestDoctorReportsBrokenRelation(t *testing.T) {
	root := utils.NewWorkspace(t)
	path := ".beacon/records/works/bc-91cd-migrar-a-postgres.yaml"
	absolute := filepath.Join(root, filepath.FromSlash(path))
	text := strings.Replace(
		utils.ReadFile(t, absolute),
		"serves: bc-0001",
		"serves: bc-ffff",
		1,
	)
	utils.WriteFile(t, root, path, text)

	result := utils.Beacon(t, root, "doctor", "--json")
	parsed := parseEnvelope(t, result.Stdout)

	if result.Code != 1 {
		t.Fatalf("code = %d", result.Code)
	}
	if parsed.Error.Details.Findings[0].Code != "graph_target_missing" {
		t.Fatalf("findings = %+v", parsed.Error.Details.Findings)
	}
}

func TestDoctorRejectsNewerSchema(t *testing.T) {
	root := utils.NewWorkspace(t)
	utils.WriteFile(t, root, ".beacon/config.yaml", "schema: 2\n")

	result := utils.Beacon(t, root, "doctor", "--json")
	parsed := parseEnvelope(t, result.Stdout)

	if result.Code != 1 {
		t.Fatalf("code = %d", result.Code)
	}
	if parsed.Error.Details.Findings[0].Code != "record_schema_newer" {
		t.Fatalf("findings = %+v", parsed.Error.Details.Findings)
	}
}

func TestDoctorWarnsAboutLongGuideWithoutFailing(t *testing.T) {
	root := utils.NewWorkspace(t)
	utils.WriteFile(
		t,
		root,
		".beacon/config.yaml",
		"schema: 1\nlimits:\n  guide: 1\n",
	)

	result := utils.Beacon(t, root, "doctor", "--json")
	parsed := parseEnvelope(t, result.Stdout)

	if result.Code != 0 || !parsed.OK {
		t.Fatalf("result = %+v", result)
	}
	if !strings.Contains(string(parsed.Data), "guide_too_long") {
		t.Fatalf("data = %s", parsed.Data)
	}
}

func TestCommandsFailOutsideAProject(t *testing.T) {
	for _, command := range []string{"doctor", "fmt"} {
		t.Run(command, func(t *testing.T) {
			result := utils.Beacon(t, t.TempDir(), command, "--json")
			parsed := parseEnvelope(t, result.Stdout)

			if result.Code != 1 || parsed.Error == nil {
				t.Fatalf("result = %+v", result)
			}
			if parsed.Error.Code != "project_missing" {
				t.Fatalf("code = %q", parsed.Error.Code)
			}
		})
	}
}

func TestUsageErrorsExitWithOne(t *testing.T) {
	cases := [][]string{
		{"unknown", "--json"},
		{"doctor", "extra", "--json"},
		{"doctor", "--nope", "--json"},
	}

	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			result := utils.Beacon(t, utils.NewWorkspace(t), args...)
			parsed := parseEnvelope(t, result.Stdout)

			if result.Code != 1 || parsed.Error == nil {
				t.Fatalf("result = %+v", result)
			}
			if parsed.Error.Code != "usage_invalid" {
				t.Fatalf("code = %q", parsed.Error.Code)
			}
		})
	}
}
