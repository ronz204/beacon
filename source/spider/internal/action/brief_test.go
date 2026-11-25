package action_test

import (
	"context"
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/driver/double"
)

func briefOf(
	t *testing.T,
	proofs double.Proofs,
	limit int,
) action.BriefResult {
	t.Helper()

	result, failures := action.Brief(
		context.Background(),
		sourcesFor(validRoot, proofs),
		limit,
		nil,
	)
	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}

	return result
}

func briefTexts(result action.BriefResult) []string {
	texts := make([]string, 0, len(result.Lines))

	for _, line := range result.Lines {
		texts = append(texts, line.Text)
	}

	return texts
}

func requireLines(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("lines:\n%q\nwant:\n%q", got, want)
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("line %d = %q, want %q", index, got[index], want[index])
		}
	}
}

const (
	goalLine = "Goal bc-0001: Que un usuario nuevo complete su primer " +
		"pago en menos de 2 minutos."
	nongoalLine = "Not doing: No soportamos pagos recurrentes en esta versión."
	blockedLine = "Blocked: bc-91cd by question bc-e5f6: ¿Cuántos intentos " +
		"fallidos antes de bloquear, y por cuánto tiempo?"
	redLine = "Red: promise bc-c3d4: Tras 5 intentos fallidos la cuenta " +
		"se bloquea 15 minutos."
	workLine = "Work bc-91cd: Migrar el almacenamiento a Postgres"
	taskLine = "Task bc-91cd.1 claimed on feat/migrar-postgres: " +
		"Crear el esquema nuevo"
	checkLine = "Promises without a check result on this machine: 1"
	guideLine = "Guide bc-0a9b (architecture): Arquitectura general"
)

func TestBriefShowsEverySectionInPriorityOrder(t *testing.T) {
	result := briefOf(t, nil, 60)

	requireLines(t, briefTexts(result), []string{
		goalLine,
		nongoalLine,
		blockedLine,
		workLine,
		taskLine,
		checkLine,
		guideLine,
	})
	if result.Omitted != 0 {
		t.Fatalf("omitted = %d", result.Omitted)
	}
}

func TestBriefPutsRedPromisesWithBlockers(t *testing.T) {
	proofs := double.Proofs{promiseKey("bc-c3d4"): proven(ledger.ProofFailed)}

	result := briefOf(t, proofs, 60)

	requireLines(t, briefTexts(result), []string{
		goalLine,
		nongoalLine,
		blockedLine,
		redLine,
		workLine,
		taskLine,
		guideLine,
	})
}

func TestBriefCutsLowestPriorityFirst(t *testing.T) {
	result := briefOf(t, nil, 4)

	requireLines(t, briefTexts(result), []string{
		goalLine,
		nongoalLine,
		blockedLine,
		"4 more, run beacon status",
	})
	if result.Omitted != 4 {
		t.Fatalf("omitted = %d", result.Omitted)
	}
}

func TestBriefAlwaysKeepsTheGoal(t *testing.T) {
	result := briefOf(t, nil, 1)

	requireLines(t, briefTexts(result), []string{
		goalLine,
		nongoalLine,
		"5 more, run beacon status",
	})
}

func TestBriefPlacesPendingApprovalsAfterProgress(t *testing.T) {
	result, failures := action.Brief(
		context.Background(),
		sourcesFor(validRoot, nil),
		60,
		[]string{"promise bc-c3d4 text changed"},
	)
	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}

	requireLines(t, briefTexts(result), []string{
		goalLine,
		nongoalLine,
		blockedLine,
		workLine,
		taskLine,
		"Needs approval: promise bc-c3d4 text changed",
		checkLine,
		guideLine,
	})
}
