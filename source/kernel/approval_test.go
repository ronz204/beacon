package kernel_test

import (
	"testing"

	"github.com/ronz204/beacon/source/kernel"
)

func TestParseApprovalWithoutVia(t *testing.T) {
	got, err := kernel.ParseApproval(kernel.ApprovalInput{
		By: "human:jane",
		At: "2026-11-22",
	})

	if err != nil {
		t.Fatal(err)
	}
	if got.By().String() != "human:jane" {
		t.Fatalf("by = %q", got.By())
	}
	if got.At().String() != "2026-11-22" {
		t.Fatalf("at = %q", got.At())
	}
	if _, found := got.Via(); found {
		t.Fatal("via should be absent")
	}
}

func TestParseApprovalWithAgentVia(t *testing.T) {
	got, err := kernel.ParseApproval(kernel.ApprovalInput{
		By:  "human:jane",
		At:  "2026-11-22",
		Via: "agent:sample-agent",
	})
	if err != nil {
		t.Fatal(err)
	}

	via, found := got.Via()

	if !found || via.String() != "agent:sample-agent" {
		t.Fatalf("via = %q, found = %v", via, found)
	}
}

func TestParseApprovalRejectsInvalid(t *testing.T) {
	cases := []struct {
		name  string
		input kernel.ApprovalInput
		code  kernel.KernelCode
	}{
		{
			"by",
			kernel.ApprovalInput{By: "jane", At: "2026-11-22"},
			kernel.KernelActorInvalid,
		},
		{
			"at",
			kernel.ApprovalInput{By: "human:jane", At: "2026-02-30"},
			kernel.KernelDateInvalid,
		},
		{
			"via format",
			kernel.ApprovalInput{
				By:  "human:jane",
				At:  "2026-11-22",
				Via: "sample-agent",
			},
			kernel.KernelActorInvalid,
		},
		{
			"via human",
			kernel.ApprovalInput{
				By:  "human:jane",
				At:  "2026-11-22",
				Via: "human:john",
			},
			kernel.KernelApprovalInvalid,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := kernel.ParseApproval(c.input)

			requireKernelCode(t, err, c.code)
		})
	}
}
