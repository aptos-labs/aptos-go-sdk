package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	defaultMetadata = "0x2ebb2ccac5e027a87fa0e2e5f656a3a4238d6a48d93ec9b610d570fc0aa0df12"
	defaultStore    = "0x8a9d57692a9d4deb1680eaf107b83c152436e10f7bb521143fa403fa95ef76a"
	defaultOwner    = "0xc67545d6f3d36ed01efc9b28cbfd0c1ae326d5d262dd077a29539bcee0edce9e"
)

func TestAccountSpecialString(t *testing.T) {
	var aa AccountAddress
	aa[31] = 3
	aas := aa.String()
	if aas != "0x3" {
		t.Errorf("wanted 0x3 got %s", aas)
	}

	var aa2 AccountAddress
	err := aa2.ParseStringRelaxed("0x3")
	if err != nil {
		t.Errorf("unexpected err %s", err)
	}
	if aa2 != aa {
		t.Errorf("aa2 != aa")
	}
}

func TestAccountAddress_ParseStringRelaxed_Error(t *testing.T) {
	var owner AccountAddress
	err := owner.ParseStringRelaxed("0x")
	assert.Error(t, err)
	err = owner.ParseStringRelaxed("0xF1234567812345678123456781234567812345678123456781234567812345678")
	assert.Error(t, err)
}
func TestAccountAddress_ObjectAddressFromObject(t *testing.T) {
	var owner AccountAddress
	err := owner.ParseStringRelaxed(defaultOwner)
	assert.NoError(t, err)

	var objectAddress AccountAddress
	err = objectAddress.ParseStringRelaxed(defaultMetadata)
	assert.NoError(t, err)

	var expectedDerivedAddress AccountAddress
	err = expectedDerivedAddress.ParseStringRelaxed(defaultStore)
	assert.NoError(t, err)

	derivedAddress := owner.ObjectAddressFromObject(&objectAddress)
	assert.NoError(t, err)

	assert.Equal(t, expectedDerivedAddress, derivedAddress)
}
