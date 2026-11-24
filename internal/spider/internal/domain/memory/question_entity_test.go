package memory_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
)

func sampleQuestionInput() memory.QuestionInput {
	return memory.QuestionInput{
		ID:     must(memory.ParseQuestionID("bc-e5f6")),
		Author: must(kernel.ParseActor("agent:sample-agent")),
		Text:   sampleText("How many failed attempts before locking?"),
		Blocks: []memory.QuestionBlockID{
			must(memory.ParseQuestionBlockID("bc-91cd.1")),
		},
		Origin: must(memory.ParseQuestionOriginID("bc-91cd")),
	}
}

func answeredState(answer kernel.RecordText) memory.QuestionState {
	input := sampleQuestionInput()

	return memory.QuestionState{
		ID:      input.ID,
		Created: must(kernel.ParseCalendarDate("2026-11-22")),
		Author:  input.Author,
		Text:    input.Text,
		Status:  memory.QuestionAnswered,
		Answer:  answer,
	}
}

func TestQuestionReferenceIDs(t *testing.T) {
	_, err := memory.ParseQuestionID("bc-e5f6.1")
	requireCode(t, err, memory.QuestionIDInvalid)

	_, err = memory.ParseQuestionBlockID("blocked")
	requireCode(t, err, memory.QuestionBlockInvalid)

	_, err = memory.ParseQuestionOriginID("bc-91cd.1")
	requireCode(t, err, memory.QuestionOriginInvalid)

	block := must(memory.ParseQuestionBlockID("bc-91cd.1"))
	if block.String() != "bc-91cd.1" {
		t.Fatalf("block = %q", block)
	}
}

func TestParseQuestionStatus(t *testing.T) {
	for _, status := range memory.QuestionStatuses() {
		got, err := memory.ParseQuestionStatus(string(status))

		if err != nil {
			t.Fatal(err)
		}
		if got != status {
			t.Fatalf("got %q", got)
		}
	}

	_, err := memory.ParseQuestionStatus("closed")
	requireCode(t, err, memory.QuestionStatusInvalid)
}

func TestNewQuestionStartsOpenAndDated(t *testing.T) {
	question, err := memory.NewQuestion(sampleQuestionInput(), sampleNow)

	if err != nil {
		t.Fatal(err)
	}
	if question.Status() != memory.QuestionOpen {
		t.Fatalf("status = %q", question.Status())
	}
	if question.Created().String() != "2026-11-22" {
		t.Fatalf("created = %q", question.Created())
	}
	if _, found := question.Answer(); found {
		t.Fatal("a new question has no answer")
	}
}

func TestNewQuestionRejectsMissingField(t *testing.T) {
	input := sampleQuestionInput()
	input.Text = kernel.RecordText{}

	_, err := memory.NewQuestion(input, sampleNow)

	requireCode(t, err, memory.QuestionFieldMissing)
}

func TestValidateQuestionRequiresAnswerWhenAnswered(t *testing.T) {
	err := memory.ValidateQuestionState(answeredState(kernel.RecordText{}))

	requireCode(t, err, memory.QuestionAnswerMissing)
}

func TestValidateQuestionAcceptsAnswered(t *testing.T) {
	state := answeredState(sampleText("Five attempts, fifteen minutes"))

	if err := memory.ValidateQuestionState(state); err != nil {
		t.Fatal(err)
	}

	answer, found := memory.RestoreQuestion(state).Answer()
	if !found || answer.String() != "Five attempts, fifteen minutes" {
		t.Fatalf("answer = %q, found = %v", answer, found)
	}
}

func TestQuestionProtectsItsBlocks(t *testing.T) {
	input := sampleQuestionInput()
	question := must(memory.NewQuestion(input, sampleNow))
	replacement := must(memory.ParseQuestionBlockID("bc-ffff"))

	input.Blocks[0] = replacement
	question.Blocks()[0] = replacement

	if question.Blocks()[0].String() != "bc-91cd.1" {
		t.Fatalf("blocks = %v", question.Blocks())
	}
}

func TestQuestionIdentityAndOptionalFields(t *testing.T) {
	question := must(memory.NewQuestion(sampleQuestionInput(), sampleNow))
	restored := memory.RestoreQuestion(memory.QuestionState{
		ID: question.ID(),
	})

	if !restored.SameAs(question) {
		t.Fatal("questions with the same id should be the same")
	}
	if _, found := question.Origin(); !found {
		t.Fatal("origin should be present")
	}
	if _, found := question.Approved(); found {
		t.Fatal("approved should be absent")
	}
}
