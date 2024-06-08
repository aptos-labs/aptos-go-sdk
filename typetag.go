package aptos

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"strings"
)

//region TypeTag

type TypeTagVariant uint32

const (
	TypeTagBool    TypeTagVariant = 0
	TypeTagU8      TypeTagVariant = 1
	TypeTagU64     TypeTagVariant = 2
	TypeTagU128    TypeTagVariant = 3
	TypeTagAddress TypeTagVariant = 4
	TypeTagSigner  TypeTagVariant = 5
	TypeTagVector  TypeTagVariant = 6
	TypeTagStruct  TypeTagVariant = 7
	TypeTagU16     TypeTagVariant = 8
	TypeTagU32     TypeTagVariant = 9
	TypeTagU256    TypeTagVariant = 10
)

// TypeTagImpl is an interface describing all the different types of TypeTag.  Unfortunately because of how serialization
// works, a wrapper TypeTag struct is needed to handle the differentiation between types
type TypeTagImpl interface {
	bcs.Struct
	GetType() TypeTagVariant
	String() string
}

// TypeTag is a wrapper around a TypeTagImpl e.g. BoolTag or U8Tag for the purpose of serialization and deserialization
// Implements bcs.Struct
type TypeTag struct {
	Value TypeTagImpl
}

// String gives the canonical TypeTag string value used in Move
func (tt *TypeTag) String() string {
	return tt.Value.String()
}

//region TypeTag bcs.Struct

func (tt *TypeTag) MarshalBCS(ser *bcs.Serializer) {
	if tt.Value == nil {
		ser.SetError(fmt.Errorf("nil TypeTag"))
		return
	}
	ser.Uleb128(uint32(tt.Value.GetType()))
	ser.Struct(tt.Value)
}

func (tt *TypeTag) UnmarshalBCS(des *bcs.Deserializer) {
	variant := TypeTagVariant(des.Uleb128())
	switch variant {
	case TypeTagAddress:
		tt.Value = &AddressTag{}
	case TypeTagSigner:
		tt.Value = &SignerTag{}
	case TypeTagBool:
		tt.Value = &BoolTag{}
	case TypeTagU8:
		tt.Value = &U8Tag{}
	case TypeTagU16:
		tt.Value = &U16Tag{}
	case TypeTagU32:
		tt.Value = &U32Tag{}
	case TypeTagU64:
		tt.Value = &U64Tag{}
	case TypeTagU128:
		tt.Value = &U128Tag{}
	case TypeTagU256:
		tt.Value = &U256Tag{}
	case TypeTagVector:
		tt.Value = &VectorTag{}
	case TypeTagStruct:
		tt.Value = &StructTag{}
	default:
		des.SetError(fmt.Errorf("unknown TypeTag enum %d", variant))
		return
	}
	des.Struct(tt.Value)
}

//endregion
//endregion

//region SignerTag

type SignerTag struct{}

//region SignerTag TypeTagImpl

func (xt *SignerTag) String() string {
	return "signer"
}

func (xt *SignerTag) GetType() TypeTagVariant {
	return TypeTagSigner
}

//endregion

//region SignerTag bcs.Struct

func (xt *SignerTag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *SignerTag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region AddressTag

type AddressTag struct{}

//region AddressTag TypeTagImpl

func (xt *AddressTag) String() string {
	return "address"
}

func (xt *AddressTag) GetType() TypeTagVariant {
	return TypeTagAddress
}

//endregion

//region AddressTag bcs.Struct

func (xt *AddressTag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *AddressTag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region BoolTag

type BoolTag struct{}

//region BoolTag TypeTagImpl

func (xt *BoolTag) String() string {
	return "bool"
}

func (xt *BoolTag) GetType() TypeTagVariant {
	return TypeTagBool
}

//endregion

//region BoolTag bcs.struct

func (xt *BoolTag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *BoolTag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region U8Tag

type U8Tag struct{}

//region U8Tag TypeTagImpl

func (xt *U8Tag) String() string {
	return "u8"
}

func (xt *U8Tag) GetType() TypeTagVariant {
	return TypeTagU8
}

//endregion

//region U8Tag bcs.Struct

func (xt *U8Tag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *U8Tag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region U16Tag

type U16Tag struct{}

//region U16Tag TypeTagImpl

func (xt *U16Tag) String() string {
	return "u16"
}

func (xt *U16Tag) GetType() TypeTagVariant {
	return TypeTagU16
}

//endregion

//region U16Tag bcs.Struct

func (xt *U16Tag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *U16Tag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region U32Tag

type U32Tag struct{}

//region U32Tag TypeTagImpl

func (xt *U32Tag) String() string {
	return "u32"
}

func (xt *U32Tag) GetType() TypeTagVariant {
	return TypeTagU32
}

//endregion

//region U32Tag bcs.Struct

func (xt *U32Tag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *U32Tag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region U64Tag

type U64Tag struct{}

//region U64Tag TypeTagImpl

func (xt *U64Tag) String() string {
	return "u64"
}

func (xt *U64Tag) GetType() TypeTagVariant {
	return TypeTagU64
}

//endregion

//region U64Tag bcs.Struct

func (xt *U64Tag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *U64Tag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region U128Tag

type U128Tag struct{}

//region U128Tag TypeTagImpl

func (xt *U128Tag) String() string {
	return "u128"
}

func (xt *U128Tag) GetType() TypeTagVariant {
	return TypeTagU128
}

//endregion

//region U128Tag bcs.Struct

func (xt *U128Tag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *U128Tag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region U256Tag

type U256Tag struct{}

//region U256Tag TypeTagImpl

func (xt *U256Tag) String() string {
	return "u256"
}

func (xt *U256Tag) GetType() TypeTagVariant {
	return TypeTagU256
}

//endregion

//region U256Tag bcs.Struct

func (xt *U256Tag) MarshalBCS(_ *bcs.Serializer)     {}
func (xt *U256Tag) UnmarshalBCS(_ *bcs.Deserializer) {}

//endregion
//endregion

//region VectorTag

// VectorTag represents the vector<T> type in Move, where T is another TypeTag
type VectorTag struct {
	TypeParam TypeTag
}

//region VectorTag TypeTagImpl

func (xt *VectorTag) GetType() TypeTagVariant {
	return TypeTagVector
}

func (xt *VectorTag) String() string {
	out := strings.Builder{}
	out.WriteString("vector<")
	out.WriteString(xt.TypeParam.Value.String())
	out.WriteString(">")
	return out.String()
}

//endregion

//region TypeTagVector bcs.Struct

func (xt *VectorTag) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(&xt.TypeParam)
}

func (xt *VectorTag) UnmarshalBCS(des *bcs.Deserializer) {
	var tag TypeTag
	tag.UnmarshalBCS(des)
	xt.TypeParam = tag
}

//endregion
//endregion

//region StructTag

// StructTag represents an on-chain struct of the form address::module::name<T1,T2,...>
type StructTag struct {
	Address    AccountAddress
	Module     string
	Name       string
	TypeParams []TypeTag
}

//region StructTag TypeTagImpl

func (xt *StructTag) GetType() TypeTagVariant {
	return TypeTagStruct
}

// String outputs to the form address::module::name<type1, type2> e.g.
// 0x1::string::String or 0x42::my_mod::MultiType<u8,0x1::string::String>
func (xt *StructTag) String() string {
	out := strings.Builder{}
	out.WriteString(xt.Address.String())
	out.WriteString("::")
	out.WriteString(xt.Module)
	out.WriteString("::")
	out.WriteString(xt.Name)
	if len(xt.TypeParams) != 0 {
		out.WriteRune('<')
		for i, tp := range xt.TypeParams {
			if i != 0 {
				out.WriteRune(',')
			}
			out.WriteString(tp.String())
		}
		out.WriteRune('>')
	}
	return out.String()
}

//endregion

//region StructTag bcs.Struct

func (xt *StructTag) MarshalBCS(ser *bcs.Serializer) {
	xt.Address.MarshalBCS(ser)
	ser.WriteString(xt.Module)
	ser.WriteString(xt.Name)
	bcs.SerializeSequence(xt.TypeParams, ser)
}
func (xt *StructTag) UnmarshalBCS(des *bcs.Deserializer) {
	xt.Address.UnmarshalBCS(des)
	xt.Module = des.ReadString()
	xt.Name = des.ReadString()
	xt.TypeParams = bcs.DeserializeSequence[TypeTag](des)
}

//endregion
//endregion

//region TypeTag helpers

// NewTypeTag wraps a TypeTagImpl in a TypeTag
func NewTypeTag(inner TypeTagImpl) TypeTag {
	return TypeTag{
		Value: inner,
	}
}

// NewVectorTag creates a TypeTag for vector<inner>
func NewVectorTag(inner TypeTagImpl) *VectorTag {
	return &VectorTag{
		TypeParam: NewTypeTag(inner),
	}
}

// NewStringTag creates a TypeTag for 0x1::string::String
func NewStringTag() *StructTag {
	return &StructTag{
		Address:    AccountOne,
		Module:     "string",
		Name:       "String",
		TypeParams: []TypeTag{},
	}
}

// NewOptionTag creates a 0x1::option::Option TypeTag based on an inner type
func NewOptionTag(inner TypeTagImpl) *StructTag {
	return &StructTag{
		Address:    AccountOne,
		Module:     "option",
		Name:       "Option",
		TypeParams: []TypeTag{NewTypeTag(inner)},
	}
}

// NewObjectTag creates a 0x1::object::Object TypeTag based on an inner type
func NewObjectTag(inner TypeTagImpl) *StructTag {
	return &StructTag{
		Address:    AccountOne,
		Module:     "object",
		Name:       "Object",
		TypeParams: []TypeTag{NewTypeTag(inner)},
	}
}

// AptosCoinTypeTag is the TypeTag for 0x1::aptos_coin::AptosCoin
var AptosCoinTypeTag = TypeTag{&StructTag{
	Address: AccountOne,
	Module:  "aptos_coin",
	Name:    "AptosCoin",
}}

//endregion
