package aptos

import (
	"fmt"
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

// APTTransferTransaction Move some APT from sender to dest, only for single signer
// Amount in Octas (10^-8 APT)
//
// options may be: MaxGasAmount, GasUnitPrice, ExpirationSeconds, ValidUntil, SequenceNumber, ChainIdOption
// deprecated, please use the EntryFunction APIs
func APTTransferTransaction(client *Client, sender TransactionSigner, dest AccountAddress, amount uint64, options ...any) (rawTxn *RawTransaction, err error) {
	entryFunction, err := CoinTransferPayload(nil, dest, amount)
	if err != nil {
		return nil, err
	}

	rawTxn, err = client.BuildTransaction(sender.AccountAddress(),
		TransactionPayload{Payload: entryFunction}, options...)
	return
}
