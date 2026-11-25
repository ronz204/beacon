package kernel

import "strings"

type RecordText struct {
	value string
}

func ParseRecordText(raw string) (RecordText, error) {
	value := strings.TrimSpace(raw)

	if value == "" {
		return RecordText{}, invalidValue(KernelTextInvalid, raw)
	}

	return RecordText{value: value}, nil
}

func (t RecordText) String() string {
	return t.value
}
