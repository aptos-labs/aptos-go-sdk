package aptos

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

func ConvertTypeTag(typeArg any) (*TypeTag, error) {
	switch typeArg.(type) {
	case TypeTag:
		tag := typeArg.(TypeTag)
		return &tag, nil
	case *TypeTag:
		tag := typeArg.(*TypeTag)
		if tag == nil {
			return nil, fmt.Errorf("invalid type tag %s, cannot be nil", typeArg)
		}
		return tag, nil
	case string:
		strTypeTag := typeArg.(string)
		return ParseTypeTag(strTypeTag)
	default:
		return nil, fmt.Errorf("invalid type tag type")
	}
}

func ConvertToU8(arg any) (*uint8, error) {
	var num uint8
	switch arg.(type) {
	case int:
		num = uint8(arg.(int))
	case uint:
		num = uint8(arg.(uint))
	case uint8:
		num = arg.(uint8)
	case big.Int:
		b := arg.(big.Int)
		num = uint8(b.Uint64())
	case *big.Int:
		b := arg.(*big.Int)
		if b == nil {
			return nil, fmt.Errorf("cannot convert to uint8, input is nil")
		}
		num = uint8(b.Uint64())
	case string:
		// Convert the number
		u64, err := strconv.ParseUint(arg.(string), 10, 8)
		if err != nil {
			return nil, err
		}
		num = uint8(u64)
	default:
		return nil, fmt.Errorf("invalid input type for uint8")
	}

	return &num, nil
}

func ConvertToU16(arg any) (*uint16, error) {
	var num uint16
	switch arg.(type) {
	case int:
		num = uint16(arg.(int))
	case uint:
		num = uint16(arg.(uint))
	case uint16:
		num = arg.(uint16)
	case big.Int:
		b := arg.(big.Int)
		num = uint16(b.Uint64())
	case *big.Int:
		b := arg.(*big.Int)
		if b == nil {
			return nil, fmt.Errorf("cannot convert to uint16, input is nil")
		}
		num = uint16(b.Uint64())
	case string:
		// Convert the number
		u64, err := strconv.ParseUint(arg.(string), 10, 16)
		if err != nil {
			return nil, err
		}
		num = uint16(u64)
	default:
		return nil, fmt.Errorf("invalid input type for uint16")
	}

	return &num, nil
}

func ConvertToU32(arg any) (*uint32, error) {
	var num uint32
	switch arg.(type) {
	case int:
		num = uint32(arg.(int))
	case uint:
		num = uint32(arg.(uint))
	case uint32:
		num = arg.(uint32)
	case big.Int:
		b := arg.(big.Int)
		num = uint32(b.Uint64())
	case *big.Int:
		b := arg.(*big.Int)
		if b == nil {
			return nil, fmt.Errorf("cannot convert to uint32, input is nil")
		}
		num = uint32(b.Uint64())
	case string:
		// Convert the number
		u64, err := strconv.ParseUint(arg.(string), 10, 32)
		if err != nil {
			return nil, err
		}
		num = uint32(u64)
	default:
		return nil, fmt.Errorf("invalid input type for uint32")
	}

	return &num, nil
}

func ConvertToU64(arg any) (*uint64, error) {
	var num uint64
	switch arg.(type) {
	case int:
		num = uint64(arg.(int))
	case uint:
		num = uint64(arg.(uint))
	case uint64:
		num = arg.(uint64)
	case big.Int:
		b := arg.(big.Int)
		num = b.Uint64()
	case *big.Int:
		b := arg.(*big.Int)
		if b == nil {
			return nil, fmt.Errorf("cannot convert to uint64, input is nil")
		}
		num = b.Uint64()
	case string:
		// Convert the number
		u64, err := strconv.ParseUint(arg.(string), 10, 64)
		if err != nil {
			return nil, err
		}
		num = u64
	default:
		return nil, fmt.Errorf("invalid input type for uint64")
	}

	return &num, nil
}

func ConvertToU128(arg any) (num *big.Int, err error) {
	switch arg.(type) {
	case int:
		num = big.NewInt(int64(arg.(int)))
	case uint:
		num = big.NewInt(int64(arg.(uint)))
	case big.Int:
		b := arg.(big.Int)
		num = &b
	case *big.Int:
		num = arg.(*big.Int)
		if num == nil {
			return nil, fmt.Errorf("cannot convert to uint128, input is nil")
		}
	case string:
		// Convert the number
		num, err = util.StrToBigInt(arg.(string))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid input type for uint128")
	}

	return num, nil
}

func ConvertToU256(arg any) (num *big.Int, err error) {
	switch arg.(type) {
	case int:
		num = big.NewInt(int64(arg.(int)))
	case uint:
		num = big.NewInt(int64(arg.(uint)))
	case big.Int:
		b := arg.(big.Int)
		num = &b
	case *big.Int:
		num = arg.(*big.Int)
		if num == nil {
			return nil, fmt.Errorf("cannot convert to uint256, input is nil")
		}
	case string:
		// Convert the number
		num, err = util.StrToBigInt(arg.(string))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid input type for uint256")
	}

	return num, nil
}

func ConvertToBool(arg any) (b bool, err error) {
	switch arg.(type) {
	case bool:
		b = arg.(bool)
	case string:
		switch arg.(string) {
		case "true":
			b = true
		case "false":
			b = false
		default:
			err = fmt.Errorf("invalid boolean input for bool")
		}
	default:
		err = fmt.Errorf("invalid input type for bool")
	}
	return b, err
}

func ConvertToAddress(arg any) (a *AccountAddress, err error) {
	switch arg.(type) {
	case AccountAddress:
		addr := arg.(AccountAddress)
		return &addr, nil
	case *AccountAddress:
		a = arg.(*AccountAddress)
		if a == nil {
			err = fmt.Errorf("invalid account address, nil")
		}
	case string:
		addr := AccountAddress{}
		err = addr.ParseStringRelaxed(arg.(string))
		if err != nil {
			return nil, err
		}
		a = &addr
	default:
		err = fmt.Errorf("invalid input type for address")
	}
	return a, err
}

// ConvertToVectorU8 returns the BCS encoded version of the bytes
func ConvertToVectorU8(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg.(type) {
	case string:
		bytes, err := util.ParseHex(arg.(string))
		if err != nil {
			return nil, err
		}
		// Serialize the bytes
		return bcs.SerializeBytes(bytes)
	case []byte:
		convertedArg := arg.([]byte)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to bytes")
		}
		return bcs.SerializeBytes(convertedArg)
	default:
		return nil, fmt.Errorf("invalid input type for vector<u8>")
	}
}

// ConvertToVectorU16 returns the BCS encoded version of the bytes
func ConvertToVectorU16(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg.(type) {
	case []uint16:
		convertedArg := arg.([]uint16)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(convertedArg, ser, func(serializer *bcs.Serializer, item uint16) {
				ser.U16(item)
			})
		})
	default:
		return nil, fmt.Errorf("invalid input type for vector<u16>")
	}
}

// ConvertToVectorU32 returns the BCS encoded version of the bytes
func ConvertToVectorU32(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg.(type) {
	case []uint32:
		convertedArg := arg.([]uint32)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(convertedArg, ser, func(serializer *bcs.Serializer, item uint32) {
				ser.U32(item)
			})
		})
	default:
		return nil, fmt.Errorf("invalid input type for vector<u32>")
	}
}

// ConvertToVectorU64 returns the BCS encoded version of the bytes
func ConvertToVectorU64(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg.(type) {
	case []uint64:
		convertedArg := arg.([]uint64)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(convertedArg, ser, func(serializer *bcs.Serializer, item uint64) {
				ser.U64(item)
			})
		})
	default:
		return nil, fmt.Errorf("invalid input type for vector<u64>")
	}
}

// ConvertToVectorU128 returns the BCS encoded version of the bytes
func ConvertToVectorU128(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg.(type) {
	case []big.Int:
		convertedArg := arg.([]big.Int)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(convertedArg, ser, func(serializer *bcs.Serializer, item big.Int) {
				ser.U128(item)
			})
		})
	default:
		return nil, fmt.Errorf("invalid input type for vector<u128>")
	}
}

// ConvertToVectorU256 returns the BCS encoded version of the bytes
func ConvertToVectorU256(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg.(type) {
	case []big.Int:
		convertedArg := arg.([]big.Int)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(convertedArg, ser, func(serializer *bcs.Serializer, item big.Int) {
				ser.U256(item)
			})
		})
	default:
		return nil, fmt.Errorf("invalid input type for vector<u256>")
	}
}

// ConvertToVectorBool returns the BCS encoded version of the boolean vector
func ConvertToVectorBool(arg any) (b []byte, err error) {
	switch arg.(type) {
	case []bool:
		convertedArg := arg.([]bool)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to bool vector")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(convertedArg, ser, func(serializer *bcs.Serializer, item bool) {
				ser.Bool(item)
			})
		})
	default:
		return nil, fmt.Errorf("invalid input type for vector<bool>")
	}
}

// ConvertToVectorAddress returns the BCS encoded version of the address vector
func ConvertToVectorAddress(arg any) (b []byte, err error) {
	switch arg.(type) {
	case []AccountAddress:
		convertedArg := arg.([]AccountAddress)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to address vector")
		}
		return bcs.SerializeSequenceOnly(convertedArg)
	case []*AccountAddress:
		convertedArg := arg.([]*AccountAddress)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to address vector")
		}
		// Convert []*AccountAddress to []AccountAddress
		addresses := make([]AccountAddress, len(convertedArg))
		for i, addr := range convertedArg {
			if addr == nil {
				return nil, fmt.Errorf("cannot convert nil address in vector")
			}
			addresses[i] = *addr
		}
		return bcs.SerializeSequenceOnly(addresses)
	default:
		return nil, fmt.Errorf("invalid input type for vector<address>")
	}
}

// ConvertToVectorGeneric returns the BCS encoded version of the generic vector
func ConvertToVectorGeneric(typeArg TypeTag, arg any, generics []TypeTag) (b []byte, err error) {
	genericTag, ok := typeArg.Value.(*GenericTag)
	if !ok {
		return nil, fmt.Errorf("invalid type tag for generic vector")
	}

	if genericTag.Num >= uint64(len(generics)) {
		return nil, fmt.Errorf("generic number out of bounds")
	}

	innerType := generics[genericTag.Num]

	switch arg.(type) {
	case []any:
		convertedArg := arg.([]any)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to generic vector")
		}

		// Serialize length
		lengthBytes, err := bcs.SerializeUleb128(uint32(len(convertedArg)))
		if err != nil {
			return nil, err
		}
		b = lengthBytes

		// Serialize each element
		for _, item := range convertedArg {
			val, err := ConvertArg(innerType, item, generics)
			if err != nil {
				return nil, err
			}
			b = append(b, val...)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("invalid input type for generic vector")
	}
}

// ConvertToVectorReference returns the BCS encoded version of the reference vector
func ConvertToVectorReference(typeArg TypeTag, arg any, generics []TypeTag) (b []byte, err error) {
	refTag, ok := typeArg.Value.(*ReferenceTag)
	if !ok {
		return nil, fmt.Errorf("invalid type tag for reference vector")
	}

	innerType := refTag.TypeParam

	switch arg.(type) {
	case []any:
		convertedArg := arg.([]any)
		if convertedArg == nil {
			return nil, fmt.Errorf("cannot convert nil to reference vector")
		}

		// Serialize length
		lengthBytes, err := bcs.SerializeUleb128(uint32(len(convertedArg)))
		if err != nil {
			return nil, err
		}
		b = lengthBytes

		// Serialize each element
		for _, item := range convertedArg {
			val, err := ConvertArg(innerType, item, generics)
			if err != nil {
				return nil, err
			}
			b = append(b, val...)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("invalid input type for reference vector")
	}
}

func ConvertToVector(typeArg TypeTag, arg any, generics []TypeTag) (out []byte, err error) {
	// We have to switch based on type, thanks Golang
	switch typeArg.Value.(type) {
	case *U8Tag:
		return ConvertToVectorU8(arg)
	case *U16Tag:
		return ConvertToVectorU16(arg)
	case *U32Tag:
		return ConvertToVectorU32(arg)
	case *U64Tag:
		return ConvertToVectorU64(arg)
	case *U128Tag:
		return ConvertToVectorU128(arg)
	case *U256Tag:
		return ConvertToVectorU256(arg)
	case *BoolTag:
		return ConvertToVectorBool(arg)
	case *AddressTag:
		return ConvertToVectorAddress(arg)
	case *SignerTag:
		return ConvertToVectorAddress(arg)
	case *GenericTag:
		return ConvertToVectorGeneric(typeArg, arg, generics)
	case *ReferenceTag:
		return ConvertToVectorReference(typeArg, arg, generics)
	// TODO: Handle structs
	default:
		return nil, fmt.Errorf("%s is currently not supported as an input type", typeArg.String())
	}
}

func ConvertArg(typeArg TypeTag, arg any, generics []TypeTag) (b []byte, err error) {
	switch typeArg.Value.(type) {
	case *U8Tag:
		num, err := ConvertToU8(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU8(*num)
	case *U16Tag:
		num, err := ConvertToU16(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU16(*num)
	case *U32Tag:
		num, err := ConvertToU32(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU32(*num)
	case *U64Tag:
		num, err := ConvertToU64(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU64(*num)
	case *U128Tag:
		num, err := ConvertToU128(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU128(*num)
	case *U256Tag:
		num, err := ConvertToU256(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU256(*num)
	case *BoolTag:
		bo, err := ConvertToBool(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeBool(bo)
	case *AddressTag:
		a, err := ConvertToAddress(arg)
		if err != nil {
			return nil, err
		}
		return bcs.Serialize(a)
	case *GenericTag:
		// Convert based on number
		genericTag := typeArg.Value.(*GenericTag)
		genericNum := genericTag.Num
		if genericNum >= uint64(len(generics)) {
			return nil, fmt.Errorf("generic number out of bounds")
		}

		tag := generics[genericTag.Num]
		return ConvertArg(tag, arg, generics)
	case *ReferenceTag:
		// Convert based on inner type
		refTag := typeArg.Value.(*ReferenceTag)
		return ConvertArg(refTag.TypeParam, arg, generics)
	case *VectorTag:
		// This has two paths:
		// 1. Hex strings are allowed for vector<u8>
		// 2. Otherwise, everything is just parsed as an array of the inner type
		vecTag := typeArg.Value.(*VectorTag)
		switch vecTag.TypeParam.Value.(type) {
		case *U8Tag:
			return ConvertToVectorU8(arg)
		default:
			return ConvertToVector(vecTag.TypeParam, arg, generics)
		}
	case *StructTag:
		structTag := typeArg.Value.(*StructTag)
		// TODO: We should be able to support custom structs, but for now only support known
		switch structTag.Address {
		case AccountOne:
			switch structTag.Module {
			case "object":
				switch structTag.Name {
				case "Object":
					// TODO: Move to function
					// Handle as address, inner type doesn't matter
					// TODO: Improve error message
					return ConvertArg(TypeTag{&AddressTag{}}, arg, generics)
				}
			case "string":
				switch structTag.Name {
				case "String":
					// Handle as string, we won't let bytes as an input for now here
					switch arg.(type) {
					case string:
						return bcs.SerializeBytes([]byte(arg.(string)))
					default:
						return nil, fmt.Errorf("invalid input type for 0x1::string::String")
					}
				}
			case "option":
				switch structTag.Name {
				case "Option":
					// Check it has the proper inner type
					if 1 != len(structTag.TypeParams) {
						return nil, fmt.Errorf("invalid input type for option, must have exactly one type arg")
					}
					// Get inner type
					typeParam := structTag.TypeParams[0]

					// Handle special case of "none", it's a single 0 byte
					if arg == nil {
						return bcs.SerializeU8(0)
					}

					// Otherwise, it's a single byte 1, and the encoded arg
					b, err = ConvertArg(typeParam, arg, generics)
					if err != nil {
						return nil, err
					}
					return append([]byte{1}, b...), nil
				}
			}
		}
	default:
		return nil, fmt.Errorf("unknown type argument")
	}

	return nil, fmt.Errorf("failed to convert type argument")
}
