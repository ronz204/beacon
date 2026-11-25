package match_test

import (
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
)

func FuzzMatchPath(f *testing.F) {
	f.Add("src/**", "src/auth/a.go")
	f.Add("{**,a}/b", "x/b")
	f.Add("x/{a,}.go", "x/.go")
	f.Add("a(b)+[c]?", "a(b)+c")

	matcher := match.NewMatcher()

	f.Fuzz(func(t *testing.T, raw, path string) {
		glob, err := kernel.ParsePathGlob(raw)
		if err != nil {
			return
		}

		if _, err := matcher.MatchPath(glob, path); err != nil {
			t.Fatalf("valid glob %q failed to compile: %v", raw, err)
		}
	})
}
