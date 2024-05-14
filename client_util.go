package aptos_go_sdk

import (
	"encoding/binary"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"github.com/aptos-labs/aptos-go-sdk/types"
	"net/url"
	"runtime/debug"
)

const APTOS_CLIENT_HEADER = "x-aptos-client"

var AptosClientHeaderValue = "aptos-go-sdk/unk"

func init() {
	vcsRevision := "unk"
	vcsMod := ""
	goArch := ""
	goOs := ""
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				vcsRevision = setting.Value
			case "vcs.modified":
				vcsMod = setting.Value
			case "GOARCH":
				goArch = setting.Value
			case "GOOS":
				goOs = setting.Value
			default:
			}
		}
	}
	params := url.Values{}
	if vcsMod == "true" {
		params.Set("m", "t")
	}
	params.Set("go", buildInfo.GoVersion)
	if goArch != "" {
		params.Set("a", goArch)
	}
	if goOs != "" {
		params.Set("os", goOs)
	}
	AptosClientHeaderValue = fmt.Sprintf("aptos-go-sdk/%s;%s", vcsRevision, params.Encode())
}

// Move some APT from sender to dest
// Amount in Octas (10^-8 APT)
//
// options may be: MaxGasAmount, GasUnitPrice, ExpirationSeconds, ValidUntil, SequenceNumber, ChainIdOption
func APTTransferTransaction(client *Client, sender *core.Account, dest core.AccountAddress, amount uint64, options ...any) (signedTxn *types.SignedTransaction, err error) {
	var amountBytes [8]byte
	binary.LittleEndian.PutUint64(amountBytes[:], amount)

	rawTxn, err := client.BuildTransaction(sender.Address,
		types.TransactionPayload{Payload: &types.EntryFunction{
			Module: types.ModuleId{
				Address: core.AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []types.TypeTag{},
			Args: [][]byte{
				dest[:],
				amountBytes[:],
			},
		}}, options...)
	if err != nil {
		return
	}
	signedTxn, err = rawTxn.Sign(sender)
	return
}
