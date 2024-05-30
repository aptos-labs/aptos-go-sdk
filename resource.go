package aptos

import (
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/types"
)

type MoveResource struct {
	Tag   MoveStructTag
	Value map[string]any // MoveStructValue // TODO: api/types/src/move_types.rs probably actually has more to say about what a MoveStructValue is, but at first read it effectively says map[string]any; there's probably convention elsewhere about what goes into those 'any' parts
}

func (mr *MoveResource) MarshalBCS(bcs *bcs.Serializer) {
	mr.Tag.MarshalBCS(bcs)
	// We can't unmarshal `any`, BCS needs to know what destination struct type is
	panic("TODO")
}
func (mr *MoveResource) UnmarshalBCS(bcs *bcs.Deserializer) {
	mr.Tag.UnmarshalBCS(bcs)
	// We can't unmarshal `any`, BCS needs to know what destination struct type is
	panic("TODO")
}

type MoveStructTag struct {
	Address           types.AccountAddress
	Module            string
	Name              string
	GenericTypeParams []MoveType
}

func (mst *MoveStructTag) MarshalBCS(bcs *bcs.Serializer) {
	mst.Address.MarshalBCS(bcs)
	bcs.WriteString(mst.Module)
	bcs.WriteString(mst.Name)

	for i := range mst.GenericTypeParams {
		bcs.Struct(&mst.GenericTypeParams[i])
	}
}
func (mst *MoveStructTag) UnmarshalBCS(deserializer *bcs.Deserializer) {
	mst.Address.UnmarshalBCS(deserializer)
	mst.Module = deserializer.ReadString()
	mst.Name = deserializer.ReadString()
	mst.GenericTypeParams = bcs.DeserializeSequence[MoveType](deserializer)
}

// MoveType enum
type MoveType uint8

const (
	MoveTypeBool             MoveType = 0
	MoveTypeU8               MoveType = 1
	MoveTypeU16              MoveType = 2
	MoveTypeU32              MoveType = 3
	MoveTypeU64              MoveType = 4
	MoveTypeU128             MoveType = 5
	MoveTypeU256             MoveType = 6
	MoveTypeAddress          MoveType = 7
	MoveTypeSigner           MoveType = 8
	MoveTypeVector           MoveType = 9  // contains MoveType of items of vector
	MoveTypeMoveStructTag    MoveType = 10 // contains a MoveStructTag
	MoveTypeGenericTypeParam MoveType = 11 // contains a uint16
	MoveTypeReference        MoveType = 12 // {mutable bool, to MoveType}
	MoveTypeUnparsable       MoveType = 13 // contains a string
)

func (mt *MoveType) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(*mt))
}
func (mt *MoveType) UnmarshalBCS(bcs *bcs.Deserializer) {
	*mt = MoveType(bcs.Uleb128())
}
