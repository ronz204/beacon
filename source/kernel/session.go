package kernel

import (
	"context"
	"regexp"
	"strings"
)

var sessionPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

type SessionID struct {
	value string
}

type InjectionLog interface {
	InjectedKeys(
		ctx context.Context,
		session SessionID,
	) (map[string]bool, error)
	RecordInjected(
		ctx context.Context,
		session SessionID,
		keys []string,
	) error
}

func ParseSessionID(raw string) (SessionID, error) {
	value := strings.TrimSpace(raw)

	if !sessionPattern.MatchString(value) {
		return SessionID{}, invalidValue(KernelSessionInvalid, raw)
	}

	return SessionID{value: value}, nil
}

func (s SessionID) String() string {
	return s.value
}
