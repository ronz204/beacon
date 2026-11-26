package e2e_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

const (
	capabilityPath = ".beacon/records/capabilities/bc-7f3a-auth-login.yaml"
	configPath     = ".beacon/config.yaml"
	sampleAgent    = "agent:sample-agent"
	lockoutText    = "After 5 failed attempts the account locks for 15 minutes."
)

func storedPayload(t *testing.T, root, id string) map[string]any {
	t.Helper()

	result := utils.Beacon(t, root, "get", id, "--json")

	var data map[string]any
	decodeData(t, result.Stdout, &data)

	return map[string]any{"type": data["type"], "record": data["record"]}
}

func promiseOf(payload map[string]any, id string) map[string]any {
	record := payload["record"].(map[string]any)

	for _, item := range record["promises"].([]any) {
		promise := item.(map[string]any)
		if promise["id"] == id {
			return promise
		}
	}

	return nil
}

func decodeError(t *testing.T, text string, target any) {
	t.Helper()

	var parsed struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(parsed.Error, target); err != nil {
		t.Fatal(err)
	}
}

func jsonOf(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func record(
	t *testing.T,
	root, actor, payload string,
	flags ...string,
) utils.Result {
	t.Helper()

	args := append([]string{"record", "--json"}, flags...)
	args = append(args, "-")

	return utils.BeaconWith(
		t,
		root,
		utils.Options{Stdin: payload, Actor: actor},
		args...,
	)
}

func weakenedLockout(t *testing.T, root string) string {
	t.Helper()

	payload := storedPayload(t, root, "bc-7f3a")
	promise := promiseOf(payload, "bc-c3d4")
	promise["text"] = strings.Replace(lockoutText, "5", "10", 1)

	return jsonOf(t, payload)
}

func TestRecordOfAnUnchangedRecordTouchesNothing(t *testing.T) {
	root := utils.NewRepository(t)
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	before := utils.ReadFile(t, path)
	payload := jsonOf(t, storedPayload(t, root, "bc-7f3a"))

	result := record(t, root, sampleAgent, payload)

	var data struct {
		Unchanged bool `json:"unchanged"`
		Saved     bool `json:"saved"`
	}
	decodeData(t, result.Stdout, &data)

	if result.Code != 0 || !data.Unchanged || data.Saved {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if utils.ReadFile(t, path) != before {
		t.Fatal("the file changed")
	}
}

func TestRecordBlocksAWeakenedPromiseWithoutAYes(t *testing.T) {
	root := utils.NewRepository(t)
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	before := utils.ReadFile(t, path)

	result := record(t, root, sampleAgent, weakenedLockout(t, root))

	envelope := parseRead(t, result.Stdout)
	if result.Code != 2 || envelope.OK {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if envelope.Error.Code != "approval_required" {
		t.Fatalf("error = %+v", envelope.Error)
	}
	if !hintContains(envelope.Hint, "--approve") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
	if utils.ReadFile(t, path) != before {
		t.Fatal("the file changed without a yes")
	}

	var asked struct {
		Details struct {
			Changes []struct {
				ID     string `json:"id"`
				Reason string `json:"reason"`
				Before string `json:"before"`
				After  string `json:"after"`
			} `json:"changes"`
		} `json:"details"`
	}
	decodeError(t, result.Stdout, &asked)

	change := asked.Details.Changes[0]
	if change.ID != "bc-c3d4" || change.Reason != "promise_text_changed" {
		t.Fatalf("change = %+v", change)
	}
	if change.Before != lockoutText || !strings.Contains(change.After, "10") {
		t.Fatalf("change = %+v", change)
	}
}

func TestRecordWithApproveStampsThePersonAndTheAgent(t *testing.T) {
	root := utils.NewRepository(t)
	payload := weakenedLockout(t, root)

	result := record(t, root, sampleAgent, payload, "--approve")

	text := utils.ReadFile(t, filepath.Join(root, capabilityPath))
	if result.Code != 0 {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !strings.Contains(text, "After 10 failed attempts") {
		t.Fatalf("the text did not change:\n%s", text)
	}
	if !strings.Contains(text, `by: "human:jane"`) ||
		!strings.Contains(text, `via: "agent:sample-agent"`) {
		t.Fatalf("approval missing:\n%s", text)
	}
}

func TestRecordByAPersonStampsThemWithoutVia(t *testing.T) {
	root := utils.NewRepository(t)
	payload := weakenedLockout(t, root)

	result := record(t, root, "human:sam", payload, "--approve")

	text := utils.ReadFile(t, filepath.Join(root, capabilityPath))
	if result.Code != 0 || !strings.Contains(text, `by: "human:sam"`) {
		t.Fatalf("code = %d, text:\n%s", result.Code, text)
	}
	if strings.Contains(text, "human:sam\", at: 2") &&
		strings.Contains(text, `via: "agent:sample-agent"`) {
		t.Fatalf("a person must not carry via:\n%s", text)
	}
}

func TestRecordByAnAgentNeedsTheGitName(t *testing.T) {
	root := utils.NewRepository(t)
	utils.Git(t, root, "config", "user.name", "")
	payload := weakenedLockout(t, root)

	result := record(t, root, sampleAgent, payload, "--approve")

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "identity_missing" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !hintContains(envelope.Hint, "user.name") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
}

func TestRecordAcceptsANewProposedPromiseWithoutAYes(t *testing.T) {
	root := utils.NewRepository(t)
	payload := storedPayload(t, root, "bc-7f3a")
	record := payload["record"].(map[string]any)
	record["promises"] = append(record["promises"].([]any), map[string]any{
		"text":   "Locked accounts show a neutral message.",
		"check":  map[string]any{"manual": "Lock an account and read the page"},
		"status": "proposed",
	})

	result := utils.BeaconWith(
		t,
		root,
		utils.Options{Stdin: jsonOf(t, payload), Actor: sampleAgent},
		"record", "--json", "-",
	)

	text := utils.ReadFile(t, filepath.Join(root, capabilityPath))
	if result.Code != 0 || !strings.Contains(text, "neutral message") {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestRecordRejectsAnOmittedPromise(t *testing.T) {
	root := utils.NewRepository(t)
	payload := storedPayload(t, root, "bc-7f3a")
	record := payload["record"].(map[string]any)
	record["promises"] = record["promises"].([]any)[:1]

	result := utils.BeaconWith(
		t,
		root,
		utils.Options{Stdin: jsonOf(t, payload), Actor: sampleAgent},
		"record", "--json", "-",
	)

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "promise_omitted" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !hintContains(envelope.Hint, "beacon close") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
}

func TestRecordRejectsAnUnknownRecord(t *testing.T) {
	root := utils.NewRepository(t)
	payload := storedPayload(t, root, "bc-a1b2")
	payload["record"].(map[string]any)["id"] = "bc-ffff"

	result := record(t, root, sampleAgent, jsonOf(t, payload))

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "record_missing" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestRecordRejectsAnInvalidPayload(t *testing.T) {
	root := utils.NewRepository(t)

	result := record(t, root, sampleAgent, `{"type":"question","record":{}}`)

	envelope := parseRead(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "payload_invalid" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestRecordRequiresTheStdinMarker(t *testing.T) {
	root := utils.NewRepository(t)

	result := utils.Beacon(t, root, "record", "payload.json")

	if result.Code != 1 || !strings.Contains(result.Stderr, "stdin") {
		t.Fatalf("code = %d, stderr = %s", result.Code, result.Stderr)
	}
}

const changedConfig = `{"type":"config","record":{"schema":1,` +
	`"runners":{"test-node":{"run":["pnpm","--filter","{pkg}","test",` +
	`"--","{target}"]},"visual":{"run":["pnpm","exec","playwright",` +
	`"test","--grep","{target}"],"baseline":["tests/**/*-snapshots/**"]},` +
	`"test-go":{"run":["go","test","{pkg}"]}},` +
	`"review":{"convention":"6mo","architecture":"12mo","product":"12mo",` +
	`"domain":"12mo","runbook":"6mo","decision":"6mo"},` +
	`"limits":{"guide":150,"brief":60,"hook":20},` +
	`"agents":["claude-code","codex","cursor"]}}`

func TestRecordOfChangedRunnersAsksThenStampsTheConfig(t *testing.T) {
	root := utils.NewRepository(t)

	blocked := record(t, root, sampleAgent, changedConfig)
	approved := record(t, root, sampleAgent, changedConfig, "--approve")

	text := utils.ReadFile(t, filepath.Join(root, configPath))
	if blocked.Code != 2 || approved.Code != 0 {
		t.Fatalf("blocked = %+v, approved = %+v", blocked, approved)
	}
	if !strings.Contains(text, "test-go") ||
		!strings.Contains(text, "approved") {
		t.Fatalf("config:\n%s", text)
	}
}

func TestRecordApproveOfAnUnchangedRecordWithNothingPendingTouchesNothing(
	t *testing.T,
) {
	root := utils.NewRepository(t)
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	before := utils.ReadFile(t, path)
	payload := jsonOf(t, storedPayload(t, root, "bc-7f3a"))

	result := record(t, root, sampleAgent, payload, "--approve")

	if result.Code != 0 || utils.ReadFile(t, path) != before {
		t.Fatalf("result = %+v", result)
	}
}
