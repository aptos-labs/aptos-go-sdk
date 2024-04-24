package aptos

import "fmt"

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
	BCSStruct
	GetType() TypeTagType
}

type TypeTag struct {
	Value TypeTagImpl
}

func (tt *TypeTag) MarshalBCS(bcs *Serializer) {
	bcs.Uleb128(uint64(tt.Value.GetType()))
	tt.Value.MarshalBCS(bcs)

}
func (tt *TypeTag) UnmarshalBCS(bcs *Deserializer) {
	variant := bcs.Uleb128()
	switch TypeTagType(variant) {
	case TypeTag_Bool:
		xt := &BoolTag{}
		xt.UnmarshalBCS(bcs)
		tt.Value = xt
	default:
		bcs.SetError(fmt.Errorf("unknown TypeTag enum %d", variant))
	}
}

type BoolTag struct {
	Value bool
}

func (xt *BoolTag) GetType() TypeTagType {
	return TypeTag_Bool
}

func (xt *BoolTag) MarshalBCS(bcs *Serializer) {
	bcs.Bool(xt.Value)

}
func (xt *BoolTag) UnmarshalBCS(bcs *Deserializer) {
	xt.Value = bcs.Bool()
}
