package main

import (
	"testing"

	"github.com/qimeila/aptos-go-sdk"
)

func Test_Main(t *testing.T) {
	t.Parallel()
	example(aptos.LocalnetConfig)
}
