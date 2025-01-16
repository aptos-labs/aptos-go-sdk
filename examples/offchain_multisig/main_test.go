package main

import (
	"github.com/aptos-labs/aptos-go-sdk"
	"testing"
)

func Test_Main(t *testing.T) {
	t.Parallel()
	example(aptos.LocalnetConfig)
}
