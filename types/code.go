package types

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// PublishPackagePayloadFromJsonFile publishes code created with the Aptos CLI to publish with it.
// The Aptos CLI can generate the associated file with the following CLI command:
//
//	aptos move build-publish-payload
//
// metadata must be the BCS encoded metadata from the compiler, and bytecode must be the BCS encoded bytecode from the compiler.
// bytecode must be ordered in the same order out of the compiler, or it will fail on publishing.
func PublishPackagePayloadFromJsonFile(metadata []byte, bytecode [][]byte) (*TransactionPayload, error) {
	metadataBytes, err := bcs.SerializeBytes(metadata)
	if err != nil {
		return nil, err
	}

	bytecodeBytes, err := bcs.SerializeSingle(func(ser *bcs.Serializer) {
		bcs.SerializeSequenceWithFunction(bytecode, ser, (*bcs.Serializer).WriteBytes)
	})
	if err != nil {
		return nil, err
	}

	return &TransactionPayload{Payload: &EntryFunction{
		Module: ModuleId{
			Address: AccountOne,
			Name:    "code",
		},
		Function: "publish_package_txn",
		ArgTypes: []TypeTag{},
		Args:     [][]byte{metadataBytes, bytecodeBytes},
	}}, nil
}
