package match_test

import (
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
)

func TestMatchPath(t *testing.T) {
	cases := []struct {
		glob string
		path string
		want bool
	}{
		{"src/auth/**", "src/auth", true},
		{"src/auth/**", "src/auth/login.go", true},
		{"src/auth/**", "src/auth/a/b.go", true},
		{"src/auth/**", "src/authx/a.go", false},
		{"**", "a.go", true},
		{"**", "x/y/a.go", true},
		{"**/*.go", "a.go", true},
		{"**/*.go", "x/y/a.go", true},
		{"**/*.go", "x/a.txt", false},
		{"src/*.go", "src/a.go", true},
		{"src/*.go", "src/x/a.go", false},
		{"src/?.go", "src/a.go", true},
		{"src/?.go", "src/ab.go", false},
		{"a?b", "a/b", false},
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/x/b", true},
		{"a/**/b", "a/x/y/b", true},
		{"a/**/b", "a/xb", false},
		{"{src,tests}/**", "tests/x.go", true},
		{"{src,tests}/**", "lib/x.go", false},
		{"{src/**,tests/*.go}", "src/a/b.go", true},
		{"{src/**,tests/*.go}", "tests/a.go", true},
		{"{src/**,tests/*.go}", "tests/a/b.go", false},
		{"x/{a,}.go", "x/a.go", true},
		{"x/{a,}.go", "x/.go", true},
		{"{**,a}/b", "b", true},
		{"{**,a}/b", "x/y/b", true},
		{"a*/{**,a}/?/**", "ab/x", true},
		{"x{**,a}", "xb", true},
		{"x{**,a}", "xb/c", false},
		{"a,b/c.go", "a,b/c.go", true},
		{"Src/**", "src/a.go", false},
		{"a.b/c+d(e).go", "a.b/c+d(e).go", true},
		{"a.b", "axb", false},
		{"docs/my file.md", "docs/my file.md", true},
		{"src/**", `src\auth\a.go`, true},
		{".github/**", ".github/workflows/ci.yaml", true},
	}

	matcher := match.NewMatcher()

	for _, c := range cases {
		t.Run(c.glob+" "+c.path, func(t *testing.T) {
			glob, err := kernel.ParsePathGlob(c.glob)
			if err != nil {
				t.Fatal(err)
			}

			got, err := matcher.MatchPath(glob, c.path)

			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}
