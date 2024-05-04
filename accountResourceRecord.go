package aptos

// AccountResourceRecord DeserializeSequence[AccountResourceRecord](bcs) approximates the Rust side BTreeMap<StructTag,Vec<u8>>
// They should BCS the same with a prefix Uleb128 length followed by (StructTag,[]byte) pairs.
type AccountResourceRecord struct {
	// Account::Module::Name
	Tag StructTag

	// BCS data as stored by Move contract
	Data []byte
}

func (aar *AccountResourceRecord) MarshalBCS(bcs *Serializer) {
	aar.Tag.MarshalBCS(bcs)
	bcs.WriteBytes(aar.Data)
}
func (aar *AccountResourceRecord) UnmarshalBCS(bcs *Deserializer) {
	aar.Tag.UnmarshalBCS(bcs)
	aar.Data = bcs.ReadBytes()
}
