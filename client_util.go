package aptos

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"net/url"
	"runtime/debug"
)

const ClientHeader = "x-aptos-client"

var ClientHeaderValue = "aptos-go-sdk/unk"

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
	ClientHeaderValue = fmt.Sprintf("aptos-go-sdk/%s;%s", vcsRevision, params.Encode())
}

// APTTransferTransaction Move some APT from sender to dest
// Amount in Octas (10^-8 APT)
//
// TODO: This has to be reworked to deal with fee payer and other methods, it will likely go away
//
// options may be: MaxGasAmount, GasUnitPrice, ExpirationSeconds, ValidUntil, SequenceNumber, ChainIdOption
func APTTransferTransaction(client *Client, sender *Account, dest AccountAddress, amount uint64, options ...any) (signedTxn *SignedTransaction, err error) {
	amountBytes, err := bcs.SerializeU64(amount)
	if err != nil {
		return nil, err
	}

	rawTxn, err := client.BuildTransaction(sender.Address,
		TransactionPayload{Payload: &EntryFunction{
			Module: ModuleId{
				Address: AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []TypeTag{},
			Args: [][]byte{
				dest[:],
				amountBytes,
			},
		}}, options...)
	if err != nil {
		return
	}
	return rawTxn.SignedTransaction(sender)
}
