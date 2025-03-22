package aptos

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
)

func EntryFunctionFromAbi(abi any, moduleAddress AccountAddress, moduleName string, functionName string, typeArgs []any, args []any) (entry *EntryFunction, err error) {
	var function *api.MoveFunction
	switch abi := abi.(type) {
	case *api.MoveModule:
		moduleAbi := abi
		// Find function
		for _, fun := range moduleAbi.ExposedFunctions {
			if fun.Name == functionName {
				if !fun.IsEntry {
					return nil, fmt.Errorf("function %s is not a entry function in module %s", functionName, moduleAbi.Name)
				}

				function = fun
				break
			}
		}
	case *api.MoveFunction:
		function = abi
	default:
		return nil, fmt.Errorf("unknown abi type: %T", abi)
	}

	if function == nil {
		return nil, fmt.Errorf("entry function %s not found in module %s", functionName, moduleName)
	}

	// Check type args length matches
	if len(typeArgs) != len(function.GenericTypeParams) {
		return nil, fmt.Errorf("entry function %s does not have the correct number of type arguments for function %s", functionName, functionName)
	}

	// Convert TypeTag, *TypeTag, and string to TypeTag
	// TODO: Check properties of generic type?
	convertedTypeArgs := make([]TypeTag, len(typeArgs))
	for i, typeArg := range typeArgs {
		tag, err := ConvertTypeTag(typeArg)
		if err != nil {
			return nil, err
		}
		convertedTypeArgs[i] = *tag
	}

	// Convert string types to actual types
	argTypes := make([]TypeTag, 0)
	for _, typeStr := range function.Params {
		typeArg, err := ParseTypeTag(typeStr)
		if err != nil {
			return nil, err
		}

		// If it's `signer` or `&signer` need to skip
		// TODO: only skip at the beginning
		switch innerArg := typeArg.Value.(type) {
		case *SignerTag:
			// Skip
			continue
		case *ReferenceTag:
			switch innerArg.TypeParam.Value.(type) {
			case *SignerTag:
				// Skip
				continue
			default:
				argTypes = append(argTypes, *typeArg)
			}
		default:
			argTypes = append(argTypes, *typeArg)
		}
	}

	// Check args length matches
	if len(args) != len(argTypes) {
		return nil, fmt.Errorf("entry function %s does not have the correct number of arguments for function %s", functionName, functionName)
	}

	convertedArgs := make([][]byte, len(args))
	for i, arg := range args {
		b, err := ConvertArg(argTypes[i], arg, argTypes)
		if err != nil {
			return nil, err
		}
		convertedArgs[i] = b
	}

	entry = &EntryFunction{
		Module: ModuleId{
			Address: moduleAddress,
			Name:    moduleName,
		},
		Function: functionName,
		ArgTypes: convertedTypeArgs,
		Args:     convertedArgs,
	}

	return entry, err
}

func ConvertTypeTag(typeArg any) (*TypeTag, error) {
	switch typeArg := typeArg.(type) {
	case TypeTag:
		return &typeArg, nil
	case *TypeTag:
		if typeArg == nil {
			return nil, fmt.Errorf("invalid type tag %s, cannot be nil", typeArg)
		}
		return typeArg, nil
	case string:
		return ParseTypeTag(typeArg)
	default:
		return nil, errors.New("invalid type tag type")
	}
}

func ConvertToU8(arg any) (uint8, error) {
	switch arg := arg.(type) {
	case int:
		return util.IntToU8(arg)
	case uint:
		return util.UintToU8(arg)
	case uint8:
		return arg, nil
	case big.Int:
		return util.UintToU8(uint(arg.Uint64()))
	case *big.Int:
		if arg == nil {
			return 0, errors.New("cannot convert to uint8, input is nil")
		}
		return util.UintToU8(uint(arg.Uint64()))
	case string:
		// Convert the number
		u64, err := strconv.ParseUint(arg, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(u64), nil
	default:
		return 0, errors.New("invalid input type for uint8")
	}
}

func ConvertToU16(arg any) (uint16, error) {
	switch arg := arg.(type) {
	case int:
		return util.IntToU16(arg)
	case uint:
		return util.UintToU16(arg)
	case uint16:
		return arg, nil
	case big.Int:
		return util.UintToU16(uint(arg.Uint64()))
	case *big.Int:
		if arg == nil {
			return 0, errors.New("cannot convert to uint16, input is nil")
		}
		return util.UintToU16(uint(arg.Uint64()))
	case string:
		// Convert the number
		u64, err := strconv.ParseUint(arg, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(u64), nil
	default:
		return 0, errors.New("invalid input type for uint16")
	}
}

func ConvertToU32(arg any) (uint32, error) {
	switch arg := arg.(type) {
	case int:
		return util.IntToU32(arg)
	case uint:
		return util.UintToU32(arg)
	case uint32:
		return arg, nil
	case big.Int:
		return util.UintToU32(uint(arg.Uint64()))
	case *big.Int:
		if arg == nil {
			return 0, errors.New("cannot convert to uint32, input is nil")
		}
		return util.UintToU32(uint(arg.Uint64()))
	case string:
		// Convert the number
		u64, err := strconv.ParseUint(arg, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(u64), nil
	default:
		return 0, errors.New("invalid input type for uint32")
	}
}

func ConvertToU64(arg any) (uint64, error) {
	switch arg := arg.(type) {
	case int:
		return util.IntToU64(arg)
	case uint:
		return uint64(arg), nil
	case uint64:
		return arg, nil
	case big.Int:
		return arg.Uint64(), nil
	case *big.Int:
		if arg == nil {
			return 0, errors.New("cannot convert to uint64, input is nil")
		}
		return arg.Uint64(), nil
	case string:
		return strconv.ParseUint(arg, 10, 64)
	default:
		return 0, errors.New("invalid input type for uint64")
	}
}

// TODO: Check bounds of bigints
func ConvertToU128(arg any) (num *big.Int, err error) {
	switch arg := arg.(type) {
	case int:
		return util.IntToUBigInt(arg)
	case uint:
		return util.UintToUBigInt(arg)
	case big.Int:
		return &arg, nil
	case *big.Int:
		if arg == nil {
			return nil, errors.New("cannot convert to uint128, input is nil")
		}
		return arg, nil
	case string:
		// Convert the number
		return util.StrToBigInt(arg)
	default:
		return nil, errors.New("invalid input type for uint128")
	}
}

// TODO: Check bounds of bigints
func ConvertToU256(arg any) (num *big.Int, err error) {
	switch arg := arg.(type) {
	case int:
		return util.IntToUBigInt(arg)
	case uint:
		return util.UintToUBigInt(arg)
	case big.Int:
		return &arg, nil
	case *big.Int:
		if arg == nil {
			return nil, errors.New("cannot convert to uint256, input is nil")
		}
		return arg, nil
	case string:
		// Convert the number
		return util.StrToBigInt(arg)
	default:
		return nil, errors.New("invalid input type for uint256")
	}
}

func ConvertToBool(arg any) (b bool, err error) {
	switch arg := arg.(type) {
	case bool:
		b = arg
	case string:
		switch arg {
		case "true":
			b = true
		case "false":
			b = false
		default:
			err = errors.New("invalid boolean input for bool")
		}
	default:
		err = errors.New("invalid input type for bool")
	}
	return b, err
}

func ConvertToAddress(arg any) (a *AccountAddress, err error) {
	switch arg := arg.(type) {
	case AccountAddress:
		addr := arg
		return &addr, nil
	case *AccountAddress:
		a = arg
		if a == nil {
			err = errors.New("invalid account address, nil")
		}
	case string:
		addr := AccountAddress{}
		err = addr.ParseStringRelaxed(arg)
		if err != nil {
			return nil, err
		}
		a = &addr
	default:
		err = errors.New("invalid input type for address")
	}
	return a, err
}

// ConvertToVectorU8 returns the BCS encoded version of the bytes
func ConvertToVectorU8(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg := arg.(type) {
	case string:
		bytes, err := util.ParseHex(arg)
		if err != nil {
			return nil, err
		}
		// Serialize the bytes
		return bcs.SerializeBytes(bytes)
	case []byte:
		if arg == nil {
			return nil, errors.New("cannot convert nil to bytes")
		}
		return bcs.SerializeBytes(arg)
	default:
		return nil, errors.New("invalid input type for vector<u8>")
	}
}

// ConvertToVectorU16 returns the BCS encoded version of the bytes
func ConvertToVectorU16(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg := arg.(type) {
	case []uint16:
		if arg == nil {
			return nil, errors.New("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(arg, ser, func(ser *bcs.Serializer, item uint16) {
				ser.U16(item)
			})
		})
	default:
		return nil, errors.New("invalid input type for vector<u16>")
	}
}

// ConvertToVectorU32 returns the BCS encoded version of the bytes
func ConvertToVectorU32(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg := arg.(type) {
	case []uint32:
		if arg == nil {
			return nil, errors.New("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(arg, ser, func(ser *bcs.Serializer, item uint32) {
				ser.U32(item)
			})
		})
	default:
		return nil, errors.New("invalid input type for vector<u32>")
	}
}

// ConvertToVectorU64 returns the BCS encoded version of the bytes
func ConvertToVectorU64(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg := arg.(type) {
	case []uint64:
		if arg == nil {
			return nil, errors.New("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(arg, ser, func(ser *bcs.Serializer, item uint64) {
				ser.U64(item)
			})
		})
	default:
		return nil, errors.New("invalid input type for vector<u64>")
	}
}

// ConvertToVectorU128 returns the BCS encoded version of the bytes
func ConvertToVectorU128(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg := arg.(type) {
	case []big.Int:
		if arg == nil {
			return nil, errors.New("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(arg, ser, func(ser *bcs.Serializer, item big.Int) {
				ser.U128(item)
			})
		})
	default:
		return nil, errors.New("invalid input type for vector<u128>")
	}
}

// ConvertToVectorU256 returns the BCS encoded version of the bytes
func ConvertToVectorU256(arg any) (b []byte, err error) {
	// Special case, handle hex string
	switch arg := arg.(type) {
	case []big.Int:
		if arg == nil {
			return nil, errors.New("cannot convert nil to bytes")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(arg, ser, func(ser *bcs.Serializer, item big.Int) {
				ser.U256(item)
			})
		})
	default:
		return nil, errors.New("invalid input type for vector<u256>")
	}
}

// ConvertToVectorBool returns the BCS encoded version of the boolean vector
func ConvertToVectorBool(arg any) (b []byte, err error) {
	switch arg := arg.(type) {
	case []bool:
		if arg == nil {
			return nil, errors.New("cannot convert nil to bool vector")
		}
		return bcs.SerializeSingle(func(ser *bcs.Serializer) {
			bcs.SerializeSequenceWithFunction(arg, ser, func(ser *bcs.Serializer, item bool) {
				ser.Bool(item)
			})
		})
	default:
		return nil, errors.New("invalid input type for vector<bool>")
	}
}

// ConvertToVectorAddress returns the BCS encoded version of the address vector
func ConvertToVectorAddress(arg any) (b []byte, err error) {
	switch arg := arg.(type) {
	case []AccountAddress:
		if arg == nil {
			return nil, errors.New("cannot convert nil to address vector")
		}
		return bcs.SerializeSequenceOnly(arg)
	case []*AccountAddress:
		if arg == nil {
			return nil, errors.New("cannot convert nil to address vector")
		}
		// Convert []*AccountAddress to []AccountAddress
		addresses := make([]AccountAddress, len(arg))
		for i, addr := range arg {
			if addr == nil {
				return nil, errors.New("cannot convert nil address in vector")
			}
			addresses[i] = *addr
		}
		return bcs.SerializeSequenceOnly(addresses)
	default:
		return nil, errors.New("invalid input type for vector<address>")
	}
}

// ConvertToVectorGeneric returns the BCS encoded version of the generic vector
func ConvertToVectorGeneric(typeArg TypeTag, arg any, generics []TypeTag) (b []byte, err error) {
	genericTag, ok := typeArg.Value.(*GenericTag)
	if !ok {
		return nil, errors.New("invalid type tag for generic vector")
	}

	if genericTag.Num >= uint64(len(generics)) {
		return nil, errors.New("generic number out of bounds")
	}

	innerType := generics[genericTag.Num]

	switch arg := arg.(type) {
	case []any:
		if arg == nil {
			return nil, errors.New("cannot convert nil to generic vector")
		}

		// Serialize length
		length, err := util.IntToU32(len(arg))
		if err != nil {
			return nil, err
		}
		lengthBytes, err := bcs.SerializeUleb128(length)
		if err != nil {
			return nil, err
		}
		b = lengthBytes

		// Serialize each element
		for _, item := range arg {
			val, err := ConvertArg(innerType, item, generics)
			if err != nil {
				return nil, err
			}
			b = append(b, val...)
		}
		return b, nil
	default:
		return nil, errors.New("invalid input type for generic vector")
	}
}

// ConvertToVectorReference returns the BCS encoded version of the reference vector
func ConvertToVectorReference(typeArg TypeTag, arg any, generics []TypeTag) (b []byte, err error) {
	refTag, ok := typeArg.Value.(*ReferenceTag)
	if !ok {
		return nil, errors.New("invalid type tag for reference vector")
	}

	innerType := refTag.TypeParam

	switch arg := arg.(type) {
	case []any:
		if arg == nil {
			return nil, errors.New("cannot convert nil to reference vector")
		}

		// Serialize length
		length, err := util.IntToU32(len(arg))
		if err != nil {
			return nil, err
		}
		lengthBytes, err := bcs.SerializeUleb128(length)
		if err != nil {
			return nil, err
		}
		b = lengthBytes

		// Serialize each element
		for _, item := range arg {
			val, err := ConvertArg(innerType, item, generics)
			if err != nil {
				return nil, err
			}
			b = append(b, val...)
		}
		return b, nil
	default:
		return nil, errors.New("invalid input type for reference vector")
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
	switch innerType := typeArg.Value.(type) {
	case *U8Tag:
		num, err := ConvertToU8(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU8(num)
	case *U16Tag:
		num, err := ConvertToU16(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU16(num)
	case *U32Tag:
		num, err := ConvertToU32(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU32(num)
	case *U64Tag:
		num, err := ConvertToU64(arg)
		if err != nil {
			return nil, err
		}
		return bcs.SerializeU64(num)
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
		genericNum := innerType.Num
		if genericNum >= uint64(len(generics)) {
			return nil, errors.New("generic number out of bounds")
		}

		tag := generics[genericNum]
		return ConvertArg(tag, arg, generics)
	case *ReferenceTag:
		// Convert based on inner type
		return ConvertArg(innerType.TypeParam, arg, generics)
	case *VectorTag:
		// This has two paths:
		// 1. Hex strings are allowed for vector<u8>
		// 2. Otherwise, everything is just parsed as an array of the inner type
		vecTag := innerType
		switch vecTag.TypeParam.Value.(type) {
		case *U8Tag:
			return ConvertToVectorU8(arg)
		default:
			return ConvertToVector(vecTag.TypeParam, arg, generics)
		}
	case *StructTag:
		structTag := innerType
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
					switch arg := arg.(type) {
					case string:
						return bcs.SerializeBytes([]byte(arg))
					default:
						return nil, errors.New("invalid input type for 0x1::string::String")
					}
				}
			case "option":
				switch structTag.Name {
				case "Option":
					// Check it has the proper inner type
					if 1 != len(structTag.TypeParams) {
						return nil, errors.New("invalid input type for option, must have exactly one type arg")
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
		return nil, errors.New("unknown type argument")
	}

	return nil, errors.New("failed to convert type argument")
}
