package kernel_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
)

func TestParsePathGlobAccepts(t *testing.T) {
	valid := []string{
		"**",
		"**/*.go",
		"src/auth/**",
		"apps/api/src/auth/**",
		"src/?.go",
		"src/{a,b}/*.go",
		"{src,tests}/**",
		"{src/**,tests/**}",
		"x/{a,}.go",
		"docs/my file.md",
		"a!b/c.go",
		"a,b/c.go",
		".github/workflows/*.yaml",
	}

	for _, raw := range valid {
		t.Run(raw, func(t *testing.T) {
			got, err := kernel.ParsePathGlob(raw)

			if err != nil {
				t.Fatal(err)
			}
			if got.String() != raw {
				t.Fatalf("got %q", got)
			}
		})
	}
}

func TestParsePathGlobTrims(t *testing.T) {
	got, err := kernel.ParsePathGlob("  src/**  ")

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "src/**" {
		t.Fatalf("got %q", got)
	}
}

func TestParsePathGlobRejectsInvalid(t *testing.T) {
	invalid := []string{
		"",
		"!src/**",
		"/src/**",
		"C:/src",
		`src\auth`,
		"src//a",
		"src/",
		"../x",
		"src/../x",
		"./src",
		"src/**a",
		"a**/b",
		"***",
		"src/[id]/x",
		"{a,b",
		"a}",
		"{a,{b,c}}",
		"a\tb",
		"a\xffb",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := kernel.ParsePathGlob(raw)

			requireKernelCode(t, err, kernel.KernelGlobInvalid)
		})
	}
}
