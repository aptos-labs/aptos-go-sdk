// Package transaction provides transaction building utilities for the Aptos blockchain.
//
// This package offers a fluent builder API for constructing transactions:
//
//	txn, err := transaction.New().
//	    Sender(senderAddress).
//	    EntryFunction("0x1::coin::transfer", []aptos.TypeTag{aptosCoinType}, recipient, amount).
//	    MaxGas(2000).
//	    GasPrice(100).
//	    Expiration(30 * time.Second).
//	    Build(ctx, client)
//
// # Multi-Agent Transactions
//
// For multi-agent transactions with multiple signers:
//
//	txn, err := transaction.New().
//	    Sender(sender).
//	    SecondarySigners(addr1, addr2).
//	    EntryFunction(...).
//	    Build(ctx, client)
//
// # Fee-Payer Transactions
//
// For sponsored transactions:
//
//	txn, err := transaction.New().
//	    Sender(sender).
//	    FeePayer(sponsorAddress).
//	    EntryFunction(...).
//	    Build(ctx, client)
package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	aptos "github.com/aptos-labs/aptos-go-sdk/v2"
)

// Builder provides a fluent API for constructing transactions.
type Builder struct {
	// Required fields
	sender  *aptos.AccountAddress
	payload aptos.Payload

	// Optional transaction parameters
	maxGasAmount *uint64
	gasUnitPrice *uint64
	sequenceNum  *uint64
	expiration   *time.Duration

	// Multi-agent/fee-payer fields
	secondarySigners []aptos.AccountAddress
	feePayer         *aptos.AccountAddress

	// Build options
	estimateGas          bool
	prioritizedGas       bool
	simulateBeforeSubmit bool

	// Errors accumulated during building
	errs []error
}

// New creates a new transaction builder.
func New() *Builder {
	return &Builder{}
}

// Sender sets the transaction sender address.
// This is required for all transactions.
func (b *Builder) Sender(addr aptos.AccountAddress) *Builder {
	b.sender = &addr
	return b
}

// EntryFunction sets the payload to an entry function call.
// Module format: "address::module_name::function_name" or use ModuleID.
func (b *Builder) EntryFunction(module string, typeArgs []aptos.TypeTag, args ...any) *Builder {
	moduleID, function, err := parseModuleFunction(module)
	if err != nil {
		b.errs = append(b.errs, fmt.Errorf("invalid module function: %w", err))
		return b
	}

	b.payload = &aptos.EntryFunctionPayload{
		Module:   *moduleID,
		Function: function,
		TypeArgs: typeArgs,
		Args:     args,
	}
	return b
}

// EntryFunctionWithModuleID sets the payload using a ModuleID.
func (b *Builder) EntryFunctionWithModuleID(module aptos.ModuleID, function string, typeArgs []aptos.TypeTag, args ...any) *Builder {
	b.payload = &aptos.EntryFunctionPayload{
		Module:   module,
		Function: function,
		TypeArgs: typeArgs,
		Args:     args,
	}
	return b
}

// Script sets the payload to a script execution.
func (b *Builder) Script(bytecode []byte, typeArgs []aptos.TypeTag, args ...any) *Builder {
	b.payload = &aptos.ScriptPayload{
		Code:     bytecode,
		TypeArgs: typeArgs,
		Args:     args,
	}
	return b
}

// Payload sets a custom payload.
func (b *Builder) Payload(payload aptos.Payload) *Builder {
	b.payload = payload
	return b
}

// MaxGas sets the maximum gas amount for the transaction.
// Default is 200,000 gas units.
func (b *Builder) MaxGas(amount uint64) *Builder {
	b.maxGasAmount = &amount
	return b
}

// GasPrice sets the gas unit price in octas.
// Default uses the network's suggested gas price.
func (b *Builder) GasPrice(price uint64) *Builder {
	b.gasUnitPrice = &price
	return b
}

// SequenceNumber sets a specific sequence number.
// If not set, the sequence number is fetched from the network.
func (b *Builder) SequenceNumber(seq uint64) *Builder {
	b.sequenceNum = &seq
	return b
}

// Expiration sets how long the transaction is valid.
// Default is 30 seconds.
func (b *Builder) Expiration(d time.Duration) *Builder {
	b.expiration = &d
	return b
}

// ExpirationTimestamp sets an absolute expiration timestamp.
func (b *Builder) ExpirationTimestamp(ts time.Time) *Builder {
	d := time.Until(ts)
	if d < 0 {
		b.errs = append(b.errs, errors.New("expiration timestamp is in the past"))
		return b
	}
	b.expiration = &d
	return b
}

// SecondarySigners adds secondary signers for multi-agent transactions.
func (b *Builder) SecondarySigners(addrs ...aptos.AccountAddress) *Builder {
	b.secondarySigners = append(b.secondarySigners, addrs...)
	return b
}

// FeePayer sets a fee payer for sponsored transactions.
func (b *Builder) FeePayer(addr aptos.AccountAddress) *Builder {
	b.feePayer = &addr
	return b
}

// EstimateGas enables automatic gas estimation.
func (b *Builder) EstimateGas() *Builder {
	b.estimateGas = true
	return b
}

// PrioritizedGas uses prioritized gas estimation for faster inclusion.
func (b *Builder) PrioritizedGas() *Builder {
	b.estimateGas = true
	b.prioritizedGas = true
	return b
}

// SimulateFirst simulates the transaction before submission.
func (b *Builder) SimulateFirst() *Builder {
	b.simulateBeforeSubmit = true
	return b
}

// Validate checks if the builder has all required fields.
func (b *Builder) Validate() error {
	if len(b.errs) > 0 {
		return errors.Join(b.errs...)
	}
	if b.sender == nil {
		return errors.New("sender is required")
	}
	if b.payload == nil {
		return errors.New("payload is required")
	}
	return nil
}

// Build constructs a RawTransaction from the builder configuration.
// It fetches required data (sequence number, chain ID) from the client if not provided.
func (b *Builder) Build(ctx context.Context, client aptos.Client) (*aptos.RawTransaction, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

	// Build transaction options
	var opts []aptos.TransactionOption

	if b.maxGasAmount != nil {
		opts = append(opts, aptos.WithMaxGas(*b.maxGasAmount))
	}

	if b.gasUnitPrice != nil {
		opts = append(opts, aptos.WithGasPrice(*b.gasUnitPrice))
	}

	if b.sequenceNum != nil {
		opts = append(opts, aptos.WithSequenceNumber(*b.sequenceNum))
	}

	if b.expiration != nil {
		opts = append(opts, aptos.WithExpiration(*b.expiration))
	}

	if b.estimateGas {
		if b.prioritizedGas {
			opts = append(opts, aptos.WithPrioritizedGas())
		} else {
			opts = append(opts, aptos.WithGasEstimation())
		}
	}

	if b.simulateBeforeSubmit {
		opts = append(opts, aptos.WithSimulation())
	}

	return client.BuildTransaction(ctx, *b.sender, b.payload, opts...)
}

// Options returns TransactionOptions derived from this builder's configuration.
// Useful when you want to use the builder's configuration with client.BuildTransaction directly.
func (b *Builder) Options() []aptos.TransactionOption {
	var opts []aptos.TransactionOption

	if b.maxGasAmount != nil {
		opts = append(opts, aptos.WithMaxGas(*b.maxGasAmount))
	}
	if b.gasUnitPrice != nil {
		opts = append(opts, aptos.WithGasPrice(*b.gasUnitPrice))
	}
	if b.sequenceNum != nil {
		opts = append(opts, aptos.WithSequenceNumber(*b.sequenceNum))
	}
	if b.expiration != nil {
		opts = append(opts, aptos.WithExpiration(*b.expiration))
	}
	if b.estimateGas {
		if b.prioritizedGas {
			opts = append(opts, aptos.WithPrioritizedGas())
		} else {
			opts = append(opts, aptos.WithGasEstimation())
		}
	}
	if b.simulateBeforeSubmit {
		opts = append(opts, aptos.WithSimulation())
	}

	return opts
}

// IsMultiAgent returns true if this is a multi-agent transaction.
func (b *Builder) IsMultiAgent() bool {
	return len(b.secondarySigners) > 0
}

// IsFeePayer returns true if this is a fee-payer transaction.
func (b *Builder) IsFeePayer() bool {
	return b.feePayer != nil
}

// Helper function to parse "address::module::function" format.
func parseModuleFunction(s string) (*aptos.ModuleID, string, error) {
	// Find the last "::" separator
	lastIdx := -1
	for i := len(s) - 2; i >= 0; i-- {
		if s[i] == ':' && s[i+1] == ':' {
			lastIdx = i
			break
		}
	}

	if lastIdx == -1 {
		return nil, "", errors.New("invalid format: expected 'address::module::function'")
	}

	modulePart := s[:lastIdx]
	function := s[lastIdx+2:]

	// Parse the module part
	midIdx := -1
	for i := len(modulePart) - 2; i >= 0; i-- {
		if modulePart[i] == ':' && modulePart[i+1] == ':' {
			midIdx = i
			break
		}
	}

	if midIdx == -1 {
		return nil, "", errors.New("invalid format: expected 'address::module::function'")
	}

	addrStr := modulePart[:midIdx]
	moduleName := modulePart[midIdx+2:]

	addr, err := aptos.ParseAddress(addrStr)
	if err != nil {
		return nil, "", fmt.Errorf("invalid address: %w", err)
	}

	return &aptos.ModuleID{Address: addr, Name: moduleName}, function, nil
}

// Common transaction builders

// TransferAPT creates a builder for an APT transfer transaction.
func TransferAPT(from, to aptos.AccountAddress, amount uint64) *Builder {
	return New().
		Sender(from).
		EntryFunctionWithModuleID(
			aptos.ModuleID{Address: aptos.AccountOne, Name: "aptos_account"},
			"transfer",
			nil, // No type arguments
			to.Bytes(), amount,
		)
}

// TransferCoins creates a builder for a coin transfer transaction.
func TransferCoins(from, to aptos.AccountAddress, coinType aptos.TypeTag, amount uint64) *Builder {
	return New().
		Sender(from).
		EntryFunctionWithModuleID(
			aptos.ModuleID{Address: aptos.AccountOne, Name: "coin"},
			"transfer",
			[]aptos.TypeTag{coinType},
			to.Bytes(), amount,
		)
}
