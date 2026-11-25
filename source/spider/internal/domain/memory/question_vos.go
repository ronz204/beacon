package memory

import (
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

type QuestionID struct {
	value kernel.RecordID
}

func ParseQuestionID(raw string) (QuestionID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return QuestionID{}, invalidQuestion(QuestionIDInvalid, raw)
	}

	return QuestionID{value: id}, nil
}

func (id QuestionID) String() string {
	return id.value.String()
}

type QuestionBlockID struct {
	value kernel.RecordID
}

func ParseQuestionBlockID(raw string) (QuestionBlockID, error) {
	id, err := kernel.ParseRecordID(raw)

	if err != nil {
		return QuestionBlockID{}, invalidQuestion(QuestionBlockInvalid, raw)
	}

	return QuestionBlockID{value: id}, nil
}

func (id QuestionBlockID) String() string {
	return id.value.String()
}

type QuestionOriginID struct {
	value kernel.RecordID
}

func ParseQuestionOriginID(raw string) (QuestionOriginID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return QuestionOriginID{}, invalidQuestion(QuestionOriginInvalid, raw)
	}

	return QuestionOriginID{value: id}, nil
}

func (id QuestionOriginID) String() string {
	return id.value.String()
}

type QuestionStatus string

const (
	QuestionOpen     QuestionStatus = "open"
	QuestionAnswered QuestionStatus = "answered"
)

func QuestionStatuses() []QuestionStatus {
	return []QuestionStatus{
		QuestionOpen,
		QuestionAnswered,
	}
}

func ParseQuestionStatus(raw string) (QuestionStatus, error) {
	status := QuestionStatus(strings.TrimSpace(raw))

	if !slices.Contains(QuestionStatuses(), status) {
		return "", invalidQuestion(QuestionStatusInvalid, raw)
	}

	return status, nil
}
