package main

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk"
)

func Test_Main(t *testing.T) {
	t.Parallel()
	example(aptos.LocalnetConfig, 100)
}
