package kernel_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
)

func TestParseActorAcceptsAndNormalizes(t *testing.T) {
	cases := []struct {
		raw     string
		text    string
		isHuman bool
	}{
		{"human:jane", "human:jane", true},
		{"human:jane doe", "human:jane doe", true},
		{"  human:jane  ", "human:jane", true},
		{"human: jane ", "human:jane", true},
		{"agent:sample-agent", "agent:sample-agent", false},
		{"agent:a:b", "agent:a:b", false},
	}

	for _, c := range cases {
		t.Run(c.raw, func(t *testing.T) {
			got, err := kernel.ParseActor(c.raw)

			if err != nil {
				t.Fatal(err)
			}
			if got.String() != c.text {
				t.Fatalf("text = %q, want %q", got, c.text)
			}
			if got.IsHuman() != c.isHuman {
				t.Fatalf("human = %v", got.IsHuman())
			}
		})
	}
}

func TestParseActorRejectsInvalid(t *testing.T) {
	invalid := []string{
		"",
		"jane",
		"robot:x",
		"human:",
		"human:   ",
		"Human:jane",
		"human :jane",
		"human:a\nb",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := kernel.ParseActor(raw)

			requireKernelCode(t, err, kernel.KernelActorInvalid)
		})
	}
}
