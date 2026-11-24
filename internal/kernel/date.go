package kernel

import (
	"strings"
	"time"
)

type CalendarDate struct {
	value string
}

func ParseCalendarDate(raw string) (CalendarDate, error) {
	value := strings.TrimSpace(raw)

	if _, err := time.Parse(time.DateOnly, value); err != nil {
		return CalendarDate{}, invalidValue(KernelDateInvalid, raw)
	}

	return CalendarDate{value: value}, nil
}

func (d CalendarDate) String() string {
	return d.value
}
