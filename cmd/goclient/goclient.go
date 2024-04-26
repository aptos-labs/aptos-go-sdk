package main

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
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
	nodeUrl    string = "https://api.devnet.aptoslabs.com/v1"
	faucetUrl  string = "https://faucet.devnet.aptoslabs.com"
	txnHash    string = ""
)

func getenv(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {

	args := os.Args[1:]
	var misc []string

	nodeUrl = getenv("APTOS_NODE_URL", nodeUrl)
	faucetUrl = getenv("APTOS_FAUCET_URL", faucetUrl)

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
		} else if arg == "-u" || arg == "--url" {
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

	// TODO: some of this info will be useful for putting in client HTTP headers
	if verbose {
		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			fmt.Printf("built with %s\n", buildInfo.GoVersion)
			for _, setting := range buildInfo.Settings {
				fmt.Printf("%s=%s\n", setting.Key, setting.Value)
			}
		}
	}

	client, err := aptos.NewClient(nodeUrl)
	maybefail(err, "client error: %s", err)

	var account aptos.AccountAddress
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
			resources, err := client.AccountResources(account)
			maybefail(err, "could not get account resources %s: %s", accountStr, err)
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
			data, err := client.Transactions(nil, nil)
			maybefail(err, "could not get info: %s", err)
			os.Stdout.WriteString(prettyJson(data))
		} else if arg == "naf" {
			account, err := aptos.NewAccount()
			maybefail(err, "new account: %s", err)
			amount := uint64(200_000_000)
			err = aptos.FundAccount(client, faucetUrl, account.Address, amount)
			maybefail(err, "faucet err: %s", err)
			fmt.Fprintf(os.Stdout, "new account %s funded for %d, privkey = %s\n", account.Address.String(), amount, hex.EncodeToString(account.PrivateKey.(ed25519.PrivateKey)[:]))

			bob, err := aptos.NewAccount()
			maybefail(err, "new account: %s", err)
			//amount = uint64(10_000_000)
			err = aptos.FundAccount(client, faucetUrl, bob.Address, amount)
			maybefail(err, "faucet err: %s", err)
			fmt.Fprintf(os.Stdout, "new account %s funded for %d, privkey = %s\n", bob.Address.String(), amount, hex.EncodeToString(bob.PrivateKey.(ed25519.PrivateKey)[:]))

			//time.Sleep(2 * time.Second)
			stxn, err := aptos.TransferTransaction(client, account, bob.Address, 42)
			maybefail(err, "could not make transfer txn, %s", err)
			result, err := client.SubmitTransaction(stxn)
			if err != nil {
				if he, ok := err.(*aptos.HttpError); ok {
					fmt.Fprintf(os.Stdout, "txn err:\n\t%s\n", string(he.Body))
				}
				maybefail(err, "could not submit transfer txn, %s", err)
			}
			fmt.Printf("submit txn result: %#v", string(result))

		} else if arg == "send" {
			// next three args: source addr, dest addr, amount
			var sender aptos.AccountAddress
			var dest aptos.AccountAddress
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
						Address: aptos.Account0x1,
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
