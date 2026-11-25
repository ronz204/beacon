package ledger

import "context"

type CapabilityRepository interface {
	ListCapabilities(ctx context.Context) ([]Capability, []error)
}
