package aptos

import (
	"fmt"
	"net/url"
	"runtime/debug"
)

// ClientHeader is the header key for the SDK version
const ClientHeader = "X-Aptos-Client"

// ClientHeaderValue is the header value for the SDK version
var ClientHeaderValue = "aptos-go-sdk/unk"

// Sets up the ClientHeaderValue with the SDK version
func init() {
	vcsRevision := "unk"
	vcsMod := ""
	goArch := ""
	goOs := ""
	params := url.Values{}
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		params.Set("go", buildInfo.GoVersion)
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
	if vcsMod == "true" {
		params.Set("m", "t")
	}
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
func APTTransferTransaction(client *Client, sender TransactionSigner, dest AccountAddress, amount uint64, options ...any) (*RawTransaction, error) {
	entryFunction, err := CoinTransferPayload(nil, dest, amount)
	if err != nil {
		return nil, err
	}

	return client.BuildTransaction(sender.AccountAddress(),
		TransactionPayload{Payload: entryFunction}, options...)
}
