package types

import (
	"encoding/json"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountSpecialString(t *testing.T) {
	t.Parallel()
	var aa AccountAddress
	aa[31] = 3
	aas := aa.String()
	assert.Equal(t, "0x3", aas)

	var aa2 AccountAddress
	err := aa2.ParseStringRelaxed("0x3")
	require.NoError(t, err)
	assert.Equal(t, aa, aa2)
}

func TestAccountAddress_AuthKey(t *testing.T) {
	t.Parallel()
	authKey := &crypto.AuthenticationKey{}
	var aa AccountAddress
	aa.FromAuthKey(authKey)
	assert.Equal(t, AccountZero, aa)
}

func TestSpecialAddresses(t *testing.T) {
	t.Parallel()
	var addr AccountAddress
	err := addr.ParseStringRelaxed("0x0")
	require.NoError(t, err)
	assert.Equal(t, AccountZero, addr)
	err = addr.ParseStringRelaxed("0x1")
	require.NoError(t, err)
	assert.Equal(t, AccountOne, addr)
	err = addr.ParseStringRelaxed("0x2")
	require.NoError(t, err)
	assert.Equal(t, AccountTwo, addr)
	err = addr.ParseStringRelaxed("0x3")
	require.NoError(t, err)
	assert.Equal(t, AccountThree, addr)
	err = addr.ParseStringRelaxed("0x4")
	require.NoError(t, err)
	assert.Equal(t, AccountFour, addr)
}

func TestSerialize(t *testing.T) {
	t.Parallel()
	inputs := [][]byte{
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F},
		{0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x02, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x00, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x00, 0x04, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
	}

	for i := range len(inputs) {
		addr := AccountAddress(inputs[i])
		bytes, err := bcs.Serialize(&addr)
		require.NoError(t, err)
		assert.Equal(t, bytes, inputs[i])

		newAddr := AccountAddress{}
		err = bcs.Deserialize(&newAddr, bytes)
		require.NoError(t, err)
		assert.Equal(t, addr, newAddr)
	}
}

func TestStringOutput(t *testing.T) {
	t.Parallel()
	inputs := [][]byte{
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0F},
		{0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x02, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x00, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x00, 0x04, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
		{0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x12, 0x34, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
	}
	expected := []string{
		"0x0",
		"0x1",
		"0xf",
		"0x1234123412341234123412341234123412341234123412340123456789abcdef",
		"0x0234123412341234123412341234123412341234123412340123456789abcdef",
		"0x0034123412341234123412341234123412341234123412340123456789abcdef",
		"0x0004123412341234123412341234123412341234123412340123456789abcdef",
		"0x0000123412341234123412341234123412341234123412340123456789abcdef",
	}
	expectedLong := []string{
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		"0x0000000000000000000000000000000000000000000000000000000000000001",
		"0x000000000000000000000000000000000000000000000000000000000000000f",
		"0x1234123412341234123412341234123412341234123412340123456789abcdef",
		"0x0234123412341234123412341234123412341234123412340123456789abcdef",
		"0x0034123412341234123412341234123412341234123412340123456789abcdef",
		"0x0004123412341234123412341234123412341234123412340123456789abcdef",
		"0x0000123412341234123412341234123412341234123412340123456789abcdef",
	}
	for i := range len(inputs) {
		addr := AccountAddress(inputs[i])
		assert.Equal(t, expected[i], addr.String())
		assert.Equal(t, expectedLong[i], addr.StringLong())
	}
}

func TestAccountAddress_ParseStringRelaxed_Error(t *testing.T) {
	t.Parallel()
	var owner AccountAddress
	err := owner.ParseStringRelaxed("0x")
	require.Error(t, err)
	err = owner.ParseStringRelaxed("0xF1234567812345678123456781234567812345678123456781234567812345678")
	require.Error(t, err)
	err = owner.ParseStringRelaxed("NotHex")
	require.Error(t, err)
}

func TestAccountAddress_ObjectAddressFromObject(t *testing.T) {
	t.Parallel()
	var owner AccountAddress
	err := owner.ParseStringRelaxed(defaultOwner)
	require.NoError(t, err)

	var objectAddress AccountAddress
	err = objectAddress.ParseStringRelaxed(defaultMetadata)
	require.NoError(t, err)

	var expectedDerivedAddress AccountAddress
	err = expectedDerivedAddress.ParseStringRelaxed(defaultStore)
	require.NoError(t, err)

	derivedAddress := owner.ObjectAddressFromObject(&objectAddress)
	require.NoError(t, err)

	assert.Equal(t, expectedDerivedAddress, derivedAddress)
}

func TestAccountAddress_JSON(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Address *AccountAddress `json:"address"`
	}

	str := "{\"address\":\"0x1\"}"
	var test testStruct
	err := json.Unmarshal([]byte(str), &test)
	require.NoError(t, err)
	assert.Equal(t, &AccountOne, test.Address)

	b, err := json.Marshal(test)
	require.NoError(t, err)
	assert.Equal(t, str, string(b))
}
