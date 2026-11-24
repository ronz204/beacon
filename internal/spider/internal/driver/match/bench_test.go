package match_test

import (
	"fmt"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/driver/match"
)

func BenchmarkMatchPathRepeated(b *testing.B) {
	matcher := match.NewMatcher()
	globs := make([]kernel.PathGlob, 0, 50)
	for index := range 50 {
		raw := fmt.Sprintf("src/module%d/**", index)
		glob, err := kernel.ParsePathGlob(raw)
		if err != nil {
			b.Fatal(err)
		}
		globs = append(globs, glob)
	}

	for b.Loop() {
		for _, glob := range globs {
			for file := range 40 {
				path := fmt.Sprintf("src/module%d/file.go", file)
				if _, err := matcher.MatchPath(glob, path); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}
