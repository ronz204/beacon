package assets_test

import (
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/assets"
)

const guideLimit = 30

func TestTheGuideStaysShortAndKnowsAttestAndFile(t *testing.T) {
	lines := strings.Count(assets.Method, "\n") + 1

	if lines >= guideLimit {
		t.Errorf("the guide has %d lines", lines)
	}
	for _, want := range []string{"beacon record --file", "beacon attest"} {
		if !strings.Contains(assets.Method, want) {
			t.Errorf("the guide does not mention %q", want)
		}
	}
}
