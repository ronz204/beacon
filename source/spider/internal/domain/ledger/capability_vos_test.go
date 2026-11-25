package ledger_test

import (
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
)

func TestParseRootIDsRejectChildren(t *testing.T) {
	const child = "bc-91cd.1"

	_, err := ledger.ParseCapabilityID(child)
	requireCode(t, err, ledger.CapabilityIDInvalid)

	_, err = ledger.ParsePromiseID(child)
	requireCode(t, err, ledger.PromiseIDInvalid)

	_, err = ledger.ParseOriginWorkID(child)
	requireCode(t, err, ledger.PromiseOriginInvalid)
}

func TestParsePromiseStatus(t *testing.T) {
	for _, status := range ledger.PromiseStatuses() {
		got, err := ledger.ParsePromiseStatus(string(status))

		if err != nil {
			t.Fatal(err)
		}
		if got != status {
			t.Fatalf("got %q", got)
		}
	}

	_, err := ledger.ParsePromiseStatus("draft")
	requireCode(t, err, ledger.PromiseStatusInvalid)
}

func TestParsePromiseCheckWithRunner(t *testing.T) {
	params := map[string]string{"pkg": "api", "target": "auth/lockout"}

	check, err := ledger.ParsePromiseCheck(ledger.PromiseCheckInput{
		Runner: "test-node",
		Params: params,
	})
	if err != nil {
		t.Fatal(err)
	}

	params["pkg"] = "changed"
	check.Params()["target"] = "changed"

	if check.IsManual() || check.Runner() != "test-node" {
		t.Fatalf("runner = %q", check.Runner())
	}
	if check.Params()["pkg"] != "api" {
		t.Fatalf("params = %v", check.Params())
	}
	if check.Params()["target"] != "auth/lockout" {
		t.Fatalf("params = %v", check.Params())
	}
}

func TestParsePromiseCheckManual(t *testing.T) {
	check, err := ledger.ParsePromiseCheck(ledger.PromiseCheckInput{
		Manual: "  Open the login page  ",
	})

	if err != nil {
		t.Fatal(err)
	}
	if !check.IsManual() || check.Manual() != "Open the login page" {
		t.Fatalf("manual = %q", check.Manual())
	}
}

func TestParsePromiseCheckRejectsInvalid(t *testing.T) {
	cases := map[string]ledger.PromiseCheckInput{
		"empty": {},
		"both":  {Runner: "test-node", Manual: "Open it"},
		"manual with params": {
			Manual: "Open it",
			Params: map[string]string{"pkg": "api"},
		},
		"runner option": {Runner: "--config"},
		"param option": {
			Runner: "test-node",
			Params: map[string]string{"target": "--watch"},
		},
		"param option after trim": {
			Runner: "test-node",
			Params: map[string]string{"target": "  -x"},
		},
		"param blank": {
			Runner: "test-node",
			Params: map[string]string{"target": "  "},
		},
		"param control character": {
			Runner: "test-node",
			Params: map[string]string{"target": "a\x00b"},
		},
		"key blank": {
			Runner: "test-node",
			Params: map[string]string{" ": "api"},
		},
		"key padded": {
			Runner: "test-node",
			Params: map[string]string{" pkg": "api"},
		},
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ledger.ParsePromiseCheck(input)

			requireCode(t, err, ledger.PromiseCheckInvalid)
		})
	}
}

func TestParsePromiseAttestation(t *testing.T) {
	got, err := ledger.ParsePromiseAttestation(
		ledger.PromiseAttestationInput{
			By:       "human:jane",
			At:       "2026-11-22",
			Commit:   "9f3c2ab",
			Evidence: " docs/evidence/login.png ",
		},
	)

	if err != nil {
		t.Fatal(err)
	}
	if got.By().String() != "human:jane" {
		t.Fatalf("by = %q", got.By())
	}
	if got.At().String() != "2026-11-22" {
		t.Fatalf("at = %q", got.At())
	}
	if got.Commit() != "9f3c2ab" {
		t.Fatalf("commit = %q", got.Commit())
	}
	if got.Evidence() != "docs/evidence/login.png" {
		t.Fatalf("evidence = %q", got.Evidence())
	}
}

func TestParsePromiseAttestationRejectsInvalid(t *testing.T) {
	valid := ledger.PromiseAttestationInput{
		By:     "human:jane",
		At:     "2026-11-22",
		Commit: "9f3c2ab",
	}
	agent, badDate, shortCommit, upperCommit := valid, valid, valid, valid
	agent.By = "agent:sample-agent"
	badDate.At = "2026-02-30"
	shortCommit.Commit = "9f3c"
	upperCommit.Commit = "9F3C2AB"

	cases := map[string]ledger.PromiseAttestationInput{
		"agent":        agent,
		"date":         badDate,
		"short commit": shortCommit,
		"upper commit": upperCommit,
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ledger.ParsePromiseAttestation(input)

			requireCode(t, err, ledger.PromiseAttestInvalid)
		})
	}
}
