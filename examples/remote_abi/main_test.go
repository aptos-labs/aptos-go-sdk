package main

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk"
)

func Test_Main(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test expects network connection to localnet")
	}
	t.Parallel()
	example(aptos.LocalnetConfig)
}
