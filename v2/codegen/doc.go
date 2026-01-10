// Package codegen provides code generation utilities for generating type-safe Go bindings
// from Move module ABIs.
//
// The code generator can produce:
//   - Go structs from Move structs
//   - Type-safe entry function wrappers
//   - View function helpers
//
// # Usage
//
// ABIs can be fetched from an Aptos node or loaded from local JSON files:
//
//	// From chain
//	client, _ := aptos.NewClient(aptos.DevnetConfig)
//	abi, _ := client.AccountModule(ctx, aptos.AccountOne, "coin")
//	code, _ := codegen.GenerateModule(abi.Abi, codegen.Options{PackageName: "coin"})
//
//	// From local file
//	data, _ := os.ReadFile("coin_abi.json")
//	var module api.MoveModule
//	json.Unmarshal(data, &module)
//	code, _ := codegen.GenerateModule(&module, codegen.Options{PackageName: "coin"})
package codegen
