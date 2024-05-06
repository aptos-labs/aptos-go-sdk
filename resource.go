package aptos

type MoveResource struct {
	Tag   MoveStructTag
	Value map[string]any // MoveStructValue // TODO: api/types/src/move_types.rs probably actually has more to say about what a MoveStructValue is, but at first read it effectively says map[string]any; there's probably convention elesewhere about what goes into those 'any' parts
}

func (mr *MoveResource) MarshalBCS(bcs *Serializer) {
	mr.Tag.MarshalBCS(bcs)
	// We can't unmarshal `any`, BCS needs to know what destination struct type is
	panic("TODO")
}
func (mr *MoveResource) UnmarshalBCS(bcs *Deserializer) {
	mr.Tag.UnmarshalBCS(bcs)
	// We can't unmarshal `any`, BCS needs to know what destination struct type is
	panic("TODO")
}

type MoveStructTag struct {
	Address           AccountAddress
	Module            string
	Name              string
	GenericTypeParams []MoveType
}

func (mst *MoveStructTag) MarshalBCS(bcs *Serializer) {
	mst.Address.MarshalBCS(bcs)
	bcs.WriteString(mst.Module)
	bcs.WriteString(mst.Name)
	SerializeSequence(mst.GenericTypeParams, bcs)
}
func (mst *MoveStructTag) UnmarshalBCS(bcs *Deserializer) {
	mst.Address.UnmarshalBCS(bcs)
	mst.Module = bcs.ReadString()
	mst.Name = bcs.ReadString()
	mst.GenericTypeParams = DeserializeSequence[MoveType](bcs)
}

// enum
type MoveType uint8

const (
	MoveType_Bool             MoveType = 0
	MoveType_U8               MoveType = 1
	MoveType_U16              MoveType = 2
	MoveType_U32              MoveType = 3
	MoveType_U64              MoveType = 4
	MoveType_U128             MoveType = 5
	MoveType_U256             MoveType = 6
	MoveType_Address          MoveType = 7
	MoveType_Signer           MoveType = 8
	MoveType_Vector           MoveType = 9  // contains MoveType of items of vector
	MoveType_MoveStructTag    MoveType = 10 // contains a MoveStructTag
	MoveType_GeneritTypeParam MoveType = 11 // contains a uint16
	MoveType_Reference        MoveType = 12 // {mutable bool, to MoveType}
	MoveType_Unparsable       MoveType = 13 // contains a string
)

func (mt *MoveType) MarshalBCS(bcs *Serializer) {
	bcs.Uleb128(uint32(*mt))
}
func (mt *MoveType) UnmarshalBCS(bcs *Deserializer) {
	*mt = MoveType(bcs.Uleb128())
}
