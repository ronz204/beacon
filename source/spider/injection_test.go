package spider_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/spider"
)

func TestRecordInjectionWritesTheLocalMetrics(t *testing.T) {
	project, root := newProject(t)

	err := project.RecordInjection(
		context.Background(), "sample-session", "edit", 5,
	)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, ".git", "beacon", "metrics.json")
	data, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(data), `"lines": 5`) {
		t.Errorf("metrics %s, err %v", data, err)
	}
}

func TestRecordInjectionRejectsBadInput(t *testing.T) {
	project, _ := newProject(t)
	ctx := context.Background()

	cases := []struct {
		session string
		event   string
		lines   int
	}{
		{"sample-session", "Edit", 1},
		{"sample-session", "", 1},
		{"sample-session", "edit", -1},
		{"bad session", "edit", 1},
	}

	for _, c := range cases {
		err := project.RecordInjection(ctx, c.session, c.event, c.lines)
		if err == nil {
			t.Errorf("%+v should fail", c)
		}
	}

	err := project.RecordInjection(ctx, "sample-session", "edit", -1)

	var failure spider.InjectionError
	if !errors.As(err, &failure) || failure.Code != spider.InjectionInvalid {
		t.Errorf("err = %v", err)
	}
}
