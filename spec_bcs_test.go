package aptos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"testing"

	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/internal/util"
	"github.com/cucumber/godog"
)

type TestStruct struct {
	bool    bool
	u8      uint8
	u16     uint16
	u32     uint32
	u64     uint64
	u128    big.Int
	u256    big.Int
	address *AccountAddress
	bytes   []byte
}

func (st *TestStruct) MarshalBCS(ser *bcs.Serializer) {
	ser.Bool(st.bool)
	ser.U8(st.u8)
	ser.U16(st.u16)
	ser.U32(st.u32)
	ser.U64(st.u64)
	ser.U128(st.u128)
	ser.U256(st.u256)
	ser.Struct(st.address)
	ser.WriteBytes(st.bytes)
}

func (st *TestStruct) UnmarshalBCS(des *bcs.Deserializer) {
	st.bool = des.Bool()
	st.u8 = des.U8()
	st.u16 = des.U16()
	st.u32 = des.U32()
	st.u64 = des.U64()
	st.u128 = des.U128()
	st.u256 = des.U256()
	st.address = &AccountAddress{}
	des.Struct(st.address)
	st.bytes = des.ReadBytes()
	// Custom error
	if len(st.bytes) == 0 {
		des.SetError(errors.New("invalid bytes length, must be at least 1"))
	}
}

// godogsCtxKey is the key used to store the available godogs in the context.Context.
type godogsCtxKey struct{}

func givenAddress(ctx context.Context, input string) (context.Context, error) {
	address := &AccountAddress{}
	err := address.ParseStringRelaxed(input)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, godogsCtxKey{}, address), nil
}

func givenBoolean(ctx context.Context, input string) (context.Context, error) {
	return context.WithValue(ctx, godogsCtxKey{}, parseBoolean(input)), nil
}

func givenU8(ctx context.Context, input int) (context.Context, error) {
	if input < 0 || input > 255 {
		return nil, errors.New("u8 must be between 0 and 255")
	}
	return context.WithValue(ctx, godogsCtxKey{}, (uint8)(input)), nil
}

func givenU16(ctx context.Context, input int) (context.Context, error) {
	if input < 0 || input > 65535 {
		return nil, errors.New("u16 must be between 0 and 65535")
	}
	return context.WithValue(ctx, godogsCtxKey{}, (uint16)(input)), nil
}

func givenU32(ctx context.Context, input int) (context.Context, error) {
	if input < 0 || input > 4294967295 {
		return nil, errors.New("u32 must be between 0 and 4294967295")
	}
	return context.WithValue(ctx, godogsCtxKey{}, (uint32)(input)), nil
}

func givenU64(ctx context.Context, input string) (context.Context, error) {
	val, err := StrToUint64(input)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, godogsCtxKey{}, val), nil
}

func givenU128(ctx context.Context, input string) (context.Context, error) {
	val, err := StrToBigInt(input)
	// TODO: Check that the input is a valid u128
	if err != nil {
		return nil, fmt.Errorf("u128 must be a valid number %w", err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, val), nil
}

func givenU256(ctx context.Context, input string) (context.Context, error) {
	val, err := StrToBigInt(input)
	// TODO: Check that the input is a valid u256
	if err != nil {
		return nil, fmt.Errorf("u256 must be a valid number %w", err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, val), nil
}

func givenBytes(ctx context.Context, input string) (context.Context, error) {
	return context.WithValue(ctx, godogsCtxKey{}, parseHex(input)), nil
}

func givenString(ctx context.Context, input string) (context.Context, error) {
	return context.WithValue(ctx, godogsCtxKey{}, input), nil
}

func givenStruct(ctx context.Context, items string) (context.Context, error) {
	itemList := strings.Split(items, ",")
	str := &TestStruct{
		bool:    parseBoolean(itemList[0]),
		u8:      parseU8(itemList[1]),
		u16:     parseU16(itemList[2]),
		u32:     parseU32(itemList[3]),
		u64:     parseU64(itemList[4]),
		u128:    *parseU128(itemList[5]),
		u256:    *parseU256(itemList[6]),
		address: parseAddress(itemList[7]),
		bytes:   parseHex(itemList[8]),
	}

	return context.WithValue(ctx, godogsCtxKey{}, str), nil
}

func givenSequence(ctx context.Context, itemType string, items string) (context.Context, error) {
	return context.WithValue(ctx, godogsCtxKey{}, parseSequence(itemType, items)), nil
}

func serializeAddress(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(*AccountAddress)
	if !ok {
		return ctx, errors.New("input is not *AccountAddress")
	}

	out, err := bcs.Serialize(input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize address %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeBool(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(bool)
	if !ok {
		return ctx, errors.New("input is not bool")
	}

	out, err := bcs.SerializeBool(input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize boolean %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeU8(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(uint8)
	if !ok {
		return ctx, errors.New("input is not uint8")
	}

	out, err := bcs.SerializeU8(input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize u8 %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeU16(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(uint16)
	if !ok {
		return ctx, errors.New("input is not uint16")
	}

	out, err := bcs.SerializeU16(input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize u16 %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeU32(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(uint32)
	if !ok {
		return ctx, errors.New("input is not uint32")
	}

	out, err := bcs.SerializeU32(input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize u32 %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeU64(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(uint64)
	if !ok {
		return ctx, errors.New("input is not uint64")
	}

	out, err := bcs.SerializeU64(input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize u64 %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeU128(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(*big.Int)
	if !ok {
		return ctx, errors.New("input is not *big.Int")
	}

	out, err := bcs.SerializeU128(*input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize u128 %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeU256(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(*big.Int)
	if !ok {
		return ctx, errors.New("input is not *big.Int")
	}

	out, err := bcs.SerializeU256(*input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize u256 %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeUleb128(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(uint32)
	if !ok {
		return ctx, errors.New("input is not uint32")
	}

	out, err := bcs.SerializeSingle(func(ser *bcs.Serializer) {
		ser.Uleb128(input)
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize uleb128 %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeFixedBytes(ctx context.Context, _ int) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}
	out, err := bcs.SerializeSingle(func(ser *bcs.Serializer) {
		ser.FixedBytes(input)
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize fixed bytes %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeBytes(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}
	out, err := bcs.SerializeBytes(input)
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize fixed bytes %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeString(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).(string)
	if !ok {
		return ctx, errors.New("input is not string")
	}
	out, err := bcs.SerializeSingle(func(ser *bcs.Serializer) {
		ser.WriteString(input)
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize string %v: %w", input, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, out), nil
}

func serializeSequence(ctx context.Context, itemType string) (context.Context, error) {
	ser := &bcs.Serializer{}

	switch itemType {
	case "address":
		input, ok := ctx.Value(godogsCtxKey{}).([]AccountAddress)
		if !ok {
			return ctx, errors.New("input is not []AccountAddress")
		}

		bcs.SerializeSequence(input, ser)
	case "bool":
		input, ok := ctx.Value(godogsCtxKey{}).([]bool)
		if !ok {
			return ctx, errors.New("input is not []bool")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item bool) {
			ser.Bool(item)
		})
	case "u8":
		input, ok := ctx.Value(godogsCtxKey{}).([]uint8)
		if !ok {
			return ctx, errors.New("input is not []uint8")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item uint8) {
			ser.U8(item)
		})
	case "u16":
		input, ok := ctx.Value(godogsCtxKey{}).([]uint16)
		if !ok {
			return ctx, errors.New("input is not []uint16")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item uint16) {
			ser.U16(item)
		})
	case "u32":
		input, ok := ctx.Value(godogsCtxKey{}).([]uint32)
		if !ok {
			return ctx, errors.New("input is not []uint32")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item uint32) {
			ser.U32(item)
		})
	case "u64":
		input, ok := ctx.Value(godogsCtxKey{}).([]uint64)
		if !ok {
			return ctx, errors.New("input is not []uint64")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item uint64) {
			ser.U64(item)
		})
	case "u128":
		input, ok := ctx.Value(godogsCtxKey{}).([]*big.Int)
		if !ok {
			return ctx, errors.New("input is not []*big.Int")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item *big.Int) {
			ser.U128(*item)
		})
	case "u256":
		input, ok := ctx.Value(godogsCtxKey{}).([]*big.Int)
		if !ok {
			return ctx, errors.New("input is not []*big.Int")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item *big.Int) {
			ser.U256(*item)
		})
	case "uleb128":
		input, ok := ctx.Value(godogsCtxKey{}).([]uint32)
		if !ok {
			return ctx, errors.New("input is not []uint32")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item uint32) {
			ser.Uleb128(item)
		})
	case "string":
		input, ok := ctx.Value(godogsCtxKey{}).([]string)
		if !ok {
			return ctx, errors.New("input is not []string")
		}

		bcs.SerializeSequenceWithFunction(input, ser, func(ser *bcs.Serializer, item string) {
			ser.WriteString(item)
		})
	default:
		return ctx, fmt.Errorf("unsupported serialize item type %s", itemType)
	}

	result := ser.ToBytes()
	err := ser.Error()
	if err != nil {
		return ctx, fmt.Errorf("failed to serialize %s sequence: %w", itemType, err)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeAddress(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	result := &AccountAddress{}
	err := bcs.Deserialize(result, input)
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeBool(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.Bool()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return context.WithValue(ctx, godogsCtxKey{}, fmt.Errorf("expected no remaining bytes, but got %d", remaining)), nil
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeU8(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.U8()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeU16(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.U16()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeU32(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.U32()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeU64(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.U64()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeU128(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.U128()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, &result), nil
}

func deserializeU256(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.U256()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, &result), nil
}

func deserializeUleb128(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.Uleb128()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeFixedBytes(ctx context.Context, length int) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.ReadFixedBytes(length)
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeBytes(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.ReadBytes()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeString(ctx context.Context) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	result := des.ReadString()
	err := des.Error()
	if err != nil {
		return context.WithValue(ctx, godogsCtxKey{}, err), nil
	}

	remaining := des.Remaining()
	if remaining != 0 {
		return ctx, fmt.Errorf("expected no remaining bytes, but got %d", remaining)
	}

	return context.WithValue(ctx, godogsCtxKey{}, result), nil
}

func deserializeSequence(ctx context.Context, itemType string) (context.Context, error) {
	input, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return ctx, errors.New("input is not []byte")
	}

	des := bcs.NewDeserializer(input)
	switch itemType {
	case "address":
		result := bcs.DeserializeSequence[AccountAddress](des)
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "bool":
		result := bcs.DeserializeSequenceWithFunction[bool](des, func(des *bcs.Deserializer, out *bool) {
			boolean := des.Bool()
			*out = boolean
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "u8":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out *uint8) {
			*out = des.U8()
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "u16":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out *uint16) {
			*out = des.U16()
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "u32":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out *uint32) {
			*out = des.U32()
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "u64":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out *uint64) {
			*out = des.U64()
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "u128":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out **big.Int) {
			num := des.U128()
			*out = &num
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "u256":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out **big.Int) {
			num := des.U256()
			*out = &num
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "uleb128":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out *uint32) {
			*out = des.Uleb128()
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "bytes":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out *[]byte) {
			*out = des.ReadBytes()
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	case "string":
		result := bcs.DeserializeSequenceWithFunction(des, func(des *bcs.Deserializer, out *string) {
			*out = des.ReadString()
		})
		err := des.Error()
		if err != nil {
			return context.WithValue(ctx, godogsCtxKey{}, err), nil
		}
		return context.WithValue(ctx, godogsCtxKey{}, result), nil
	default:
		return ctx, fmt.Errorf("unsupported deserialize sequence item type %s", itemType)
	}
}

func addressResult(ctx context.Context, expected string) error {
	expectedAddress := &AccountAddress{}
	err := expectedAddress.ParseStringRelaxed(expected)
	if err != nil {
		return err
	}

	result, ok := ctx.Value(godogsCtxKey{}).(*AccountAddress)
	if !ok {
		return errors.New("no result available")
	}

	// You can't compare pointers of addresses
	if *expectedAddress != *result {
		return fmt.Errorf("expected %s, but received %s", expectedAddress.String(), result.String())
	}

	return nil
}

func boolResult(ctx context.Context, expected string) error {
	expectedBool := parseBoolean(expected)

	result, ok := ctx.Value(godogsCtxKey{}).(bool)
	if !ok {
		return errors.New("no result available")
	}

	if expectedBool != result {
		return fmt.Errorf("expected %v, but received %v", expectedBool, result)
	}

	return nil
}

func u8Result(ctx context.Context, expected int) error {
	expectedU8, err := util.IntToU8(expected)
	if err != nil {
		return errors.New("invalid value for U8")
	}

	result, ok := ctx.Value(godogsCtxKey{}).(uint8)
	if !ok {
		return errors.New("no result available")
	}

	if expectedU8 != result {
		return fmt.Errorf("expected %d, but received %d", expectedU8, result)
	}

	return nil
}

func u16Result(ctx context.Context, expected int) error {
	expectedU16, err := util.IntToU16(expected)
	if err != nil {
		return errors.New("invalid value for U16")
	}

	result, ok := ctx.Value(godogsCtxKey{}).(uint16)
	if !ok {
		return errors.New("no result available")
	}

	if expectedU16 != result {
		return fmt.Errorf("expected %d, but received %d", expectedU16, result)
	}
	return nil
}

func u32Result(ctx context.Context, expected int) error {
	expectedU32, err := util.IntToU32(expected)
	if err != nil {
		return errors.New("invalid value for U32")
	}

	result, ok := ctx.Value(godogsCtxKey{}).(uint32)
	if !ok {
		return errors.New("no result available")
	}

	if expectedU32 != result {
		return fmt.Errorf("expected %d, but received %d", expectedU32, result)
	}
	return nil
}

func u64Result(ctx context.Context, expected string) error {
	expectedU64, err := StrToUint64(expected)
	if err != nil {
		return err
	}

	result, ok := ctx.Value(godogsCtxKey{}).(uint64)
	if !ok {
		return errors.New("no result available")
	}

	if expectedU64 != result {
		return fmt.Errorf("expected %d, but received %d", expectedU64, result)
	}
	return nil
}

func u128Result(ctx context.Context, expected string) error {
	expectedU128, err := StrToBigInt(expected)
	if err != nil {
		return err
	}

	result, ok := ctx.Value(godogsCtxKey{}).(*big.Int)
	if !ok {
		return errors.New("no result available")
	}

	if expectedU128.Cmp(result) != 0 {
		return fmt.Errorf("expected %d, but received %d", expectedU128, result)
	}
	return nil
}

func u256Result(ctx context.Context, expected string) error {
	expectedU256, err := StrToBigInt(expected)
	if err != nil {
		return err
	}

	result, ok := ctx.Value(godogsCtxKey{}).(*big.Int)
	if !ok {
		return errors.New("no result available")
	}

	if expectedU256.Cmp(result) != 0 {
		return fmt.Errorf("expected %d, but received %d", expectedU256, result)
	}
	return nil
}

func bytesResult(ctx context.Context, expected string) error {
	expectedBytes := parseHex(expected)
	result, ok := ctx.Value(godogsCtxKey{}).([]byte)
	if !ok {
		return errors.New("no result available")
	}

	if !bytes.Equal(expectedBytes, result) {
		return fmt.Errorf("expected 0x%X, but received 0x%X", expectedBytes, result)
	}

	return nil
}

func stringResult(ctx context.Context, expected string) error {
	result, ok := ctx.Value(godogsCtxKey{}).(string)
	if !ok {
		return errors.New("no result available")
	}

	if expected != result {
		return fmt.Errorf("expected %s, but received %s", expected, result)
	}

	return nil
}

func sequenceResult(ctx context.Context, itemType string, expectedList string) error {
	expected := parseSequence(itemType, expectedList)
	switch itemType {
	case "address":
		result, ok := ctx.Value(godogsCtxKey{}).([]AccountAddress)
		if !ok {
			return errors.New("no address result available")
		}

		addrs, ok := expected.([]AccountAddress)
		if !ok {
			return errors.New("no expected addresses available")
		}
		if !slices.Equal(addrs, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	case "bool":
		result, ok := ctx.Value(godogsCtxKey{}).([]bool)
		if !ok {
			return errors.New("no bool result available")
		}
		b, ok := expected.([]bool)
		if !ok {
			return errors.New("no bool result available")
		}
		if !slices.Equal(b, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	case "u8":
		result, ok := ctx.Value(godogsCtxKey{}).([]uint8)
		if !ok {
			return errors.New("no u8 result available")
		}
		ex, ok := expected.([]uint8)
		if !ok {
			return errors.New("no expected u8 result available")
		}
		if !slices.Equal(ex, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	case "u16":
		result, ok := ctx.Value(godogsCtxKey{}).([]uint16)
		if !ok {
			return errors.New("no u16 result available")
		}
		ex, ok := expected.([]uint16)
		if !ok {
			return errors.New("no expected u16 result available")
		}
		if !slices.Equal(ex, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	case "u32":
		result, ok := ctx.Value(godogsCtxKey{}).([]uint32)
		if !ok {
			return errors.New("no u32 result available")
		}
		ex, ok := expected.([]uint32)
		if !ok {
			return errors.New("no expected u32 result available")
		}
		if !slices.Equal(ex, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	case "u64":
		result, ok := ctx.Value(godogsCtxKey{}).([]uint64)
		if !ok {
			return errors.New("no u64 result available")
		}
		ex, ok := expected.([]uint64)
		if !ok {
			return errors.New("no expected u64 result available")
		}
		if !slices.Equal(ex, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	case "u128":
		result, ok := ctx.Value(godogsCtxKey{}).([]*big.Int)
		if !ok {
			return errors.New("no u128 result available")
		}
		expectedInts, ok := expected.([]*big.Int)
		if !ok {
			return errors.New("no expected u128 result available")
		}
		if len(expectedInts) != len(result) {
			return fmt.Errorf("expected %v, but received %v", expectedInts, result)
		}
		for i, expectedInt := range expectedInts {
			if expectedInt.Cmp(result[i]) != 0 {
				return fmt.Errorf("expected %v, but received %v", expectedInt, result[i])
			}
		}
	case "u256":
		result, ok := ctx.Value(godogsCtxKey{}).([]*big.Int)
		if !ok {
			return errors.New("no u256 result available")
		}
		expectedInts, ok := expected.([]*big.Int)
		if !ok {
			return errors.New("no expected u256 result available")
		}
		if len(expectedInts) != len(result) {
			return fmt.Errorf("expected %v, but received %v", expectedInts, result)
		}
		for i, expectedInt := range expectedInts {
			if expectedInt.Cmp(result[i]) != 0 {
				return fmt.Errorf("expected %v, but received %v", expectedInt, result[i])
			}
		}
	case "uleb128":
		result, ok := ctx.Value(godogsCtxKey{}).([]uint32)
		if !ok {
			return errors.New("no uleb128 result available")
		}
		ex, ok := expected.([]uint32)
		if !ok {
			return errors.New("no expected uleb128 result available")
		}
		if !slices.Equal(ex, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	case "string":
		result, ok := ctx.Value(godogsCtxKey{}).([]string)
		if !ok {
			return errors.New("no string result available")
		}
		ex, ok := expected.([]string)
		if !ok {
			return errors.New("no string result available")
		}
		if !slices.Equal(ex, result) {
			return fmt.Errorf("expected %v, but received %v", expected, result)
		}
	default:
		return fmt.Errorf("unsupported sequence item type %s", itemType)
	}

	return nil
}

func failResult(ctx context.Context) error {
	_, ok := ctx.Value(godogsCtxKey{}).(error)
	if !ok {
		return errors.New("no error available")
	}

	// TODO: add optional error message check
	return nil
}

func TestFeatures(t *testing.T) {
	t.Parallel()
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Given(`^address (0x[0-9a-fA-F]+)$`, givenAddress)
	sc.Given(`^bool (true|false)$`, givenBoolean)
	sc.Given(`^u8 (\d+)$`, givenU8)
	sc.Given(`^u16 (\d+)$`, givenU16)
	sc.Given(`^u32 (\d+)$`, givenU32)
	sc.Given(`^u64 (\d+)$`, givenU64)
	sc.Given(`^u128 (\d+)$`, givenU128)
	sc.Given(`^u256 (\d+)$`, givenU256)
	sc.Given(`^bytes (0x[0-9a-fA-F]*)$`, givenBytes)
	sc.Given(`^string "(.*)"$`, givenString)
	sc.Given(`^sequence of ([0-9a-zA-Z]+) \[(.*)\]$`, givenSequence)
	sc.Given(`^struct of \[(.+)\]$`, givenStruct)

	sc.When(`^I serialize as address$`, serializeAddress)
	sc.When(`^I serialize as bool$`, serializeBool)
	sc.When(`^I serialize as u8$`, serializeU8)
	sc.When(`^I serialize as u16$`, serializeU16)
	sc.When(`^I serialize as u32$`, serializeU32)
	sc.When(`^I serialize as u64$`, serializeU64)
	sc.When(`^I serialize as u128$`, serializeU128)
	sc.When(`^I serialize as u256$`, serializeU256)
	sc.When(`^I serialize as uleb128$`, serializeUleb128)
	sc.When(`^I serialize as fixed bytes with length (\d+)$`, serializeFixedBytes)
	sc.When(`^I serialize as bytes$`, serializeBytes)
	sc.When(`^I serialize as string$`, serializeString)
	sc.When(`^I serialize as sequence of ([0-9a-zA-Z]+)$`, serializeSequence)

	sc.When(`^I deserialize as address$`, deserializeAddress)
	sc.When(`^I deserialize as bool$`, deserializeBool)
	sc.When(`^I deserialize as u8$`, deserializeU8)
	sc.When(`^I deserialize as u16$`, deserializeU16)
	sc.When(`^I deserialize as u32$`, deserializeU32)
	sc.When(`^I deserialize as u64$`, deserializeU64)
	sc.When(`^I deserialize as u128$`, deserializeU128)
	sc.When(`^I deserialize as u256$`, deserializeU256)
	sc.When(`^I deserialize as uleb128$`, deserializeUleb128)
	sc.When(`^I deserialize as fixed bytes with length (\d+)$`, deserializeFixedBytes)
	sc.When(`^I deserialize as bytes$`, deserializeBytes)
	sc.When(`^I deserialize as string$`, deserializeString)
	sc.When(`^I deserialize as sequence of ([0-9a-zA-Z]+)$`, deserializeSequence)

	sc.Then(`^the result should be address (0x[0-9a-fA-F]+)$`, addressResult)
	sc.Then(`^the result should be bool (true|false)$`, boolResult)
	sc.Then(`^the result should be u8 (\d+)$`, u8Result)
	sc.Then(`^the result should be u16 (\d+)$`, u16Result)
	sc.Then(`^the result should be u32 (\d+)$`, u32Result)
	sc.Then(`^the result should be u64 (\d+)$`, u64Result)
	sc.Then(`^the result should be u128 (\d+)$`, u128Result)
	sc.Then(`^the result should be u256 (\d+)$`, u256Result)
	sc.Then(`^the result should be bytes (0x[0-9a-fA-F]*)`, bytesResult)
	sc.Then(`^the result should be string "(.*)"$`, stringResult)
	sc.Then(`^the result should be sequence of ([0-9a-zA-Z]+) \[(.*)\]$`, sequenceResult)
	sc.Then(`^the deserialization should fail$`, failResult)
}
