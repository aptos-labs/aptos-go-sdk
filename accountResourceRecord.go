package aptos

import (
	"github.com/qimeila/aptos-go-sdk/bcs"
)

// AccountResourceRecord DeserializeSequence[AccountResourceRecord](bcs) approximates the Rust side BTreeMap<StructTag,Vec<u8>>
// They should BCS the same with a prefix Uleb128 length followed by (StructTag,[]byte) pairs.
type AccountResourceRecord struct {
	// Account::Module::Name
	Tag StructTag

	// BCS data as stored by Move contract
	Data []byte
}

func (aar *AccountResourceRecord) MarshalBCS(ser *bcs.Serializer) {
	aar.Tag.MarshalBCS(ser)
	ser.WriteBytes(aar.Data)
}

func (aar *AccountResourceRecord) UnmarshalBCS(des *bcs.Deserializer) {
	aar.Tag.UnmarshalBCS(des)
	aar.Data = des.ReadBytes()
}
