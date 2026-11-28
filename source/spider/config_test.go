package spider_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/spider"
)

const configFile = ".beacon/config.yaml"

func configFields(t *testing.T, project spider.Project) map[string]any {
	t.Helper()

	stored, err := project.Config()
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]any
	if err := json.Unmarshal(stored.Data, &fields); err != nil {
		t.Fatal(err)
	}

	return fields
}

func TestRunnersExposeRunAndBaseline(t *testing.T) {
	project, _ := newProject(t)

	runners, findings := project.Runners()

	requireNoFindings(t, findings)
	if len(runners) != 2 || runners[1].Name != "visual" {
		t.Fatalf("runners = %+v", runners)
	}
	if len(runners[1].Baseline) != 1 || runners[0].Run[0] != "pnpm" {
		t.Fatalf("runners = %+v", runners)
	}
}

func TestUnchangedConfigPayloadKeepsEveryByte(t *testing.T) {
	project, root := newProject(t)
	before := readFile(t, root, configFile)

	payload := payloadOf(t, "config", configFields(t, project))
	prepared := save(t, project, payload)

	if !prepared.Unchanged || readFile(t, root, configFile) != before {
		t.Fatalf("config changed:\n%s", readFile(t, root, configFile))
	}
}

func TestConfigRunnerChangeIsStamped(t *testing.T) {
	project, root := newProject(t)
	fields := configFields(t, project)
	runners := fields["runners"].(map[string]any)
	runners["test-go"] = map[string]any{"run": []string{"go", "test", "{pkg}"}}

	prepared, err := project.Prepare(
		context.Background(),
		payloadOf(t, "config", fields),
		author,
	)
	if err != nil {
		t.Fatal(err)
	}
	stamps := map[string]spider.Approval{
		"config": {By: "human:jane", At: "2026-11-25"},
	}
	if err := project.Save(context.Background(), prepared, stamps); err != nil {
		t.Fatal(err)
	}

	text := readFile(t, root, configFile)
	if !strings.Contains(text, "test-go") ||
		!strings.Contains(text, "human:jane") {
		t.Fatalf("config:\n%s", text)
	}
	if runners, findings := project.Runners(); len(findings) != 0 ||
		len(runners) != 3 {
		t.Fatalf("runners = %+v, findings = %+v", runners, findings)
	}
}

func TestConfigRejectsAnEmptyRunner(t *testing.T) {
	project, _ := newProject(t)
	fields := configFields(t, project)
	fields["runners"].(map[string]any)["broken"] = map[string]any{
		"run": []string{},
	}

	_, err := project.Prepare(
		context.Background(),
		payloadOf(t, "config", fields),
		author,
	)

	requireCode(t, err, string(spider.SavePayloadInvalid))
}

func TestStateChangesSeeConfigAndBaselines(t *testing.T) {
	project, root := newProject(t)
	writeSource(t, root, "tests/auth/login-snapshots/home.png", "png")
	writeFile(t, root, configFile, strings.Replace(
		readFile(t, root, configFile),
		"brief: 60",
		"brief: 40",
		1,
	))

	changes, err := project.StateChanges(context.Background(), "HEAD")
	if err != nil {
		t.Fatal(err)
	}

	kinds := map[string]string{}
	for _, change := range changes {
		kind := change.Kind
		if change.Baseline {
			kind = "baseline"
		} else if change.After != nil {
			kind = change.After.Type
		}
		kinds[change.Path] = kind
		if change.Baseline && len(change.Runners) != 1 {
			t.Fatalf("runners = %v", change.Runners)
		}
	}
	if kinds[configFile] != "config" ||
		kinds["tests/auth/login-snapshots/home.png"] != "baseline" {
		t.Fatalf("changes = %v", kinds)
	}
}

func TestPacksAreAbsentUntilTheSectionIsWritten(t *testing.T) {
	project, _ := newProject(t)

	packs, findings := project.Packs()

	requireNoFindings(t, findings)
	if packs.Present {
		t.Errorf("packs = %+v", packs)
	}
}

func TestPacksPayloadAddsTheSectionAndKeepsTheRest(t *testing.T) {
	project, root := newProject(t)
	next := spider.PacksConfig{
		Enabled:  []string{"core", "git"},
		Stacks:   []string{"go"},
		Language: spider.PacksLanguage{Code: "en"},
	}

	payload, err := project.PacksPayload(next)
	if err != nil {
		t.Fatal(err)
	}

	save(t, project, payload)
	got, findings := project.Packs()

	requireNoFindings(t, findings)
	if !got.Present || len(got.Enabled) != 2 || got.Language.Code != "en" {
		t.Fatalf("packs = %+v", got)
	}

	text := readFile(t, root, configFile)
	want := "packs:\n  enabled: [core, git]\n  stacks: [go]\n" +
		"  language:\n    code: en\n"
	if !strings.Contains(text, want) || !strings.Contains(text, "runners:") {
		t.Errorf("config.yaml:\n%s", text)
	}
}
