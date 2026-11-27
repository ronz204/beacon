package benchmark_test

import (
	"os"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

func TestMain(m *testing.M) {
	os.Exit(utilities.RunSuite(m))
}
