package action_test

import (
	"context"
	"testing"

	"github.com/ronz204/beacon/source/snares/internal/action"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
)

type fakePorts struct {
	check weaver.CheckView
}

func (fakePorts) Doctor(context.Context) []spider.Finding { return nil }

func (fakePorts) HasHead(context.Context) bool { return true }

func (fakePorts) Diff(
	context.Context,
	string,
	string,
) (weaver.DiffView, []spider.Signal, error) {
	return weaver.DiffView{}, nil, nil
}

func (f fakePorts) Check(
	context.Context,
	string,
	bool,
	string,
) (weaver.CheckView, []spider.Signal, error) {
	return f.check, nil, nil
}

func TestPreCommitOnlyWarnsAboutDrift(t *testing.T) {
	ports := fakePorts{check: weaver.CheckView{
		Promises: []weaver.PromiseResultView{{ID: "bp-1", Drifted: true}},
	}}

	result, err := action.PreCommit{Ports: ports}.Run(context.Background())

	if err != nil || len(result.Blocks) != 0 || len(result.Warnings) != 1 {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
}
