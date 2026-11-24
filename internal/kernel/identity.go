package kernel

import (
	"regexp"
	"strings"
)

var recordIDPattern = regexp.MustCompile(
	`^bc-[0-9a-f]{4,}(\.[1-9][0-9]*)*$`,
)

type RecordID struct {
	value string
}

func ParseRecordID(raw string) (RecordID, error) {
	value := strings.TrimSpace(raw)

	if !recordIDPattern.MatchString(value) {
		return RecordID{}, invalidValue(KernelRecordIDInvalid, raw)
	}

	return RecordID{value: value}, nil
}

func (id RecordID) String() string {
	return id.value
}

func (id RecordID) Parent() (RecordID, bool) {
	cut := strings.LastIndexByte(id.value, '.')

	if cut < 0 {
		return RecordID{}, false
	}

	return RecordID{value: id.value[:cut]}, true
}
