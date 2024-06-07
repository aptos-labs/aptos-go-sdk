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

func CreateMultisigAccountPayload(requiredSigners uint64, additionalAddresses []AccountAddress, metadataKeys []string, metadataValues []byte) (*EntryFunction, error) {
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
		}}, nil
}

func MultisigCreateAddOwnerTransaction(owner AccountAddress) *EntryFunction {
	return multisigOwnerPayloadCommon("add_owner", owner)
}

func MultisigCreateRemoveOwnerTransaction(owner AccountAddress) *EntryFunction {
	return multisigOwnerPayloadCommon("remove_owner", owner)
}

func MultisigCreateChangeThresholdTransaction(numSignaturesRequired uint64) (*EntryFunction, error) {
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

func MultisigCreateTransactionWithPayload(multisigAddress AccountAddress, payload *MultisigTransactionPayload) (*EntryFunction, error) {
	payloadBytes, err := bcs.Serialize(payload)
	if err != nil {
		return nil, err
	}
	// Serialize and add the number of bytes in front
	payloadBytes2, err := bcs.SerializeBytes(payloadBytes)
	return multisigTransactionCommon("create_transaction", multisigAddress, [][]byte{payloadBytes2}), nil
}

func MultisigCreateTransactionWithHash(multisigAddress AccountAddress, payload *MultisigTransactionPayload) (*EntryFunction, error) {
	payloadBytes, err := bcs.Serialize(payload)
	if err != nil {
		return nil, err
	}
	hash := Sha3256Hash([][]byte{payloadBytes})

	// Serialize and add the number of bytes in front
	hashBytes, err := bcs.SerializeBytes(hash)
	return multisigTransactionCommon("create_transaction_with_hash", multisigAddress, [][]byte{hashBytes}), nil
}

func MultisigApproveTransaction(multisigAddress AccountAddress, transactionId uint64) (*EntryFunction, error) {
	return multisigTransactionWithTransactionIdCommon("approve_transaction", multisigAddress, transactionId)
}

func MultisigRejectTransaction(multisigAddress AccountAddress, transactionId uint64) (*EntryFunction, error) {
	return multisigTransactionWithTransactionIdCommon("reject_transaction", multisigAddress, transactionId)
}

func multisigTransactionWithTransactionIdCommon(functionName string, multisigAddress AccountAddress, transactionId uint64) (*EntryFunction, error) {
	transactionIdBytes, err := bcs.SerializeU64(transactionId)
	if err != nil {
		return nil, err
	}
	return multisigTransactionCommon(functionName, multisigAddress, [][]byte{transactionIdBytes}), nil
}

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

func multisigOwnerPayloadCommon(functionName string, owner AccountAddress) *EntryFunction {
	return &EntryFunction{
		Module:   ModuleId{Address: AccountOne, Name: "multisig_account"},
		Function: functionName,
		ArgTypes: []TypeTag{},
		Args:     [][]byte{owner[:]},
	}
}
