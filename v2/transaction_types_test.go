package aptos

import (
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRawTransaction_BCSRoundTrip(t *testing.T) {
	// Create a sample raw transaction
	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             42,
		MaxGasAmount:               200000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: []TypeTag{AptosCoinTypeTag},
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	// Serialize
	data, err := bcs.Serialize(txn)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize
	txn2 := &RawTransaction{}
	err = bcs.Deserialize(txn2, data)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, txn.Sender, txn2.Sender)
	assert.Equal(t, txn.SequenceNumber, txn2.SequenceNumber)
	assert.Equal(t, txn.MaxGasAmount, txn2.MaxGasAmount)
	assert.Equal(t, txn.GasUnitPrice, txn2.GasUnitPrice)
	assert.Equal(t, txn.ExpirationTimestampSeconds, txn2.ExpirationTimestampSeconds)
	assert.Equal(t, txn.ChainID, txn2.ChainID)
}

func TestRawTransaction_SigningMessage(t *testing.T) {
	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	msg, err := txn.SigningMessage()
	require.NoError(t, err)
	assert.NotEmpty(t, msg)
	// The signing message should start with the prehash (32 bytes)
	assert.GreaterOrEqual(t, len(msg), 32)
}

func TestSignedTransaction_Hash(t *testing.T) {
	// Create a signed transaction
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := SignTransaction(key, txn)
	require.NoError(t, err)

	hash, err := signed.Hash()
	require.NoError(t, err)
	assert.True(t, len(hash) > 2)
	assert.Equal(t, "0x", hash[:2])
}

func TestEntryFunctionPayload_BCSRoundTrip(t *testing.T) {
	payload := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "aptos_account"},
		Function: "transfer",
		TypeArgs: []TypeTag{AptosCoinTypeTag},
		Args:     []any{AccountTwo[:], uint64(1000)},
	}

	// Serialize via serializePayload
	ser := bcs.NewSerializer()
	serializePayload(ser, payload)
	require.NoError(t, ser.Error())
	data := ser.ToBytes()

	// Deserialize
	des := bcs.NewDeserializer(data)
	payload2 := deserializePayload(des)
	require.NoError(t, des.Error())

	ef, ok := payload2.(*EntryFunctionPayload)
	require.True(t, ok)
	assert.Equal(t, payload.Module.Address, ef.Module.Address)
	assert.Equal(t, payload.Module.Name, ef.Module.Name)
	assert.Equal(t, payload.Function, ef.Function)
	assert.Len(t, ef.TypeArgs, 1)
}

func TestScriptPayload_BCSRoundTrip(t *testing.T) {
	payload := &ScriptPayload{
		Code:     []byte{0x01, 0x02, 0x03, 0x04},
		TypeArgs: nil,
		Args:     []any{uint64(100), true, AccountOne},
	}

	// Serialize
	ser := bcs.NewSerializer()
	serializePayload(ser, payload)
	require.NoError(t, ser.Error())
	data := ser.ToBytes()

	// Deserialize
	des := bcs.NewDeserializer(data)
	payload2 := deserializePayload(des)
	require.NoError(t, des.Error())

	sp, ok := payload2.(*ScriptPayload)
	require.True(t, ok)
	assert.Equal(t, payload.Code, sp.Code)
}

func TestSerializeArg_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		arg  any
	}{
		{"bytes", []byte{1, 2, 3}},
		{"string", "hello"},
		{"uint8", uint8(255)},
		{"uint16", uint16(65535)},
		{"uint32", uint32(4294967295)},
		{"uint64", uint64(18446744073709551615)},
		{"bool", true},
		{"address", AccountOne},
		{"address_ptr", &AccountOne},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := serializeArg(tt.arg)
			require.NoError(t, err)
			assert.NotEmpty(t, data)
		})
	}
}

func TestSerializeArg_NilAddress(t *testing.T) {
	var nilAddr *AccountAddress
	_, err := serializeArg(nilAddr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil address")
}

func TestSerializeArg_UnsupportedType(t *testing.T) {
	_, err := serializeArg(struct{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported argument type")
}

func TestScriptArg_BCSRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		arg  any
	}{
		{"bool_true", true},
		{"bool_false", false},
		{"u8", uint8(123)},
		{"u16", uint16(12345)},
		{"u32", uint32(1234567)},
		{"u64", uint64(1234567890123)},
		{"bytes", []byte{1, 2, 3, 4, 5}},
		{"address", AccountOne},
		{"address_ptr", &AccountTwo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := bcs.NewSerializer()
			serializeScriptArg(ser, tt.arg)
			require.NoError(t, ser.Error())

			des := bcs.NewDeserializer(ser.ToBytes())
			result := deserializeScriptArg(des)
			require.NoError(t, des.Error())
			assert.NotNil(t, result)
		})
	}
}

func TestScriptArg_UnsupportedType(t *testing.T) {
	ser := bcs.NewSerializer()
	serializeScriptArg(ser, struct{}{})
	require.Error(t, ser.Error())
}

func TestSingleSenderAuthenticator_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	singleAuth := &SingleSenderAuthenticator{Sender: auth}

	// Serialize
	data, err := bcs.Serialize(singleAuth)
	require.NoError(t, err)

	// Deserialize via deserializeTransactionAuthenticator
	des := bcs.NewDeserializer(data)
	txnAuth := deserializeTransactionAuthenticator(des)
	require.NoError(t, des.Error())

	_, ok := txnAuth.(*SingleSenderAuthenticator)
	assert.True(t, ok)
}

func TestSingleSenderAuthenticator_Verify(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	singleAuth := &SingleSenderAuthenticator{Sender: auth}
	assert.True(t, singleAuth.Verify(msg))
	assert.False(t, singleAuth.Verify([]byte("wrong message")))
}

func TestEd25519TransactionAuthenticator_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	ed25519Auth := &Ed25519TransactionAuthenticator{Sender: auth}

	// Serialize
	data, err := bcs.Serialize(ed25519Auth)
	require.NoError(t, err)

	// Deserialize
	des := bcs.NewDeserializer(data)
	txnAuth := deserializeTransactionAuthenticator(des)
	require.NoError(t, des.Error())

	_, ok := txnAuth.(*Ed25519TransactionAuthenticator)
	assert.True(t, ok)
}

func TestEd25519TransactionAuthenticator_Verify(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	ed25519Auth := &Ed25519TransactionAuthenticator{Sender: auth}
	assert.True(t, ed25519Auth.Verify(msg))
}

func TestMultiAgentAuthenticator_BCSRoundTrip(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	multiAuth := &MultiAgentAuthenticator{
		Sender:                   auth1,
		SecondarySignerAddresses: []AccountAddress{AccountTwo},
		SecondarySigners:         []*AccountAuthenticator{auth2},
	}

	// Serialize
	data, err := bcs.Serialize(multiAuth)
	require.NoError(t, err)

	// Deserialize
	des := bcs.NewDeserializer(data)
	txnAuth := deserializeTransactionAuthenticator(des)
	require.NoError(t, des.Error())

	result, ok := txnAuth.(*MultiAgentAuthenticator)
	require.True(t, ok)
	assert.NotNil(t, result.Sender)
	assert.Len(t, result.SecondarySignerAddresses, 1)
	assert.Len(t, result.SecondarySigners, 1)
}

func TestMultiAgentAuthenticator_Verify(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	multiAuth := &MultiAgentAuthenticator{
		Sender:                   auth1,
		SecondarySignerAddresses: []AccountAddress{AccountTwo},
		SecondarySigners:         []*AccountAuthenticator{auth2},
	}

	assert.True(t, multiAuth.Verify(msg))
	assert.False(t, multiAuth.Verify([]byte("wrong message")))
}

func TestFeePayerAuthenticator_BCSRoundTrip(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	feePayerAuth := &FeePayerAuthenticator{
		Sender:                   auth1,
		SecondarySignerAddresses: []AccountAddress{},
		SecondarySigners:         []*AccountAuthenticator{},
		FeePayerAddress:          AccountThree,
		FeePayerAuth:             auth2,
	}

	// Serialize
	data, err := bcs.Serialize(feePayerAuth)
	require.NoError(t, err)

	// Deserialize
	des := bcs.NewDeserializer(data)
	txnAuth := deserializeTransactionAuthenticator(des)
	require.NoError(t, des.Error())

	result, ok := txnAuth.(*FeePayerAuthenticator)
	require.True(t, ok)
	assert.NotNil(t, result.Sender)
	assert.Equal(t, AccountThree, result.FeePayerAddress)
	assert.NotNil(t, result.FeePayerAuth)
}

func TestFeePayerAuthenticator_Verify(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	feePayerAuth := &FeePayerAuthenticator{
		Sender:                   auth1,
		SecondarySignerAddresses: []AccountAddress{},
		SecondarySigners:         []*AccountAuthenticator{},
		FeePayerAddress:          AccountThree,
		FeePayerAuth:             auth2,
	}

	assert.True(t, feePayerAuth.Verify(msg))
	assert.False(t, feePayerAuth.Verify([]byte("wrong message")))
}

func TestMultiAgentTransaction_SigningMessage(t *testing.T) {
	rawTxn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	multiAgentTxn := &MultiAgentTransaction{
		RawTxn:           rawTxn,
		SecondarySigners: []AccountAddress{AccountTwo},
	}

	msg, err := multiAgentTxn.SigningMessage()
	require.NoError(t, err)
	assert.NotEmpty(t, msg)
}

func TestFeePayerTransaction_SigningMessage(t *testing.T) {
	rawTxn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	feePayerTxn := &FeePayerTransaction{
		RawTxn:           rawTxn,
		SecondarySigners: []AccountAddress{},
		FeePayer:         AccountThree,
	}

	msg, err := feePayerTxn.SigningMessage()
	require.NoError(t, err)
	assert.NotEmpty(t, msg)
}

func TestSignTransaction(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := SignTransaction(key, txn)
	require.NoError(t, err)
	assert.NotNil(t, signed)
	assert.NotNil(t, signed.Transaction)
	assert.NotNil(t, signed.Authenticator)
}

func TestNewFeePayerSignedTransaction(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := NewFeePayerSignedTransaction(
		txn,
		auth1,
		[]AccountAddress{},
		[]*AccountAuthenticator{},
		AccountThree,
		auth2,
	)
	require.NoError(t, err)
	assert.NotNil(t, signed)
}

func TestNewMultiAgentSignedTransaction(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := NewMultiAgentSignedTransaction(
		txn,
		auth1,
		[]AccountAddress{AccountTwo},
		[]*AccountAuthenticator{auth2},
	)
	require.NoError(t, err)
	assert.NotNil(t, signed)
}

func TestNewSingleSenderSignedTransaction(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := NewSingleSenderSignedTransaction(txn, auth)
	require.NoError(t, err)
	assert.NotNil(t, signed)
}

func TestSignFeePayerTransaction(t *testing.T) {
	sender, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	feePayer, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := SignFeePayerTransaction(
		sender,
		feePayer,
		txn,
		AccountThree,
		nil,
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, signed)
}

func TestSignMultiAgentTransaction(t *testing.T) {
	sender, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	secondary, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := SignMultiAgentTransaction(
		sender,
		[]Signer{secondary},
		[]AccountAddress{AccountTwo},
		txn,
	)
	require.NoError(t, err)
	assert.NotNil(t, signed)
}

func TestSimulationAuthenticator(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	auth := SimulationAuthenticator(key)
	assert.NotNil(t, auth)
}

func TestRawTransactionWithDataPrehash(t *testing.T) {
	prehash := RawTransactionWithDataPrehash()
	assert.Len(t, prehash, 32)
}

func TestSignedTransaction_BCSRoundTrip(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	txn := &RawTransaction{
		Sender:                     AccountOne,
		SequenceNumber:             1,
		MaxGasAmount:               10000,
		GasUnitPrice:               100,
		ExpirationTimestampSeconds: 1700000000,
		ChainID:                    4,
		Payload: &EntryFunctionPayload{
			Module:   ModuleID{Address: AccountOne, Name: "coin"},
			Function: "transfer",
			TypeArgs: nil,
			Args:     []any{AccountTwo[:], uint64(1000)},
		},
	}

	signed, err := SignTransaction(key, txn)
	require.NoError(t, err)

	// Serialize
	data, err := bcs.Serialize(signed)
	require.NoError(t, err)

	// Deserialize
	signed2 := &SignedTransaction{}
	err = bcs.Deserialize(signed2, data)
	require.NoError(t, err)

	assert.Equal(t, signed.Transaction.Sender, signed2.Transaction.Sender)
	assert.Equal(t, signed.Transaction.SequenceNumber, signed2.Transaction.SequenceNumber)
}

func TestSingleSenderAuthenticator_UnmarshalBCS(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	singleAuth := &SingleSenderAuthenticator{Sender: auth}

	// Serialize
	data, err := bcs.Serialize(singleAuth)
	require.NoError(t, err)

	// Deserialize - note: UnmarshalBCS doesn't read variant, so skip it
	des := bcs.NewDeserializer(data)
	_ = des.Uleb128() // Skip variant

	singleAuth2 := &SingleSenderAuthenticator{}
	singleAuth2.UnmarshalBCS(des)

	require.NoError(t, des.Error())
	assert.True(t, singleAuth2.Verify(msg))
}

func TestEd25519TransactionAuthenticator_UnmarshalBCS(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	ed25519Auth := &Ed25519TransactionAuthenticator{Sender: auth}

	// Serialize
	data, err := bcs.Serialize(ed25519Auth)
	require.NoError(t, err)

	// Deserialize - note: UnmarshalBCS doesn't read variant, so skip it
	des := bcs.NewDeserializer(data)
	_ = des.Uleb128() // Skip variant

	ed25519Auth2 := &Ed25519TransactionAuthenticator{}
	ed25519Auth2.UnmarshalBCS(des)

	require.NoError(t, des.Error())
	assert.True(t, ed25519Auth2.Verify(msg))
}

func TestMultiAgentAuthenticator_UnmarshalBCS(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	multiAuth := &MultiAgentAuthenticator{
		Sender:                   auth1,
		SecondarySignerAddresses: []AccountAddress{AccountTwo},
		SecondarySigners:         []*AccountAuthenticator{auth2},
	}

	// Serialize
	data, err := bcs.Serialize(multiAuth)
	require.NoError(t, err)

	// Deserialize - note: UnmarshalBCS doesn't read variant, so skip it
	des := bcs.NewDeserializer(data)
	_ = des.Uleb128() // Skip variant

	multiAuth2 := &MultiAgentAuthenticator{}
	multiAuth2.UnmarshalBCS(des)

	require.NoError(t, des.Error())
	assert.True(t, multiAuth2.Verify(msg))
}

func TestFeePayerAuthenticator_UnmarshalBCS(t *testing.T) {
	senderKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	feePayerKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	senderAuth, err := senderKey.Sign(msg)
	require.NoError(t, err)
	feePayerAuth, err := feePayerKey.Sign(msg)
	require.NoError(t, err)

	feeAuth := &FeePayerAuthenticator{
		Sender:                   senderAuth,
		SecondarySignerAddresses: []AccountAddress{},
		SecondarySigners:         []*AccountAuthenticator{},
		FeePayerAddress:          AccountThree,
		FeePayerAuth:             feePayerAuth,
	}

	// Serialize
	data, err := bcs.Serialize(feeAuth)
	require.NoError(t, err)

	// Deserialize - note: UnmarshalBCS doesn't read variant, so skip it
	des := bcs.NewDeserializer(data)
	_ = des.Uleb128() // Skip variant

	feeAuth2 := &FeePayerAuthenticator{}
	feeAuth2.UnmarshalBCS(des)

	require.NoError(t, des.Error())
	assert.True(t, feeAuth2.Verify(msg))
}

func TestDeserializeTransactionAuthenticator_Ed25519(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	ed25519Auth := &Ed25519TransactionAuthenticator{Sender: auth}

	// Serialize
	data, err := bcs.Serialize(ed25519Auth)
	require.NoError(t, err)

	// Deserialize using the function
	des := bcs.NewDeserializer(data)
	result := deserializeTransactionAuthenticator(des)

	require.NoError(t, des.Error())
	require.NotNil(t, result)
	assert.True(t, result.Verify(msg))
}

func TestDeserializeTransactionAuthenticator_SingleSender(t *testing.T) {
	key, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth, err := key.Sign(msg)
	require.NoError(t, err)

	singleAuth := &SingleSenderAuthenticator{Sender: auth}

	// Serialize
	data, err := bcs.Serialize(singleAuth)
	require.NoError(t, err)

	// Deserialize using the function
	des := bcs.NewDeserializer(data)
	result := deserializeTransactionAuthenticator(des)

	require.NoError(t, des.Error())
	require.NotNil(t, result)
	assert.True(t, result.Verify(msg))
}

func TestDeserializeTransactionAuthenticator_MultiAgent(t *testing.T) {
	key1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	key2, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	auth1, err := key1.Sign(msg)
	require.NoError(t, err)
	auth2, err := key2.Sign(msg)
	require.NoError(t, err)

	multiAuth := &MultiAgentAuthenticator{
		Sender:                   auth1,
		SecondarySignerAddresses: []AccountAddress{AccountTwo},
		SecondarySigners:         []*AccountAuthenticator{auth2},
	}

	// Serialize
	data, err := bcs.Serialize(multiAuth)
	require.NoError(t, err)

	// Deserialize using the function
	des := bcs.NewDeserializer(data)
	result := deserializeTransactionAuthenticator(des)

	require.NoError(t, des.Error())
	require.NotNil(t, result)
	assert.True(t, result.Verify(msg))
}

func TestDeserializeTransactionAuthenticator_FeePayer(t *testing.T) {
	senderKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	feePayerKey, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)

	msg := []byte("test message")
	senderAuth, err := senderKey.Sign(msg)
	require.NoError(t, err)
	feePayerAuth, err := feePayerKey.Sign(msg)
	require.NoError(t, err)

	feeAuth := &FeePayerAuthenticator{
		Sender:                   senderAuth,
		SecondarySignerAddresses: []AccountAddress{},
		SecondarySigners:         []*AccountAuthenticator{},
		FeePayerAddress:          AccountThree,
		FeePayerAuth:             feePayerAuth,
	}

	// Serialize
	data, err := bcs.Serialize(feeAuth)
	require.NoError(t, err)

	// Deserialize using the function
	des := bcs.NewDeserializer(data)
	result := deserializeTransactionAuthenticator(des)

	require.NoError(t, des.Error())
	require.NotNil(t, result)
	assert.True(t, result.Verify(msg))
}
