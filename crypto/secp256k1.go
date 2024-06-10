package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

//region Secp256k1PrivateKey

const Secp256k1PrivateKeyLength = 32

// Secp256k1PublicKeyLength we use the uncompressed version
const Secp256k1PublicKeyLength = 65

// Secp256k1SignatureLength is the Secp256k1 signature without the recovery bit
const Secp256k1SignatureLength = ethCrypto.SignatureLength - 1

// Secp256k1PrivateKey is a private key that can be used with SingleSigner.  It cannot stand on its own.
// Implements MessageSigner, CryptoMaterial
type Secp256k1PrivateKey struct {
	Inner *ecdsa.PrivateKey
}

func GenerateSecp256k1Key() (*Secp256k1PrivateKey, error) {
	priv, err := ethCrypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	return &Secp256k1PrivateKey{priv}, nil
}

//region Secp256k1PrivateKey MessageSigner

func (key *Secp256k1PrivateKey) VerifyingKey() VerifyingKey {
	return &Secp256k1PublicKey{
		&key.Inner.PublicKey,
	}
}

func (key *Secp256k1PrivateKey) SignMessage(msg []byte) (sig Signature, err error) {
	hash := util.Sha3256Hash([][]byte{msg})
	// TODO: The eth library doesn't protect against malleability issues, so we need to handle those
	signature, err := secp256k1.Sign(hash, key.Bytes())
	if err != nil {
		return nil, err
	}

	// Strip the recovery bit
	return &Secp256k1Signature{
		signature[0:64],
	}, nil
}

//endregion

//region Secp256k1PrivateKey CryptoMaterial

func (key *Secp256k1PrivateKey) Bytes() []byte {
	return ethCrypto.FromECDSA(key.Inner)
}

func (key *Secp256k1PrivateKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != Secp256k1PrivateKeyLength {
		return fmt.Errorf("invalid secp256k1 private key size %d", len(bytes))
	}
	key.Inner, err = ethCrypto.ToECDSA(bytes)
	return err
}

func (key *Secp256k1PrivateKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

//endregion

func (key *Secp256k1PrivateKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion
//endregion

//region Secp256k1PublicKey

// Secp256k1PublicKey is the corresponding public key for Secp256k1PrivateKey, it cannot be used on its own
// Implements VerifyingKey, CryptoMaterial, bcs.Struct
type Secp256k1PublicKey struct {
	Inner *ecdsa.PublicKey
}

//region Secp256k1PublicKey VerifyingKey

func (key *Secp256k1PublicKey) Verify(msg []byte, sig Signature) bool {
	switch sig.(type) {
	case *Secp256k1Signature:
		typedSig := sig.(*Secp256k1Signature)

		return secp256k1.VerifySignature(key.Bytes(), msg, typedSig.Bytes())
	default:
		panic(fmt.Errorf("invalid signature type %T", sig))
		return false
	}
}

//endregion

//region Secp256k1PublicKey CryptoMaterial

func (key *Secp256k1PublicKey) Bytes() []byte {
	return ethCrypto.FromECDSAPub(key.Inner)
}

func (key *Secp256k1PublicKey) FromBytes(bytes []byte) (err error) {
	key.Inner, err = ethCrypto.UnmarshalPubkey(bytes)
	return err
}

func (key *Secp256k1PublicKey) ToHex() string {
	return util.BytesToHex(key.Bytes())
}

func (key *Secp256k1PublicKey) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return key.FromBytes(bytes)
}

//endregion

//region Secp256k1PublicKey bcs.Struct

func (key *Secp256k1PublicKey) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(key.Bytes())
}
func (key *Secp256k1PublicKey) UnmarshalBCS(des *bcs.Deserializer) {
	kb := des.ReadBytes()
	pubKey, err := ethCrypto.UnmarshalPubkey(kb)
	if err != nil {
		des.SetError(err)
		return
	}
	key.Inner = pubKey
}

//endregion
//endregion

//region Secp256k1Authenticator

// Secp256k1Authenticator is the authenticator for Secp256k1, but it cannot stand on its own and must be used with SingleKeyAuthenticator
// Implements AccountAuthenticatorImpl, bcs.Struct
type Secp256k1Authenticator struct {
	PubKey *Secp256k1PublicKey
	Sig    *Secp256k1Signature
}

//region Secp256k1Authenticator AccountAuthenticatorImpl

func (ea *Secp256k1Authenticator) PublicKey() VerifyingKey {
	return ea.PubKey
}

func (ea *Secp256k1Authenticator) Signature() Signature {
	return ea.Sig
}

func (ea *Secp256k1Authenticator) Verify(msg []byte) bool {
	return ea.PubKey.Verify(msg, ea.Sig)
}

//endregion

//region Secp256k1Authenticator bcs.Struct

func (ea *Secp256k1Authenticator) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(ea.PublicKey())
	ser.Struct(ea.Signature())
}

func (ea *Secp256k1Authenticator) UnmarshalBCS(des *bcs.Deserializer) {
	ea.PubKey = &Secp256k1PublicKey{}
	des.Struct(ea.PubKey)
	err := des.Error()
	if err != nil {
		return
	}
	ea.Sig = &Secp256k1Signature{}
	des.Struct(ea.Sig)
}

//endregion
//endregion

//region Secp256k1Signature

// Secp256k1Signature a wrapper for serialization of Secp256k1 signatures
// Implements Signature, CryptoMaterial
type Secp256k1Signature struct {
	Inner []byte
}

//region Secp256k1Signature CryptoMaterial

func (e *Secp256k1Signature) Bytes() []byte {
	return e.Inner
}

func (e *Secp256k1Signature) FromBytes(bytes []byte) (err error) {
	if len(bytes) != Secp256k1SignatureLength {
		return fmt.Errorf("invalid secp256k1 signature size %d, expected %d", len(bytes), Secp256k1SignatureLength)
	}
	e.Inner = bytes
	return nil
}

func (e *Secp256k1Signature) ToHex() string {
	return util.BytesToHex(e.Bytes())
}

func (e *Secp256k1Signature) FromHex(hexStr string) (err error) {
	bytes, err := util.ParseHex(hexStr)
	if err != nil {
		return err
	}
	return e.FromBytes(bytes)
}

//endregion

//region Secp256k1Signature bcs.Struct

func (e *Secp256k1Signature) MarshalBCS(ser *bcs.Serializer) {
	ser.WriteBytes(e.Bytes())
}

func (e *Secp256k1Signature) UnmarshalBCS(des *bcs.Deserializer) {
	bytes := des.ReadBytes()
	if des.Error() != nil {
		return
	}
	err := e.FromBytes(bytes)
	if err != nil {
		des.SetError(err)
	}
}

//endregion
//endregion
