package blocks_test

import (
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/assets"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

func TestParseBlockAcceptsBothStyles(t *testing.T) {
	cases := []string{
		"<!-- BEGIN BEACON (x) -->\nbody\n<!-- END BEACON -->",
		"# BEGIN BEACON\nbody\n# END BEACON\n",
		"<!-- BEGIN BEACON -->\r\nbody\r\n<!-- END BEACON -->\r\n",
	}

	for _, raw := range cases {
		if _, err := blocks.ParseBlock(raw); err != nil {
			t.Errorf("ParseBlock(%q) failed: %v", raw, err)
		}
	}
}

func TestParseBlockNormalizesLineEndings(t *testing.T) {
	raw := "# BEGIN BEACON\r\nbody\r\n# END BEACON\r\n"

	block, err := blocks.ParseBlock(raw)
	if err != nil {
		t.Fatal(err)
	}

	want := "# BEGIN BEACON\nbody\n# END BEACON"
	if block.String() != want {
		t.Errorf("String() = %q, want %q", block.String(), want)
	}
}

func TestParseBlockRejectsBrokenMarkers(t *testing.T) {
	cases := []string{
		"",
		"body",
		"<!-- BEGIN BEACON -->\nbody",
		"body\n<!-- END BEACON -->",
		"<!-- BEGIN BEACON -->\nbody\n# END BEACON",
		"# BEGIN BEACON\n# BEGIN BEACON\n# END BEACON",
		"# BEGIN BEACON\n# END BEACON\nx\n# END BEACON",
	}

	for _, raw := range cases {
		if _, err := blocks.ParseBlock(raw); err == nil {
			t.Errorf("ParseBlock(%q) should fail", raw)
		}
	}
}

func TestMethodGuideIsAValidShortBlock(t *testing.T) {
	block, err := blocks.ParseBlock(assets.Method)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Count(block.String(), "\n") + 1
	if lines > 30 {
		t.Errorf("method block has %d lines, want at most 30", lines)
	}
}
