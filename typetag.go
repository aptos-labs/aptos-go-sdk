package aptos

import (
	"fmt"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/core"
)

type TypeTagType uint64

const (
	TypeTag_Bool           TypeTagType = 0
	TypeTag_U8             TypeTagType = 1
	TypeTag_U64            TypeTagType = 2
	TypeTag_U128           TypeTagType = 3
	TypeTag_AccountAddress TypeTagType = 4
	TypeTag_Signer         TypeTagType = 5
	TypeTag_Vector         TypeTagType = 6
	TypeTag_Struct         TypeTagType = 7
	TypeTag_U16            TypeTagType = 8
	TypeTag_U32            TypeTagType = 9
	TypeTag_U256           TypeTagType = 10
)

type TypeTagImpl interface {
	bcs.Struct
	GetType() TypeTagType
	String() string
}

type TypeTag struct {
	Value TypeTagImpl
}

func (tt *TypeTag) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Uleb128(uint32(tt.Value.GetType()))
	tt.Value.MarshalBCS(bcs)

}
func (tt *TypeTag) UnmarshalBCS(bcs *bcs.Deserializer) {
	variant := bcs.Uleb128()
	switch TypeTagType(variant) {
	case TypeTag_Bool:
		xt := &BoolTag{}
		xt.UnmarshalBCS(bcs)
		tt.Value = xt
	case TypeTag_U8:
		xt := &U8Tag{}
		xt.UnmarshalBCS(bcs)
		tt.Value = xt
	case TypeTag_U16:
		xt := &U16Tag{}
		xt.UnmarshalBCS(bcs)
		tt.Value = xt
	case TypeTag_U32:
		xt := &U32Tag{}
		xt.UnmarshalBCS(bcs)
		tt.Value = xt
	case TypeTag_U64:
		xt := &U64Tag{}
		xt.UnmarshalBCS(bcs)
		tt.Value = xt
	case TypeTag_Struct:
		xt := &StructTag{}
		xt.UnmarshalBCS(bcs)
		tt.Value = xt
	default:
		bcs.SetError(fmt.Errorf("unknown TypeTag enum %d", variant))
	}
}

func (tt *TypeTag) String() string {
	return tt.Value.String()
}

func NewTypeTag(v any) *TypeTag {
	switch tv := v.(type) {
	case uint8:
		return &TypeTag{
			Value: &U8Tag{Value: tv},
		}
	}
	return nil
}

type BoolTag struct {
	Value bool
}

func (xt *BoolTag) String() string {
	return "bool"
}

func (xt *BoolTag) GetType() TypeTagType {
	return TypeTag_Bool
}

func (xt *BoolTag) MarshalBCS(bcs *bcs.Serializer) {
	bcs.Bool(xt.Value)

}
func (xt *BoolTag) UnmarshalBCS(bcs *bcs.Deserializer) {
	xt.Value = bcs.Bool()
}

type U8Tag struct {
	Value uint8
}

func (xt *U8Tag) String() string {
	return "u8"
}

func (xt *U8Tag) GetType() TypeTagType {
	return TypeTag_U8
}

func (xt *U8Tag) MarshalBCS(bcs *bcs.Serializer) {
	bcs.U8(xt.Value)

}
func (xt *U8Tag) UnmarshalBCS(bcs *bcs.Deserializer) {
	xt.Value = bcs.U8()
}

type U16Tag struct {
	Value uint16
}

func (xt *U16Tag) String() string {
	return "u16"
}

func (xt *U16Tag) GetType() TypeTagType {
	return TypeTag_U16
}

func (xt *U16Tag) MarshalBCS(bcs *bcs.Serializer) {
	bcs.U16(xt.Value)

}
func (xt *U16Tag) UnmarshalBCS(bcs *bcs.Deserializer) {
	xt.Value = bcs.U16()
}

type U32Tag struct {
	Value uint32
}

func (xt *U32Tag) String() string {
	return "u32"
}

func (xt *U32Tag) GetType() TypeTagType {
	return TypeTag_U32
}

func (xt *U32Tag) MarshalBCS(bcs *bcs.Serializer) {
	bcs.U32(xt.Value)

}
func (xt *U32Tag) UnmarshalBCS(bcs *bcs.Deserializer) {
	xt.Value = bcs.U32()
}

type U64Tag struct {
	Value uint64
}

func (xt *U64Tag) String() string {
	return "u64"
}

func (xt *U64Tag) GetType() TypeTagType {
	return TypeTag_U64
}

func (xt *U64Tag) MarshalBCS(bcs *bcs.Serializer) {
	bcs.U64(xt.Value)

}
func (xt *U64Tag) UnmarshalBCS(bcs *bcs.Deserializer) {
	xt.Value = bcs.U64()
}

type AccountAddressTag struct {
	Value core.AccountAddress
}

func (xt *AccountAddressTag) GetType() TypeTagType {
	return TypeTag_AccountAddress
}

func (xt *AccountAddressTag) MarshalBCS(bcs *bcs.Serializer) {
	xt.Value.MarshalBCS(bcs)

}
func (xt *AccountAddressTag) UnmarshalBCS(bcs *bcs.Deserializer) {
	xt.Value.UnmarshalBCS(bcs)
}

type StructTag struct {
	Address    core.AccountAddress
	Module     string
	Name       string
	TypeParams []TypeTag
}

func (st *StructTag) MarshalBCS(serializer *bcs.Serializer) {
	st.Address.MarshalBCS(serializer)
	serializer.WriteString(st.Module)
	serializer.WriteString(st.Name)
	bcs.SerializeSequence(st.TypeParams, serializer)
}
func (st *StructTag) UnmarshalBCS(deserializer *bcs.Deserializer) {
	st.Address.UnmarshalBCS(deserializer)
	st.Module = deserializer.ReadString()
	st.Name = deserializer.ReadString()
	st.TypeParams = bcs.DeserializeSequence[TypeTag](deserializer)
}

func (st *StructTag) GetType() TypeTagType {
	return TypeTag_Struct
}

func (st *StructTag) String() string {
	out := strings.Builder{}
	out.WriteString(st.Address.String())
	out.WriteString("::")
	out.WriteString(st.Module)
	out.WriteString("::")
	out.WriteString(st.Name)
	if len(st.TypeParams) != 0 {
		out.WriteRune('<')
		for i, tp := range st.TypeParams {
			if i != 0 {
				out.WriteRune(',')
			}
			out.WriteString(tp.String())
		}
		out.WriteRune('>')
	}
	return out.String()
}
