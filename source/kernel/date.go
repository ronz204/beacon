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

func CalendarDateOf(moment time.Time) (CalendarDate, error) {
	return ParseCalendarDate(moment.Format(time.DateOnly))
}

func (d CalendarDate) Before(other CalendarDate) bool {
	return d.value < other.value
}

func (d CalendarDate) Plus(months, days int) CalendarDate {
	day, err := time.Parse(time.DateOnly, d.value)
	if err != nil {
		return d
	}

	later := day.AddDate(0, months, days)

	return CalendarDate{value: later.Format(time.DateOnly)}
}

func (d CalendarDate) String() string {
	return d.value
}
