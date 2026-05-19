package aptos

import (
	"bytes"
	"crypto/sha3"
	"math/big"
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

	// Verify the message starts with the expected prehash prefix.
	// The prehash is SHA3-256("APTOS::RawTransaction") — Aptos signs the
	// SHA3-256 of the salt concatenated with the BCS-encoded raw txn.
	expectedPrehash := sha3.Sum256([]byte(RawTransactionSalt))
	assert.True(t, bytes.HasPrefix(msg, expectedPrehash[:]),
		"signing message should start with SHA3-256 of RawTransactionSalt")

	// The message should be longer than just the prehash (prehash + serialized txn)
	assert.Greater(t, len(msg), 32, "signing message should contain prehash + serialized transaction")

	// Signing the same transaction twice should produce identical messages
	msg2, err := txn.SigningMessage()
	require.NoError(t, err)
	assert.Equal(t, msg, msg2, "signing message should be deterministic")
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

	// Hash must be exactly SHA3-256(prefix || 0x00 || bcs(SignedTransaction)).
	// Building the expected hash by hand pins the prefix construction —
	// catches regressions in slice-aliasing or wrong variant byte.
	signedBytes, err := bcs.Marshal(signed)
	require.NoError(t, err)
	prefix := sha3.Sum256([]byte("APTOS::Transaction"))
	expected := make([]byte, 0, len(prefix)+1+len(signedBytes))
	expected = append(expected, prefix[:]...)
	expected = append(expected, 0)
	expected = append(expected, signedBytes...)
	want := sha3.Sum256(expected)
	wantHex := "0x" + hexEncode(want[:])
	assert.Equal(t, wantHex, hash)
}

// hexEncode returns the lowercase hex string of b without "0x" prefix.
// Local helper to avoid importing "encoding/hex" twice.
func hexEncode(b []byte) string {
	const hexDigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexDigits[v>>4]
		out[i*2+1] = hexDigits[v&0x0f]
	}
	return string(out)
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
	// Each case asserts the exact BCS bytes serializeArg should emit for
	// a given Move-typed argument. This is the contract that
	// EntryFunctionPayload.Args relies on; assertNotEmpty wasn't enough
	// to catch a vector<u8> length-prefix regression.
	addr := AccountOne
	tests := []struct {
		name string
		arg  any
		want []byte
	}{
		{"bytes", []byte{1, 2, 3}, []byte{0x03, 0x01, 0x02, 0x03}},
		{"empty bytes", []byte{}, []byte{0x00}},
		{"string", "hi", []byte{0x02, 'h', 'i'}},
		{"empty string", "", []byte{0x00}},
		{"uint8", uint8(0x7f), []byte{0x7f}},
		{"uint16", uint16(0x1234), []byte{0x34, 0x12}},
		{"uint32", uint32(0xdeadbeef), []byte{0xef, 0xbe, 0xad, 0xde}},
		{"uint64", uint64(1), []byte{1, 0, 0, 0, 0, 0, 0, 0}},
		{"bool true", true, []byte{0x01}},
		{"bool false", false, []byte{0x00}},
		{"int8", int8(-1), []byte{0xff}},
		{"int16", int16(-1), []byte{0xff, 0xff}},
		{"int32", int32(-1), []byte{0xff, 0xff, 0xff, 0xff}},
		{"int64", int64(-1), []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{"address", addr, addr[:]},
		{"address_ptr", &addr, addr[:]},
		{"option none", None(), []byte{0x00}},
		{"option some uint64", Some(uint64(7)), []byte{0x01, 7, 0, 0, 0, 0, 0, 0, 0}},
		{"option some string", Some("ab"), []byte{0x01, 0x02, 'a', 'b'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := serializeArg(tt.arg)
			require.NoError(t, err)
			assert.Equal(t, tt.want, data)
		})
	}
}

func TestSerializeArg_BigInts(t *testing.T) {
	// Byte-exact tests differentiating signed/unsigned and width.
	// `assert.Len` is not enough: U128(1) and I128(1) both produce
	// 16 bytes, so a length-only test would silently pass if the
	// signed/unsigned dispatch were swapped.

	// u128(1) and u256(1): little-endian, all zero except the lowest byte.
	u128One := make([]byte, 16)
	u128One[0] = 0x01
	u256One := make([]byte, 32)
	u256One[0] = 0x01

	// i128(-1) and i256(-1): two's complement, all 0xff.
	i128NegOne := bytes.Repeat([]byte{0xff}, 16)
	i256NegOne := bytes.Repeat([]byte{0xff}, 32)

	cases := []struct {
		name string
		arg  any
		want []byte
	}{
		{"u128 = 1", U128Arg{Value: big.NewInt(1)}, u128One},
		{"u256 = 1", U256Arg{Value: big.NewInt(1)}, u256One},
		{"i128 = -1", I128Arg{Value: big.NewInt(-1)}, i128NegOne},
		{"i256 = -1", I256Arg{Value: big.NewInt(-1)}, i256NegOne},
		// i128(1) must NOT equal u128(1) byte-for-byte if dispatch
		// is correct — both happen to be the same for value 1, so
		// also test a negative on the signed path which is illegal
		// for the unsigned path.
		{"i128 = 1", I128Arg{Value: big.NewInt(1)}, u128One}, // shares value but goes through signed path
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := serializeArg(tc.arg)
			require.NoError(t, err)
			assert.Equal(t, tc.want, data)
		})
	}
}

func TestSerializeArg_NilBigInts(t *testing.T) {
	cases := []any{
		U128Arg{Value: nil},
		U256Arg{Value: nil},
		I128Arg{Value: nil},
		I256Arg{Value: nil},
	}
	for _, c := range cases {
		_, err := serializeArg(c)
		require.Error(t, err)
	}
}

func TestSerializeArg_NilArg(t *testing.T) {
	_, err := serializeArg(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil argument")
}

func TestSerializeArg_NilAddress(t *testing.T) {
	var nilAddr *AccountAddress
	_, err := serializeArg(nilAddr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil address")
}

func TestSerializeArg_OptionValueAndPointer(t *testing.T) {
	// Option may be passed as either *Option (the common case from
	// Some/None) or Option-by-value. Both must produce identical bytes.
	byValue, err := serializeArg(Option{Value: uint64(7)})
	require.NoError(t, err)
	byPtr, err := serializeArg(&Option{Value: uint64(7)})
	require.NoError(t, err)
	assert.Equal(t, byPtr, byValue)
	assert.Equal(t, []byte{0x01, 7, 0, 0, 0, 0, 0, 0, 0}, byValue)
}

// fakeMarshaler implements bcs.Marshaler. Used to confirm that
// serializeArg honors the bcs.Marshaler case path.
type fakeMarshaler struct{ payload []byte }

func (f *fakeMarshaler) MarshalBCS(ser *bcs.Serializer) {
	ser.FixedBytes(f.payload)
}

func (f *fakeMarshaler) UnmarshalBCS(_ *bcs.Deserializer) {}

func TestSerializeArg_BCSMarshaler(t *testing.T) {
	got, err := serializeArg(&fakeMarshaler{payload: []byte{0xde, 0xad}})
	require.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xad}, got)
}

func TestSerializeArg_NestedOption(t *testing.T) {
	// Option<Option<u64>> = Some(Some(1)) should produce:
	//   ULEB128(1) || ULEB128(1) || 8 bytes
	data, err := serializeArg(Some(Some(uint64(1))))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x01, 1, 0, 0, 0, 0, 0, 0, 0}, data)

	// Option<Option<u64>> = Some(None) -> 0x01 || 0x00
	data, err = serializeArg(Some(None()))
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x00}, data)
}

func TestEntryFunction_VectorU8ArgEncoding(t *testing.T) {
	// Regression test for the vector<u8> length-prefix bug.
	//
	// For a vector<u8> arg containing N bytes, the on-chain BCS must be:
	//   outer ULEB128(N+sizeof(uleb128(N))) || ULEB128(N) || N bytes
	// i.e. the inner Move value is itself length-prefixed before the
	// EntryFunction args[]vec wraps each element in WriteBytes.
	payload := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "test"},
		Function: "f",
		TypeArgs: nil,
		Args:     []any{[]byte{0xaa, 0xbb, 0xcc}},
	}

	ser := bcs.NewSerializer()
	serializePayload(ser, payload)
	require.NoError(t, ser.Error())
	data := ser.ToBytes()

	// Round-trip preserves the raw arg bytes — confirm the arg vector
	// element contains the inner length prefix.
	des := bcs.NewDeserializer(data)
	got := deserializePayload(des)
	require.NoError(t, des.Error())
	ef, ok := got.(*EntryFunctionPayload)
	require.True(t, ok)

	// EntryFunction args after deserialization are []byte buffers
	// containing the raw BCS of each arg.
	require.Len(t, ef.Args, 1)
	rawArg, ok := ef.Args[0].([]byte)
	require.True(t, ok, "deserialized arg should be []byte")
	// Inner encoding: ULEB128(3) || 0xaa 0xbb 0xcc.
	assert.Equal(t, []byte{0x03, 0xaa, 0xbb, 0xcc}, rawArg)
}

func TestEntryFunction_AddressArgEncoding(t *testing.T) {
	// An `address` arg must be encoded as the 32 raw fixed bytes
	// (no inner length prefix), wrapped in a single outer
	// ULEB128(32) prefix by the EntryFunction args vector.
	payload := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "test"},
		Function: "f",
		TypeArgs: nil,
		Args:     []any{AccountOne},
	}

	ser := bcs.NewSerializer()
	serializePayload(ser, payload)
	require.NoError(t, ser.Error())

	des := bcs.NewDeserializer(ser.ToBytes())
	got := deserializePayload(des)
	require.NoError(t, des.Error())
	ef, _ := got.(*EntryFunctionPayload)
	require.Len(t, ef.Args, 1)
	rawArg, _ := ef.Args[0].([]byte)
	// Expect the 32 raw bytes of AccountOne — no inner ULEB128 length.
	assert.Equal(t, AccountOne[:], rawArg)
}

func TestEntryFunction_OptionArgEncoding(t *testing.T) {
	// Option<u64> args used to be impossible to express. Confirm that
	// Some(x) and None() round-trip through the EntryFunction encoding.
	none := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "test"},
		Function: "f",
		Args:     []any{None()},
	}
	ser := bcs.NewSerializer()
	serializePayload(ser, none)
	require.NoError(t, ser.Error())
	des := bcs.NewDeserializer(ser.ToBytes())
	ef, _ := deserializePayload(des).(*EntryFunctionPayload)
	require.NoError(t, des.Error())
	require.Len(t, ef.Args, 1)
	rawArg, _ := ef.Args[0].([]byte)
	assert.Equal(t, []byte{0x00}, rawArg) // Move Option::None

	some := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "test"},
		Function: "f",
		Args:     []any{Some(uint64(42))},
	}
	ser = bcs.NewSerializer()
	serializePayload(ser, some)
	require.NoError(t, ser.Error())
	des = bcs.NewDeserializer(ser.ToBytes())
	ef, _ = deserializePayload(des).(*EntryFunctionPayload)
	require.NoError(t, des.Error())
	rawArg, _ = ef.Args[0].([]byte)
	assert.Equal(t, []byte{0x01, 42, 0, 0, 0, 0, 0, 0, 0}, rawArg)
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
	require.NotNil(t, auth)

	// Ed25519 private key should produce Ed25519 authenticator variant
	assert.Equal(t, AccountAuthenticatorEd25519, auth.Variant,
		"Ed25519 key should produce Ed25519 authenticator variant")

	// The authenticator should contain a valid public key
	pubKey := auth.PubKey()
	assert.NotNil(t, pubKey, "simulation authenticator should contain a public key")

	// The public key should match the key's verifying key
	expectedPubKey := key.VerifyingKey()
	assert.Equal(t, expectedPubKey.Bytes(), pubKey.Bytes(),
		"simulation authenticator public key should match the signer's public key")
}

func TestRawTransactionWithDataPrehash(t *testing.T) {
	prehash := RawTransactionWithDataPrehash()
	assert.Len(t, prehash, 32)

	// Verify the prehash matches the expected SHA3-256 of the salt.
	expectedHash := sha3.Sum256([]byte(RawTransactionWithDataSalt))
	assert.Equal(t, expectedHash[:], prehash,
		"prehash should equal SHA3-256 of RawTransactionWithDataSalt")

	// Calling again should return the same value (deterministic)
	prehash2 := RawTransactionWithDataPrehash()
	assert.Equal(t, prehash, prehash2, "prehash should be deterministic")
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

func TestPayloadType_ViewPayload(t *testing.T) {
	t.Parallel()
	p := &ViewPayload{
		Module:   ModuleID{Address: AccountOne, Name: "coin"},
		Function: "balance",
	}
	assert.Equal(t, "view_function", p.payloadType())
}

func TestPayloadType_EntryFunctionPayload(t *testing.T) {
	t.Parallel()
	p := &EntryFunctionPayload{
		Module:   ModuleID{Address: AccountOne, Name: "coin"},
		Function: "transfer",
	}
	assert.Equal(t, "entry_function_payload", p.payloadType())
}

func TestPayloadType_ScriptPayload(t *testing.T) {
	t.Parallel()
	p := &ScriptPayload{Code: []byte{0x01}}
	assert.Equal(t, "script_payload", p.payloadType())
}

func TestDeserializeTransactionAuthenticator_UnknownVariant(t *testing.T) {
	t.Parallel()
	ser := bcs.NewSerializer()
	ser.Uleb128(99) // Unknown variant
	data := ser.ToBytes()

	des := bcs.NewDeserializer(data)
	result := deserializeTransactionAuthenticator(des)
	assert.Nil(t, result)
	assert.Error(t, des.Error())
	assert.Contains(t, des.Error().Error(), "unknown transaction authenticator variant")
}

func TestDeserializePayload_UnknownVariant(t *testing.T) {
	t.Parallel()
	ser := bcs.NewSerializer()
	ser.Uleb128(99) // Unknown variant
	data := ser.ToBytes()

	des := bcs.NewDeserializer(data)
	result := deserializePayload(des)
	assert.Nil(t, result)
	require.Error(t, des.Error())
	assert.Contains(t, des.Error().Error(), "unknown payload variant")
}

func TestSerializePayload_UnsupportedType(t *testing.T) {
	t.Parallel()
	ser := bcs.NewSerializer()
	serializePayload(ser, &ViewPayload{}) // ViewPayload is not supported in BCS serialization
	assert.Error(t, ser.Error())
	assert.Contains(t, ser.Error().Error(), "unsupported payload type")
}

func TestDeserializeScriptArg_UnknownVariant(t *testing.T) {
	t.Parallel()
	ser := bcs.NewSerializer()
	ser.U8(255) // Unknown variant
	data := ser.ToBytes()

	des := bcs.NewDeserializer(data)
	result := deserializeScriptArg(des)
	assert.Nil(t, result)
	assert.Error(t, des.Error())
	assert.Contains(t, des.Error().Error(), "unknown script argument variant")
}
