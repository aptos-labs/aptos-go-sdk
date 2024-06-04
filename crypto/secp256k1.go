package crypto

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/decred/dcrd/dcrec/secp256k1"
	"math/big"
)

//region Secp256k1PrivateKey

// Secp256k1PrivateKey is a private key that can be used with SingleSigner.  It cannot stand on its own.
// Implements MessageSigner, CryptoMaterial
type Secp256k1PrivateKey struct {
	Inner *secp256k1.PrivateKey
}

func GenerateSecp256k1Key() Secp256k1PrivateKey {
	priv, _ := secp256k1.GeneratePrivateKey()

	return Secp256k1PrivateKey{priv}
}

//region Secp256k1PrivateKey MessageSigner

func (key *Secp256k1PrivateKey) VerifyingKey() VerifyingKey {
	pubKey := key.Inner.PubKey()
	return &Secp256k1PublicKey{
		pubKey,
	}
}

func (key *Secp256k1PrivateKey) SignMessage(msg []byte) (sig Signature, err error) {
	signature, err := key.Inner.Sign(msg)
	if err != nil {
		return nil, err
	}
	return &Secp256k1Signature{
		signature,
	}, nil
}

//endregion

//region Secp256k1PrivateKey CryptoMaterial

func (key *Secp256k1PrivateKey) Bytes() []byte {
	return key.Inner.Serialize()
}

func (key *Secp256k1PrivateKey) FromBytes(bytes []byte) (err error) {
	if len(bytes) != secp256k1.PrivKeyBytesLen {
		return fmt.Errorf("invalid secp256k1 private key size %d", len(bytes))
	}
	num := big.NewInt(0)
	num.SetBytes(bytes)
	key.Inner = secp256k1.NewPrivateKey(num)
	return nil
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
	Inner *secp256k1.PublicKey
}

//region Secp256k1PublicKey VerifyingKey

func (key *Secp256k1PublicKey) Verify(msg []byte, sig Signature) bool {
	switch sig.(type) {
	case *Secp256k1Signature:
		typedSig := sig.(*Secp256k1Signature)
		return typedSig.Inner.Verify(msg, key.Inner)
	default:
		return false
	}
}

//endregion

//region Secp256k1PublicKey CryptoMaterial

func (key *Secp256k1PublicKey) Bytes() []byte {
	return key.Inner.SerializeUncompressed()
}

func (key *Secp256k1PublicKey) FromBytes(bytes []byte) (err error) {
	key.Inner, err = secp256k1.ParsePubKey(bytes)
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
	pubKey, err := secp256k1.ParsePubKey(kb)
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
// Implements AuthenticatorImpl, bcs.Struct
// TODO: We might want a different interface for this one
type Secp256k1Authenticator struct {
	PubKey *Secp256k1PublicKey
	Sig    *Secp256k1Signature
}

//region Secp256k1Authenticator AuthenticatorImpl

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
	Inner *secp256k1.Signature
}

//region Secp256k1Signature CryptoMaterial

func (e *Secp256k1Signature) Bytes() []byte {
	// TODO: This library doesn't seem to work properly with the Rust implementation, the signatures are the wrong bytes
	// Golang for some reason outputs big ints as big endian, so we need to flip the bytes
	output := make([]byte, 64)
	r := e.Inner.GetR()
	//slices.Reverse(r)
	s := e.Inner.GetS()
	//slices.Reverse(s)
	copy(output[0:32], r.Bytes()[:])

	copy(output[32:64], s.Bytes()[:])
	return output
}

func (e *Secp256k1Signature) FromBytes(bytes []byte) (err error) {
	// We unfortunately have custom serialization, so we need custom deserialization here
	r := &big.Int{}
	r.SetBytes(bytes[0:32])
	s := &big.Int{}
	s.SetBytes(bytes[32:64])

	e.Inner = secp256k1.NewSignature(r, s)
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
