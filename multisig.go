package aptos

import "github.com/aptos-labs/aptos-go-sdk/bcs"

// FetchNextMultisigAddress retrieves the next multisig address to be created from the given account
func (client *Client) FetchNextMultisigAddress(address AccountAddress) (*AccountAddress, error) {
	viewResponse, err := client.View(&ViewPayload{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "multisig_account",
		},
		Function: "get_next_multisig_account_address",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{address[:]},
	})
	if err != nil {
		return nil, err
	}
	multisigAddress := &AccountAddress{}
	err = multisigAddress.ParseStringRelaxed(viewResponse[0].(string))
	if err != nil {
		return nil, err
	}

	return multisigAddress, nil
}

// -- Multisig payloads --

// MultisigCreateAccountPayload creates a payload for setting up a multisig
//
// Required signers must be between 1 and the number of addresses total (sender + additional addresses).
// Metadata values must be BCS encoded values
func MultisigCreateAccountPayload(requiredSigners uint64, additionalAddresses []AccountAddress, metadataKeys []string, metadataValues []byte) (*EntryFunction, error) {
	// Serialize arguments
	additionalOwners, err := bcs.SerializeSequenceOnly(additionalAddresses)
	if err != nil {
		return nil, err
	}

	requiredSignersBytes, err := bcs.SerializeU64(requiredSigners)
	if err != nil {
		return nil, err
	}

	// TODO: This is a little better than before, but maybe we make some of these ahead of time
	metadataKeysBytes, err := bcs.SerializeSingle(func(ser *bcs.Serializer) {
		bcs.SerializeSequenceWithFunction(metadataKeys, ser, func(ser *bcs.Serializer, item string) {
			ser.WriteString(item)
		})
	})
	if err != nil {
		return nil, err
	}

	return &EntryFunction{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "multisig_account",
		},
		Function: "create_with_owners",
		ArgTypes: []TypeTag{},
		Args: [][]byte{
			additionalOwners,     // Addresses of the other 2 in the 3 owners
			requiredSignersBytes, // The number of required signatures 2-of-3
			metadataKeysBytes,    // Metadata keys for any metadata you want to add to the account
			metadataValues,       // Values for the metadata added, must be BCS encoded
		},
	}, nil
}

// MultisigAddOwnerPayload creates a payload to add an owner from the multisig
func MultisigAddOwnerPayload(owner AccountAddress) *EntryFunction {
	return multisigOwnerPayloadCommon("add_owner", owner)
}

// MultisigRemoveOwnerPayload creates a payload to remove an owner from the multisig
func MultisigRemoveOwnerPayload(owner AccountAddress) *EntryFunction {
	return multisigOwnerPayloadCommon("remove_owner", owner)
}

// MultisigChangeThresholdPayload creates a payload to change the number of signatures required for a transaction to pass.
//
// For example, changing a 2-of-3 to a 3-of-3, the value for numSignaturesRequired would be 3
func MultisigChangeThresholdPayload(numSignaturesRequired uint64) (*EntryFunction, error) {
	thresholdBytes, err := bcs.SerializeU64(numSignaturesRequired)
	if err != nil {
		return nil, err
	}
	return &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "multisig_account"},
		Function: "update_signatures_required",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{thresholdBytes},
	}, nil
}

// MultisigCreateTransactionPayload creates a transaction to be voted upon in an on-chain multisig
//
// Note, this serializes an EntryFunction payload, and sends it as an argument in the transaction.  If the
// entry function payload is large, use MultisigCreateTransactionPayloadWithHash.  The advantage of this over the
// hash version, is visibility on-chain.
func MultisigCreateTransactionPayload(multisigAddress AccountAddress, payload *MultisigTransactionPayload) (*EntryFunction, error) {
	payloadBytes, err := bcs.Serialize(payload)
	if err != nil {
		return nil, err
	}
	// Serialize and add the number of bytes in front
	payloadBytes2, err := bcs.SerializeBytes(payloadBytes)
	if err != nil {
		return nil, err
	}
	return multisigTransactionCommon("create_transaction", multisigAddress, [][]byte{payloadBytes2}), nil
}

// MultisigCreateTransactionPayloadWithHash creates a transaction to be voted upon in an on-chain multisig
//
// This differs from MultisigCreateTransactionPayload by instead taking a SHA3-256 hash of the payload and using that as
// the identifier of the transaction.  The transaction intent will not be stored on-chain, only the hash of it.
func MultisigCreateTransactionPayloadWithHash(multisigAddress AccountAddress, payload *MultisigTransactionPayload) (*EntryFunction, error) {
	payloadBytes, err := bcs.Serialize(payload)
	if err != nil {
		return nil, err
	}
	hash := Sha3256Hash([][]byte{payloadBytes})

	// Serialize and add the number of bytes in front
	hashBytes, err := bcs.SerializeBytes(hash)
	if err != nil {
		return nil, err
	}
	return multisigTransactionCommon("create_transaction_with_hash", multisigAddress, [][]byte{hashBytes}), nil
}

// MultisigApprovePayload generates a payload for approving a transaction on-chain.  The caller must be an owner of the
// multisig
func MultisigApprovePayload(multisigAddress AccountAddress, transactionId uint64) (*EntryFunction, error) {
	return multisigTransactionWithTransactionIdCommon("approve_transaction", multisigAddress, transactionId)
}

// MultisigRejectPayload generates a payload for rejecting a transaction on-chain.  The caller must be an owner of the
// multisig
func MultisigRejectPayload(multisigAddress AccountAddress, transactionId uint64) (*EntryFunction, error) {
	return multisigTransactionWithTransactionIdCommon("reject_transaction", multisigAddress, transactionId)
}

// multisigTransactionWithTransactionIdCommon is a helper for functions that take TransactionId
func multisigTransactionWithTransactionIdCommon(functionName string, multisigAddress AccountAddress, transactionId uint64) (*EntryFunction, error) {
	transactionIdBytes, err := bcs.SerializeU64(transactionId)
	if err != nil {
		return nil, err
	}
	return multisigTransactionCommon(functionName, multisigAddress, [][]byte{transactionIdBytes}), nil
}

// multisigOwnerPayloadCommon is a helper for owner based multisig operations
func multisigTransactionCommon(functionName string, multisigAddress AccountAddress, additionalArgs [][]byte) *EntryFunction {
	return &EntryFunction{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "multisig_account",
		},
		Function: functionName,
		ArgTypes: []TypeTag{},
		Args:     append([][]byte{multisigAddress[:]}, additionalArgs...),
	}
}

// multisigOwnerPayloadCommon is a helper for owner based multisig operations
func multisigOwnerPayloadCommon(functionName string, owner AccountAddress) *EntryFunction {
	return &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "multisig_account"},
		Function: functionName,
		ArgTypes: []TypeTag{},
		Args:     [][]byte{owner[:]},
	}
}
