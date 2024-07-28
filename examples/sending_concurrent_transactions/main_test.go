package main

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/client"
)

func Test_Main(t *testing.T) {
	t.Parallel()
	example(client.LocalnetConfig, 100)
}
