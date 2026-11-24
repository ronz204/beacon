package ledger_test

import (
	"errors"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func requireCode[C ~string](t *testing.T, err error, want C) {
	t.Helper()

	var got kernel.Error[C]
	if !errors.As(err, &got) {
		t.Fatalf("expected domain error %q, got %v", want, err)
	}
	if got.Code != want {
		t.Fatalf("code = %q, want %q", got.Code, want)
	}
}

func TestParseGoalIDAcceptsRootID(t *testing.T) {
	got, err := ledger.ParseGoalID("bc-0001")

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "bc-0001" {
		t.Fatalf("got %q", got)
	}
}

func TestParseGoalIDRejectsInvalid(t *testing.T) {
	invalid := []string{
		"",
		"goal",
		"bc-0001.1",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := ledger.ParseGoalID(raw)

			requireCode(t, err, ledger.GoalIDInvalid)
		})
	}
}

func TestParseGoalStatusAcceptsClosedSet(t *testing.T) {
	for _, status := range ledger.GoalStatuses() {
		t.Run(string(status), func(t *testing.T) {
			got, err := ledger.ParseGoalStatus(" " + string(status) + " ")

			if err != nil {
				t.Fatal(err)
			}
			if got != status {
				t.Fatalf("got %q", got)
			}
		})
	}
}

func TestParseGoalStatusRejectsInvalid(t *testing.T) {
	invalid := []string{
		"",
		"paused",
		"ACTIVE",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := ledger.ParseGoalStatus(raw)

			requireCode(t, err, ledger.GoalStatusInvalid)
		})
	}
}
