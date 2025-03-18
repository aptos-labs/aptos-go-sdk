package types

import (
	"errors"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

const (
	defaultMetadata = "0x2ebb2ccac5e027a87fa0e2e5f656a3a4238d6a48d93ec9b610d570fc0aa0df12"
	defaultStore    = "0x8a9d57692a9d4deb1680eaf107b83c152436e10f7bb521143fa403fa95ef76a"
	defaultOwner    = "0xc67545d6f3d36ed01efc9b28cbfd0c1ae326d5d262dd077a29539bcee0edce9e"
)

func TestGenerateEd25519Account(t *testing.T) {
	message := []byte{0x12, 0x34}
	account, err := NewEd25519Account()
	assert.NoError(t, err)
	output, err := account.Sign(message)
	assert.NoError(t, err)
	assert.Equal(t, crypto.AccountAuthenticatorEd25519, output.Variant)
	assert.True(t, output.Auth.Verify(message))
}

func TestGenerateSingleSignerEd25519Account(t *testing.T) {
	message := []byte{0x12, 0x34}
	account, err := NewEd25519SingleSignerAccount()
	assert.NoError(t, err)
	output, err := account.Sign(message)
	assert.NoError(t, err)
	assert.Equal(t, crypto.AccountAuthenticatorSingleSender, output.Variant)
	assert.True(t, output.Auth.Verify(message))
}

func TestGenerateSecp256k1Account(t *testing.T) {
	message := []byte{0x12, 0x34}
	account, err := NewSecp256k1Account()
	assert.NoError(t, err)
	output, err := account.Sign(message)
	assert.NoError(t, err)
	assert.Equal(t, crypto.AccountAuthenticatorSingleSender, output.Variant)
	assert.True(t, output.Auth.Verify(message))
}

func TestNewAccountFromSigner(t *testing.T) {
	message := []byte{0x12, 0x34}
	key, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	account, err := NewAccountFromSigner(key)
	assert.NoError(t, err)
	output, err := account.Sign(message)
	assert.NoError(t, err)
	assert.Equal(t, crypto.AccountAuthenticatorEd25519, output.Variant)
	assert.True(t, output.Auth.Verify(message))

	authKey := key.AuthKey()
	assert.Equal(t, authKey[:], account.Address[:])
}

func TestNewAccountFromSignerWithAddress(t *testing.T) {
	message := []byte{0x12, 0x34}
	key, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	account, err := NewAccountFromSigner(key, AccountZero)
	assert.NoError(t, err)
	output, err := account.Sign(message)
	assert.NoError(t, err)
	assert.Equal(t, crypto.AccountAuthenticatorEd25519, output.Variant)
	assert.True(t, output.Auth.Verify(message))

	outputSig, err := account.SignMessage(message)
	assert.NoError(t, err)
	assert.True(t, account.Signer.PubKey().Verify(message, outputSig))

	assert.Equal(t, AccountZero, account.Address)
	assert.Equal(t, AccountZero, account.AccountAddress())
	assert.Equal(t, key.AuthKey(), account.AuthKey())
	assert.Equal(t, key.PubKey(), account.PubKey())
}

func TestNewAccountFromSignerWithAddressMulti(t *testing.T) {
	key, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	_, err = NewAccountFromSigner(key, AccountZero, AccountOne)
	assert.Error(t, err)
}

type WrapperSigner struct {
	signer crypto.Signer
}

func (w *WrapperSigner) Sign(_ []byte) (*crypto.AccountAuthenticator, error) {
	return nil, errors.New("not implemented")
}
func (w *WrapperSigner) SignMessage(_ []byte) (crypto.Signature, error) {
	return nil, errors.New("not implemented")
}
func (w *WrapperSigner) SimulationAuthenticator() *crypto.AccountAuthenticator {
	return nil
}
func (w *WrapperSigner) AuthKey() *crypto.AuthenticationKey {
	return &crypto.AuthenticationKey{}
}
func (w *WrapperSigner) PubKey() crypto.PublicKey {
	// Note this is just for testing
	return &crypto.Ed25519PublicKey{}
}

func TestAccount_ExtractMessageSigner(t *testing.T) {
	ed25519PrivateKey, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)
	ed25519Account, err := NewAccountFromSigner(ed25519PrivateKey)
	assert.NoError(t, err)

	ed25519Out, ok := ed25519Account.MessageSigner()
	assert.True(t, ok)
	assert.Equal(t, ed25519PrivateKey, ed25519Out)

	ed25519SingleSignerAccount, err := NewAccountFromSigner(crypto.NewSingleSigner(ed25519PrivateKey))
	assert.NoError(t, err)

	ed25519Out, ok = ed25519SingleSignerAccount.MessageSigner()
	assert.True(t, ok)
	assert.Equal(t, ed25519PrivateKey, ed25519Out)

	secp256k1PrivateKey, err := crypto.GenerateSecp256k1Key()
	assert.NoError(t, err)
	secp256k1SingleSignerAccount, err := NewAccountFromSigner(crypto.NewSingleSigner(secp256k1PrivateKey))
	assert.NoError(t, err)

	secp256k1Out, ok := secp256k1SingleSignerAccount.MessageSigner()
	assert.True(t, ok)
	assert.Equal(t, secp256k1PrivateKey, secp256k1Out)

	wrapperSigner := &WrapperSigner{signer: secp256k1SingleSignerAccount}
	customAccount, err := NewAccountFromSigner(wrapperSigner)
	assert.NoError(t, err)
	out, ok := customAccount.MessageSigner()
	assert.False(t, ok)
	assert.Nil(t, out)
}

func TestAccount_ExtractPrivateKeyString(t *testing.T) {
	ed25519PrivateKey, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)
	ed25519Account, err := NewAccountFromSigner(ed25519PrivateKey)
	assert.NoError(t, err)

	ed25519KeyString, err := ed25519Account.PrivateKeyString()
	assert.NoError(t, err)
	expectedEd25519String, err := ed25519PrivateKey.ToAIP80()
	assert.NoError(t, err)
	assert.Equal(t, expectedEd25519String, ed25519KeyString)

	ed25519SingleSignerAccount, err := NewAccountFromSigner(crypto.NewSingleSigner(ed25519PrivateKey))
	assert.NoError(t, err)

	ed25519SingleSignerKeyString, err := ed25519SingleSignerAccount.PrivateKeyString()
	assert.NoError(t, err)
	assert.Equal(t, expectedEd25519String, ed25519SingleSignerKeyString)

	secp256k1PrivateKey, err := crypto.GenerateSecp256k1Key()
	assert.NoError(t, err)
	secp256k1SingleSignerAccount, err := NewAccountFromSigner(crypto.NewSingleSigner(secp256k1PrivateKey))
	assert.NoError(t, err)

	expectedSecp256k1String, err := secp256k1PrivateKey.ToAIP80()
	assert.NoError(t, err)
	secp256k1SingleSignerKeyString, err := secp256k1SingleSignerAccount.PrivateKeyString()
	assert.NoError(t, err)
	assert.Equal(t, expectedSecp256k1String, secp256k1SingleSignerKeyString)

	wrapperSigner := &WrapperSigner{signer: secp256k1SingleSignerAccount}
	customAccount, err := NewAccountFromSigner(wrapperSigner)
	assert.NoError(t, err)
	out, err := customAccount.PrivateKeyString()
	assert.Error(t, err)
	assert.Empty(t, out)
}
