package kernel

import (
	"strings"
	"unicode"
)

type Actor struct {
	value string
}

func ParseActor(raw string) (Actor, error) {
	text := strings.TrimSpace(raw)
	kind, name, found := strings.Cut(text, ":")
	name = strings.TrimSpace(name)

	if !found || !isActorKind(kind) {
		return Actor{}, invalidValue(KernelActorInvalid, raw)
	}
	if !isActorName(name) {
		return Actor{}, invalidValue(KernelActorInvalid, raw)
	}

	return Actor{value: kind + ":" + name}, nil
}

func (a Actor) IsHuman() bool {
	return strings.HasPrefix(a.value, "human:")
}

func (a Actor) String() string {
	return a.value
}

func isActorKind(kind string) bool {
	return kind == "human" || kind == "agent"
}

func isActorName(name string) bool {
	return name != "" && !hasControlCharacter(name)
}

func hasControlCharacter(text string) bool {
	return strings.IndexFunc(text, unicode.IsControl) >= 0
}
