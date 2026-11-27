package discount_test

import (
	"testing"

	"sample/discount"
)

func TestTenPercentOverOneHundred(t *testing.T) {
	if got := discount.Apply(20000); got != 18000 {
		t.Errorf("Apply(20000) = %d, want 18000", got)
	}
}
