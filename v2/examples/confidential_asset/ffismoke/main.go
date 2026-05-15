//go:build cgo

// ffismoke verifies confidential-asset-bindings FFI linkage (BatchRangeProof → BatchVerifyProof).
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/aptos-labs/aptos-go-sdk/v2/confidentialasset"
)

func main() {
	err := confidentialasset.RunBindingsBatchSmoke()
	if errors.Is(err, confidentialasset.ErrBindingsSmokeSkipped) {
		fmt.Println("SKIP_CONFIDENTIAL_BINDINGS=1 — skipped FFI smoke.")
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("aptosconfidential FFI OK (BatchRangeProof → BatchVerifyProof)")
}
