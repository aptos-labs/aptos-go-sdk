package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/core"
	"log/slog"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk"
)

var (
	verbose    bool   = false
	accountStr string = ""
	network    string = ""
	txnHash    string = ""
)

func getenv(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

const APTOS_CLIENT_HEADER = "x-aptos-client"

var AptosClientHeaderValue = "aptos-go-sdk/unk"

func init() {
	vcsRevision := ""
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
	if vcsRevision != "" {
		AptosClientHeaderValue = fmt.Sprintf("aptos-go-sdk/%s", vcsRevision)
	}
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
}

func main() {

	args := os.Args[1:]
	var misc []string

	network = getenv("APTOS_NETWORK", network)
	nodeUrl := getenv("APTOS_NODE_URL", "")
	faucetUrl := getenv("APTOS_FAUCET_URL", "")

	// there may be better command frameworks, but in a pinch I can write what I want faster than I can learn one
	argi := 0
	for argi < len(args) {
		arg := args[argi]
		if arg == "-v" || arg == "--verbose" {
			verbose = true
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
		} else if arg == "-a" || arg == "--account" {
			accountStr = args[argi+1]
			argi++
		} else if arg == "-n" || arg == "--network" {
			network = args[argi+1]
			argi++
		} else if arg == "-u" || arg == "--node" {
			nodeUrl = args[argi+1]
			argi++
		} else if arg == "-F" || arg == "--faucet" {
			faucetUrl = args[argi+1]
			argi++
		} else if arg == "-t" || arg == "--txn" {
			txnHash = args[argi+1]
			argi++
		} else {
			misc = append(misc, arg)
		}
		argi++
	}

	var client *aptos.Client
	var err error

	if network != "" {
		client, err = aptos.NewClientFromNetworkName(network)
		maybefail(err, "client error: %s", err)
	} else {
		client, err = aptos.NewClient(aptos.NetworkConfig{
			NodeUrl:   nodeUrl,
			FaucetUrl: faucetUrl,
		})
		maybefail(err, "client error: %s", err)
	}

	var account core.AccountAddress
	if accountStr != "" {
		err := account.ParseStringRelaxed(accountStr)
		maybefail(err, "could not parse address: %s", err)
	}

	argi = 0
	for argi < len(misc) {
		arg := misc[argi]
		if arg == "account" {
			data, err := client.Account(account)
			maybefail(err, "could not get account %s: %s", accountStr, err)
			os.Stdout.WriteString(prettyJson(data))
		} else if arg == "account_resources" {
			localAccountStr := accountStr
			if (localAccountStr == "") && ((argi + 1) < len(misc)) {
				localAccountStr = misc[argi+1]
				err := account.ParseStringRelaxed(localAccountStr)
				maybefail(err, "could not parse address %#v: %s", localAccountStr, err)
				argi++
			}
			resources, err := client.AccountResources(account)
			maybefail(err, "could not get account resources %s: %s", localAccountStr, err)
			os.Stdout.WriteString(prettyJson(resources))
		} else if arg == "txn_by_hash" {
			data, err := client.TransactionByHash(txnHash)
			maybefail(err, "could not get txn %s: %s %s", txnHash, err, hebody(err))
			os.Stdout.WriteString(prettyJson(data))
		} else if arg == "info" {
			data, err := client.Info()
			maybefail(err, "could not get info: %s", err)
			os.Stdout.WriteString(prettyJson(data))
		} else if arg == "transactions" {
			exceptSystem := false

			if (argi+1 < len(misc)) && (misc[argi+1] == "--except-system") {
				// filter out system "block_metadata_transaction" and "state_checkpoint_transaction"
				exceptSystem = true
				argi++
			}
			data, err := client.Transactions(nil, nil)
			maybefail(err, "could not get info: %s", err)
			timestamps := make([]int64, 0, len(data))
			for _, rec := range data {
				if tsX, ok := rec["timestamp"]; ok {
					if tss, ok := tsX.(string); ok {
						tsv, err := strconv.ParseInt(tss, 10, 64)
						if err == nil {
							timestamps = append(timestamps, tsv)
						}
					}
				}
			}
			if len(timestamps) > 0 {
				mints := timestamps[0]
				maxts := timestamps[0]
				nowts := time.Now().UnixMicro()
				for _, t := range timestamps[1:] {
					if t < mints {
						mints = t
					}
					if t > maxts {
						maxts = t
					}
				}
				mindt := nowts - mints
				maxdt := nowts - maxts
				slog.Info("got txns", "len", len(data), "maxAge", float64(mindt)*0.000001, "minAge", float64(maxdt)*0.0000001)
			}
			if exceptSystem {
				nd := make([]map[string]any, 0, len(data))
				for _, rec := range data {
					if recType, ok := rec["type"]; ok {
						if recType == "state_checkpoint_transaction" || recType == "block_metadata_transaction" || recType == "validator_transaction" {
							continue
						}
					}
					nd = append(nd, rec)
				}
				slog.Debug("txns filtered", "orig", len(data), "kept", len(nd))
				data = nd
			}
			os.Stdout.WriteString(prettyJson(data))
		} else if arg == "naf" {
			alice, err := core.NewEd25519Account()
			maybefail(err, "new account: %s", err)
			amount := uint64(200_000_000)
			err = client.Fund(alice.Address, amount)
			maybefail(err, "faucet err: %s", err)
			fmt.Fprintf(os.Stdout, "new account %s funded for %d\n", alice.Address.String(), amount)

			bob, err := core.NewEd25519Account()
			maybefail(err, "new account: %s", err)
			//amount = uint64(10_000_000)
			err = client.Fund(bob.Address, amount)
			maybefail(err, "faucet err: %s", err)
			fmt.Fprintf(os.Stdout, "new account %s funded for %d\n", bob.Address.String(), amount)

			time.Sleep(2 * time.Second)
			stxn, err := aptos.APTTransferTransaction(client, alice, bob.Address, 42)
			maybefail(err, "could not make transfer txn, %s", err)
			slog.Debug("transfer", "stxn", stxn)
			submitStart := time.Now()
			result, err := client.SubmitTransaction(stxn)
			if err != nil {
				if he, ok := err.(*aptos.HttpError); ok {
					fmt.Fprintf(os.Stdout, "txn err:\n\t%s\n", string(he.Body))
				}
				maybefail(err, "could not submit transfer txn, %s", err)
			}
			fmt.Printf("submit txn result:\n%s\n", prettyJson(result))
			txnHashX, haveHash := result["hash"]
			var txnHash string
			if haveHash {
				txnHash, haveHash = txnHashX.(string)
			}
			if haveHash {
				txnInfo, err := client.WaitForTransaction(txnHash)
				dt := time.Now().Sub(submitStart)
				if err != nil {
					fmt.Printf("txn wait done in %.2f seconds, err=%s\n", dt.Seconds(), err)
				} else {
					fmt.Printf("txn wait done in %.2f seconds, success result:\n%s\n", dt.Seconds(), prettyJson(txnInfo))
				}
			}
			fmt.Printf("alice addr %s\n", alice.Address.String())
			fmt.Printf("bob   addr %s\n", bob.Address.String())
		} else if arg == "send" {
			// next three args: source addr, dest addr, amount
			var sender core.AccountAddress
			var dest core.AccountAddress
			var amount uint64
			err := sender.ParseStringRelaxed(misc[argi+1])
			maybefail(err, "bad sender, %s", err)
			err = dest.ParseStringRelaxed(misc[argi+2])
			maybefail(err, "bad dest, %s", err)
			amount, err = strconv.ParseUint(misc[argi+3], 10, 64)
			maybefail(err, "bad amount, %s", err)

			var sn uint64
			if getenv("DUMMY", "") == "" {
				info, err := client.Account(sender)
				maybefail(err, "could not get sender account info, %s", err)
				sn, err = info.SequenceNumber()
				maybefail(err, "bad sequence number, %s", err)
			} else {
				sn = 0
			}

			now := time.Now().Unix()

			var amountbytes [8]byte
			binary.LittleEndian.PutUint64(amountbytes[:], amount)
			txn := aptos.RawTransaction{
				Sender:         sender,
				SequenceNumber: sn + 1,
				Payload: aptos.TransactionPayload{Payload: &aptos.EntryFunction{
					Module: aptos.ModuleId{
						Address: core.AccountOne,
						Name:    "aptos_account",
					},
					Function: "transfer",
					// ArgTypes: []aptos.TypeTag{
					// 	aptos.TypeTag{Value: &aptos.AccountAddressTag{Value: dest}},
					// 	aptos.TypeTag{Value: &aptos.U64Tag{Value: amount}},
					// },
					ArgTypes: []aptos.TypeTag{},
					Args: [][]byte{
						dest[:],
						amountbytes[:],
					},
				}},
				MaxGasAmount:              1000,
				GasUnitPrice:              2000,
				ExpirationTimetampSeconds: uint64(now + 100),
				ChainId:                   4,
			}
			txnblob, err := txn.SignableBytes()
			maybefail(err, "txn SignableBytes, %s", err)
			//ser := aptos.Serializer{}
			//txn.MarshalBCS(&ser)
			//err = ser.Error()
			//maybefail(err, "txn BCS, %s", err)
			//txnblob := ser.ToBytes()
			enc := hex.NewEncoder(os.Stdout)
			enc.Write(txnblob)
			os.Stdout.WriteString("\n")
			argi += 3
		} else {
			fmt.Fprintf(os.Stderr, "bad action %#v", arg)
			os.Exit(1)
		}
		argi++
	}
}

func maybefail(err error, msg string, args ...any) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

func prettyJson(x any) string {
	out := strings.Builder{}
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "  ")
	enc.Encode(x)
	return out.String()
}

func hebody(err error) string {
	he, ok := err.(*aptos.HttpError)
	if !ok {
		return ""
	}
	return string(he.Body)
}
