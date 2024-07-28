package types

import (
	"fmt"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

//region TypeTag

// TypeTagVariant is an enum representing the different types of TypeTag
type TypeTagVariant uint32

const (
	TypeTagBool    TypeTagVariant = 0  // Represents the bool type in Move BoolTag
	TypeTagU8      TypeTagVariant = 1  // Represents the u8 type in Move U8Tag
	TypeTagU64     TypeTagVariant = 2  // Represents the u64 type in Move U64Tag
	TypeTagU128    TypeTagVariant = 3  // Represents the u128 type in Move U128Tag
	TypeTagAddress TypeTagVariant = 4  // Represents the address type in Move AddressTag
	TypeTagSigner  TypeTagVariant = 5  // Represents the signer type in Move SignerTag
	TypeTagVector  TypeTagVariant = 6  // Represents the vector type in Move VectorTag
	TypeTagStruct  TypeTagVariant = 7  // Represents the struct type in Move StructTag
	TypeTagU16     TypeTagVariant = 8  // Represents the u16 type in Move U16Tag
	TypeTagU32     TypeTagVariant = 9  // Represents the u32 type in Move U32Tag
	TypeTagU256    TypeTagVariant = 10 // Represents the u256 type in Move U256Tag
)

// TypeTagImpl is an interface describing all the different types of [TypeTag].  Unfortunately because of how serialization
// works, a wrapper TypeTag struct is needed to handle the differentiation between types
type TypeTagImpl interface {
	bcs.Struct
	// GetType returns the TypeTagVariant for this [TypeTag]
	GetType() TypeTagVariant
	// String returns the canonical Move string representation of this [TypeTag]
	String() string
}

// TypeTag is a wrapper around a [TypeTagImpl] e.g. [BoolTag] or [U8Tag] for the purpose of serialization and deserialization
// Implements:
//   - [bcs.Struct]
type TypeTag struct {
	Value TypeTagImpl
}

// String gives the canonical TypeTag string value used in Move
func (tt *TypeTag) String() string {
	return tt.Value.String()
}

//region TypeTag bcs.Struct

// MarshalBCS serializes the TypeTag to bytes
//
// Implements:
//   - [bcs.Marshaler]
func (tt *TypeTag) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(tt.Value.GetType()))
	ser.Struct(tt.Value)
}

// UnmarshalBCS deserializes the TypeTag from bytes
//
// Implements:
//   - [bcs.Unmarshaler]
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

// SignerTag represents the signer type in Move
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

// AddressTag represents the address type in Move
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

// BoolTag represents the bool type in Move
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

// U8Tag represents the u8 type in Move
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

// U16Tag represents the u16 type in Move
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

// U32Tag represents the u32 type in Move
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

// U64Tag represents the u64 type in Move
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

// U128Tag represents the u128 type in Move
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

// U256Tag represents the u256 type in Move
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

// VectorTag represents the vector<T> type in Move, where T is another [TypeTag]
type VectorTag struct {
	TypeParam TypeTag // TypeParam is the type of the elements in the vector
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

// StructTag represents an on-chain struct of the form address::module::name<T1,T2,...> and each T is a [TypeTag]
type StructTag struct {
	Address    AccountAddress // Address is the address of the module
	Module     string         // Module is the name of the module
	Name       string         // Name is the name of the struct
	TypeParams []TypeTag      // TypeParams are the TypeTags of the type parameters
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
