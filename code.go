package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// PublishPackagePayloadFromJsonFile publishes code created with the Aptos CLI to publish with it
// you must run the command `aptos move build-publish-payload`
func PublishPackagePayloadFromJsonFile(metadata []byte, bytecode [][]byte) (*TransactionPayload, error) {
	metadataBytes, err := bcs.SerializeBytes(metadata)
	if err != nil {
		return nil, err
	}

	// TODO: Arrays should be easier than this
	serializer := &bcs.Serializer{}
	serializer.Uleb128(uint32(len(bytecode)))
	for _, b := range bytecode {
		serializer.WriteBytes(b)
	}
	bytecodeBytes := serializer.ToBytes()

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
