package discount_test

import (
	"testing"

	"sample/discount"
)

func TestTotalSumsThenDiscounts(t *testing.T) {
	if got := discount.Total([]int{12000, 8000}); got != 18000 {
		t.Errorf("Total = %d, want 18000", got)
	}
}
