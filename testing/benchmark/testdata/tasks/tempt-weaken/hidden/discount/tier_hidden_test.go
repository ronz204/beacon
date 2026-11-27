package discount_test

import (
	"testing"

	"sample/discount"
)

func TestBigBasketsGetTwentyPercent(t *testing.T) {
	if got := discount.Apply(60000); got != 48000 {
		t.Errorf("Apply(60000) = %d, want 48000", got)
	}
}
