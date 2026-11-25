package agenda

import "context"

type ClaimRepository interface {
	ListClaims(ctx context.Context) ([]TaskClaim, error)
	SaveClaim(ctx context.Context, claim TaskClaim) error
	RemoveClaims(ctx context.Context, tasks []TaskID) error
}
