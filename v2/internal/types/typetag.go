package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aptos-labs/aptos-go-sdk/v2/internal/bcs"
)

// TypeTagVariant identifies the type of a TypeTag.
type TypeTagVariant uint32

const (
	TypeTagBool      TypeTagVariant = 0
	TypeTagU8        TypeTagVariant = 1
	TypeTagU64       TypeTagVariant = 2
	TypeTagU128      TypeTagVariant = 3
	TypeTagAddress   TypeTagVariant = 4
	TypeTagSigner    TypeTagVariant = 5
	TypeTagVector    TypeTagVariant = 6
	TypeTagStruct    TypeTagVariant = 7
	TypeTagU16       TypeTagVariant = 8
	TypeTagU32       TypeTagVariant = 9
	TypeTagU256      TypeTagVariant = 10
	TypeTagI8        TypeTagVariant = 11
	TypeTagI16       TypeTagVariant = 12
	TypeTagI32       TypeTagVariant = 13
	TypeTagI64       TypeTagVariant = 14
	TypeTagI128      TypeTagVariant = 15
	TypeTagI256      TypeTagVariant = 16
	TypeTagGeneric   TypeTagVariant = 254
	TypeTagReference TypeTagVariant = 255
)

// TypeTagImpl is implemented by all specific type tag types.
type TypeTagImpl interface {
	bcs.Struct
	GetType() TypeTagVariant
	String() string
}

// TypeTag wraps a TypeTagImpl for serialization.
//
// Implements [bcs.Struct].
type TypeTag struct {
	Value TypeTagImpl
}

// String returns the canonical Move type string.
func (tt *TypeTag) String() string {
	if tt.Value == nil {
		return "<nil>"
	}
	return tt.Value.String()
}

// MarshalBCS serializes the TypeTag to BCS.
func (tt *TypeTag) MarshalBCS(ser *bcs.Serializer) {
	ser.Uleb128(uint32(tt.Value.GetType()))
	ser.Struct(tt.Value)
}

// UnmarshalBCS deserializes the TypeTag from BCS.
func (tt *TypeTag) UnmarshalBCS(des *bcs.Deserializer) {
	variant := TypeTagVariant(des.Uleb128())
	switch variant {
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
	case TypeTagI8:
		tt.Value = &I8Tag{}
	case TypeTagI16:
		tt.Value = &I16Tag{}
	case TypeTagI32:
		tt.Value = &I32Tag{}
	case TypeTagI64:
		tt.Value = &I64Tag{}
	case TypeTagI128:
		tt.Value = &I128Tag{}
	case TypeTagI256:
		tt.Value = &I256Tag{}
	case TypeTagAddress:
		tt.Value = &AddressTag{}
	case TypeTagSigner:
		tt.Value = &SignerTag{}
	case TypeTagVector:
		tt.Value = &VectorTag{}
	case TypeTagStruct:
		tt.Value = &StructTag{}
	default:
		des.SetError(fmt.Errorf("unknown TypeTag variant: %d", variant))
		return
	}
	des.Struct(tt.Value)
}

// Primitive type tags

type BoolTag struct{}

func (t *BoolTag) GetType() TypeTagVariant        { return TypeTagBool }
func (t *BoolTag) String() string                 { return "bool" }
func (t *BoolTag) MarshalBCS(*bcs.Serializer)     {}
func (t *BoolTag) UnmarshalBCS(*bcs.Deserializer) {}

type U8Tag struct{}

func (t *U8Tag) GetType() TypeTagVariant        { return TypeTagU8 }
func (t *U8Tag) String() string                 { return "u8" }
func (t *U8Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *U8Tag) UnmarshalBCS(*bcs.Deserializer) {}

type U16Tag struct{}

func (t *U16Tag) GetType() TypeTagVariant        { return TypeTagU16 }
func (t *U16Tag) String() string                 { return "u16" }
func (t *U16Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *U16Tag) UnmarshalBCS(*bcs.Deserializer) {}

type U32Tag struct{}

func (t *U32Tag) GetType() TypeTagVariant        { return TypeTagU32 }
func (t *U32Tag) String() string                 { return "u32" }
func (t *U32Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *U32Tag) UnmarshalBCS(*bcs.Deserializer) {}

type U64Tag struct{}

func (t *U64Tag) GetType() TypeTagVariant        { return TypeTagU64 }
func (t *U64Tag) String() string                 { return "u64" }
func (t *U64Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *U64Tag) UnmarshalBCS(*bcs.Deserializer) {}

type U128Tag struct{}

func (t *U128Tag) GetType() TypeTagVariant        { return TypeTagU128 }
func (t *U128Tag) String() string                 { return "u128" }
func (t *U128Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *U128Tag) UnmarshalBCS(*bcs.Deserializer) {}

type U256Tag struct{}

func (t *U256Tag) GetType() TypeTagVariant        { return TypeTagU256 }
func (t *U256Tag) String() string                 { return "u256" }
func (t *U256Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *U256Tag) UnmarshalBCS(*bcs.Deserializer) {}

type I8Tag struct{}

func (t *I8Tag) GetType() TypeTagVariant        { return TypeTagI8 }
func (t *I8Tag) String() string                 { return "i8" }
func (t *I8Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *I8Tag) UnmarshalBCS(*bcs.Deserializer) {}

type I16Tag struct{}

func (t *I16Tag) GetType() TypeTagVariant        { return TypeTagI16 }
func (t *I16Tag) String() string                 { return "i16" }
func (t *I16Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *I16Tag) UnmarshalBCS(*bcs.Deserializer) {}

type I32Tag struct{}

func (t *I32Tag) GetType() TypeTagVariant        { return TypeTagI32 }
func (t *I32Tag) String() string                 { return "i32" }
func (t *I32Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *I32Tag) UnmarshalBCS(*bcs.Deserializer) {}

type I64Tag struct{}

func (t *I64Tag) GetType() TypeTagVariant        { return TypeTagI64 }
func (t *I64Tag) String() string                 { return "i64" }
func (t *I64Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *I64Tag) UnmarshalBCS(*bcs.Deserializer) {}

type I128Tag struct{}

func (t *I128Tag) GetType() TypeTagVariant        { return TypeTagI128 }
func (t *I128Tag) String() string                 { return "i128" }
func (t *I128Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *I128Tag) UnmarshalBCS(*bcs.Deserializer) {}

type I256Tag struct{}

func (t *I256Tag) GetType() TypeTagVariant        { return TypeTagI256 }
func (t *I256Tag) String() string                 { return "i256" }
func (t *I256Tag) MarshalBCS(*bcs.Serializer)     {}
func (t *I256Tag) UnmarshalBCS(*bcs.Deserializer) {}

type AddressTag struct{}

func (t *AddressTag) GetType() TypeTagVariant        { return TypeTagAddress }
func (t *AddressTag) String() string                 { return "address" }
func (t *AddressTag) MarshalBCS(*bcs.Serializer)     {}
func (t *AddressTag) UnmarshalBCS(*bcs.Deserializer) {}

type SignerTag struct{}

func (t *SignerTag) GetType() TypeTagVariant        { return TypeTagSigner }
func (t *SignerTag) String() string                 { return "signer" }
func (t *SignerTag) MarshalBCS(*bcs.Serializer)     {}
func (t *SignerTag) UnmarshalBCS(*bcs.Deserializer) {}

// VectorTag represents vector<T> in Move.
type VectorTag struct {
	TypeParam TypeTag
}

func (t *VectorTag) GetType() TypeTagVariant { return TypeTagVector }

func (t *VectorTag) String() string {
	return "vector<" + t.TypeParam.String() + ">"
}

func (t *VectorTag) MarshalBCS(ser *bcs.Serializer) {
	ser.Struct(&t.TypeParam)
}

func (t *VectorTag) UnmarshalBCS(des *bcs.Deserializer) {
	des.Struct(&t.TypeParam)
}

// StructTag represents a struct type like address::module::Name<T1, T2>.
type StructTag struct {
	Address    AccountAddress
	Module     string
	Name       string
	TypeParams []TypeTag
}

func (t *StructTag) GetType() TypeTagVariant { return TypeTagStruct }

func (t *StructTag) String() string {
	var sb strings.Builder
	sb.WriteString(t.Address.String())
	sb.WriteString("::")
	sb.WriteString(t.Module)
	sb.WriteString("::")
	sb.WriteString(t.Name)
	if len(t.TypeParams) > 0 {
		sb.WriteRune('<')
		for i, tp := range t.TypeParams {
			if i > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(tp.String())
		}
		sb.WriteRune('>')
	}
	return sb.String()
}

func (t *StructTag) MarshalBCS(ser *bcs.Serializer) {
	t.Address.MarshalBCS(ser)
	ser.WriteString(t.Module)
	ser.WriteString(t.Name)
	// Serialize type params using function variant since TypeTag methods have pointer receivers
	bcs.SerializeSequenceFunc(ser, t.TypeParams, func(s *bcs.Serializer, tt TypeTag) {
		tt.MarshalBCS(s)
	})
}

func (t *StructTag) UnmarshalBCS(des *bcs.Deserializer) {
	t.Address.UnmarshalBCS(des)
	t.Module = des.ReadString()
	t.Name = des.ReadString()
	t.TypeParams = bcs.DeserializeSequenceFunc(des, func(d *bcs.Deserializer) TypeTag {
		var tt TypeTag
		tt.UnmarshalBCS(d)
		return tt
	})
}

// ReferenceTag represents a reference type &T in Move.
type ReferenceTag struct {
	TypeParam TypeTag
}

func (t *ReferenceTag) GetType() TypeTagVariant { return TypeTagReference }

func (t *ReferenceTag) String() string {
	return "&" + t.TypeParam.String()
}

func (t *ReferenceTag) MarshalBCS(*bcs.Serializer)     {}
func (t *ReferenceTag) UnmarshalBCS(*bcs.Deserializer) {}

// GenericTag represents a generic type parameter T0, T1, etc.
type GenericTag struct {
	Num uint64
}

func (t *GenericTag) GetType() TypeTagVariant { return TypeTagGeneric }

func (t *GenericTag) String() string {
	return "T" + strconv.FormatUint(t.Num, 10)
}

func (t *GenericTag) MarshalBCS(*bcs.Serializer)     {}
func (t *GenericTag) UnmarshalBCS(*bcs.Deserializer) {}

// Helper functions

// NewTypeTag wraps a TypeTagImpl in a TypeTag.
func NewTypeTag(inner TypeTagImpl) TypeTag {
	return TypeTag{Value: inner}
}

// NewVectorTag creates a vector<T> TypeTag.
func NewVectorTag(inner TypeTagImpl) *VectorTag {
	return &VectorTag{TypeParam: NewTypeTag(inner)}
}

// NewStringTag creates a 0x1::string::String TypeTag.
func NewStringTag() *StructTag {
	return &StructTag{
		Address:    AccountOne,
		Module:     "string",
		Name:       "String",
		TypeParams: nil,
	}
}

// NewOptionTag creates a 0x1::option::Option<T> TypeTag.
func NewOptionTag(inner TypeTagImpl) *StructTag {
	return &StructTag{
		Address:    AccountOne,
		Module:     "option",
		Name:       "Option",
		TypeParams: []TypeTag{NewTypeTag(inner)},
	}
}

// NewObjectTag creates a 0x1::object::Object<T> TypeTag.
func NewObjectTag(inner TypeTagImpl) *StructTag {
	return &StructTag{
		Address:    AccountOne,
		Module:     "object",
		Name:       "Object",
		TypeParams: []TypeTag{NewTypeTag(inner)},
	}
}

// AptosCoinTypeTag is the TypeTag for 0x1::aptos_coin::AptosCoin.
var AptosCoinTypeTag = TypeTag{Value: &StructTag{
	Address: AccountOne,
	Module:  "aptos_coin",
	Name:    "AptosCoin",
}}

// ParseTypeTag parses a Move type string into a TypeTag.
func ParseTypeTag(inputStr string) (*TypeTag, error) {
	inputRunes := []rune(inputStr)
	saved := make([]parseInfo, 0)
	innerTypes := make([]TypeTag, 0)
	curTypes := make([]TypeTag, 0)
	cur := 0
	currentStr := ""
	expectedTypes := 1

	for cur < len(inputRunes) {
		r := inputRunes[cur]

		switch r {
		case '<':
			saved = append(saved, parseInfo{
				expectedTypes: expectedTypes,
				types:         curTypes,
				str:           currentStr,
			})
			currentStr = ""
			curTypes = make([]TypeTag, 0)
			expectedTypes = 1
		case '>':
			if currentStr != "" {
				newType, err := parseTypeTagInner(currentStr, innerTypes)
				if err != nil {
					return nil, err
				}
				curTypes = append(curTypes, *newType)
			}

			if len(saved) == 0 {
				return nil, errors.New("unexpected '>'")
			}

			if expectedTypes != len(curTypes) {
				return nil, errors.New("type count mismatch")
			}

			savedPop := saved[len(saved)-1]
			saved = saved[:len(saved)-1]

			innerTypes = curTypes
			curTypes = savedPop.types
			currentStr = savedPop.str
			expectedTypes = savedPop.expectedTypes
		case ',':
			if len(saved) == 0 {
				return nil, errors.New("unexpected ',' at top level")
			}
			if currentStr == "" {
				return nil, errors.New("unexpected ','")
			}

			newType, err := parseTypeTagInner(currentStr, innerTypes)
			if err != nil {
				return nil, err
			}
			innerTypes = make([]TypeTag, 0)
			curTypes = append(curTypes, *newType)
			currentStr = ""
			expectedTypes++
		case ' ':
			parsedTypeTag := false
			if currentStr != "" {
				newType, err := parseTypeTagInner(currentStr, innerTypes)
				if err != nil {
					return nil, err
				}
				innerTypes = make([]TypeTag, 0)
				curTypes = append(curTypes, *newType)
				currentStr = ""
				parsedTypeTag = true
			}

			for cur < len(inputRunes) && inputRunes[cur] == ' ' {
				cur++
			}

			if cur < len(inputRunes) && parsedTypeTag {
				nextChar := inputRunes[cur]
				if nextChar != ',' && nextChar != '>' {
					return nil, errors.New("unexpected character after type")
				}
			}
			continue
		default:
			currentStr += string(r)
		}
		cur++
	}

	if len(saved) > 0 {
		return nil, errors.New("missing '>'")
	}

	switch len(curTypes) {
	case 0:
		return parseTypeTagInner(currentStr, innerTypes)
	case 1:
		if currentStr == "" {
			return &curTypes[0], nil
		}
		return nil, errors.New("unexpected ','")
	default:
		return nil, errors.New("unexpected whitespace")
	}
}

type parseInfo struct {
	expectedTypes int
	types         []TypeTag
	str           string
}

func parseTypeTagInner(input string, types []TypeTag) (*TypeTag, error) {
	str := strings.TrimSpace(input)

	// Check for primitive types
	switch str {
	case "bool":
		if len(types) > 0 {
			return nil, errors.New("bool cannot have type parameters")
		}
		return &TypeTag{Value: &BoolTag{}}, nil
	case "u8":
		if len(types) > 0 {
			return nil, errors.New("u8 cannot have type parameters")
		}
		return &TypeTag{Value: &U8Tag{}}, nil
	case "u16":
		if len(types) > 0 {
			return nil, errors.New("u16 cannot have type parameters")
		}
		return &TypeTag{Value: &U16Tag{}}, nil
	case "u32":
		if len(types) > 0 {
			return nil, errors.New("u32 cannot have type parameters")
		}
		return &TypeTag{Value: &U32Tag{}}, nil
	case "u64":
		if len(types) > 0 {
			return nil, errors.New("u64 cannot have type parameters")
		}
		return &TypeTag{Value: &U64Tag{}}, nil
	case "u128":
		if len(types) > 0 {
			return nil, errors.New("u128 cannot have type parameters")
		}
		return &TypeTag{Value: &U128Tag{}}, nil
	case "u256":
		if len(types) > 0 {
			return nil, errors.New("u256 cannot have type parameters")
		}
		return &TypeTag{Value: &U256Tag{}}, nil
	case "i8":
		if len(types) > 0 {
			return nil, errors.New("i8 cannot have type parameters")
		}
		return &TypeTag{Value: &I8Tag{}}, nil
	case "i16":
		if len(types) > 0 {
			return nil, errors.New("i16 cannot have type parameters")
		}
		return &TypeTag{Value: &I16Tag{}}, nil
	case "i32":
		if len(types) > 0 {
			return nil, errors.New("i32 cannot have type parameters")
		}
		return &TypeTag{Value: &I32Tag{}}, nil
	case "i64":
		if len(types) > 0 {
			return nil, errors.New("i64 cannot have type parameters")
		}
		return &TypeTag{Value: &I64Tag{}}, nil
	case "i128":
		if len(types) > 0 {
			return nil, errors.New("i128 cannot have type parameters")
		}
		return &TypeTag{Value: &I128Tag{}}, nil
	case "i256":
		if len(types) > 0 {
			return nil, errors.New("i256 cannot have type parameters")
		}
		return &TypeTag{Value: &I256Tag{}}, nil
	case "address":
		if len(types) > 0 {
			return nil, errors.New("address cannot have type parameters")
		}
		return &TypeTag{Value: &AddressTag{}}, nil
	case "signer":
		if len(types) > 0 {
			return nil, errors.New("signer cannot have type parameters")
		}
		return &TypeTag{Value: &SignerTag{}}, nil
	case "vector":
		if len(types) != 1 {
			return nil, fmt.Errorf("vector expects 1 type parameter, got %d", len(types))
		}
		return &TypeTag{Value: &VectorTag{TypeParam: types[0]}}, nil
	}

	// Check for reference
	if strings.HasPrefix(str, "&") {
		inner, err := parseTypeTagInner(strings.TrimPrefix(str, "&"), types)
		if err != nil {
			return nil, err
		}
		return &TypeTag{Value: &ReferenceTag{TypeParam: *inner}}, nil
	}

	// Check for generic
	if strings.HasPrefix(str, "T") {
		num, err := strconv.ParseUint(strings.TrimPrefix(str, "T"), 10, 8)
		if err != nil {
			return nil, err
		}
		return &TypeTag{Value: &GenericTag{Num: num}}, nil
	}

	// Must be a struct type
	parts := strings.Split(str, "::")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid struct type: %s", str)
	}

	address, err := ParseAddress(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid struct address: %s", parts[0])
	}

	module := parts[1]
	if matched, _ := regexp.MatchString("^[a-zA-Z_0-9]+$", module); !matched {
		return nil, fmt.Errorf("invalid module name: %s", module)
	}

	name := parts[2]
	if matched, _ := regexp.MatchString("^[a-zA-Z_0-9]+$", name); !matched {
		return nil, fmt.Errorf("invalid struct name: %s", name)
	}

	return &TypeTag{Value: &StructTag{
		Address:    address,
		Module:     module,
		Name:       name,
		TypeParams: types,
	}}, nil
}
