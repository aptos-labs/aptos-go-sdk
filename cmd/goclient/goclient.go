package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	aptos "github.com/aptos-labs/aptos-go-sdk"
)

var (
	verbose    bool   = false
	accountStr string = ""
	baseUrl    string = ""
)

func main() {

	args := os.Args[1:]
	var misc []string

	// there may be better command frameworks, but in a pinch I can write what I want faster than I can learn one
	argi := 0
	for argi < len(args) {
		arg := args[argi]
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else if arg == "-a" || arg == "--account" {
			accountStr = args[argi+1]
			argi++
		} else if arg == "-u" || arg == "--url" {
			baseUrl = args[argi+1]
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

	client, err := aptos.NewClient(baseUrl)
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
		} else if arg == "info" {
			data, err := client.Info()
			maybefail(err, "could not get info: %s", err)
			os.Stdout.WriteString(prettyJson(data))
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
