package deltas

import "github.com/ronz204/beacon/source/kernel"

type ApprovalCode string

const (
	ApprovalIdentityMissing ApprovalCode = "identity_missing"
)

type ApprovalError = kernel.Error[ApprovalCode]

var approvalErrors = kernel.ErrorCatalog[ApprovalCode]{
	ApprovalIdentityMissing: "an agent approval needs the git user.name",
}
