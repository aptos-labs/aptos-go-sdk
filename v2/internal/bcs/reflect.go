package bcs

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// marshalReflect uses reflection to serialize a value.
func marshalReflect(ser *Serializer, v any) error {
	return marshalValue(ser, reflect.ValueOf(v))
}

func marshalValue(ser *Serializer, v reflect.Value) error {
	// Handle pointers by dereferencing
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ErrNilValue
		}
		v = v.Elem()
	}

	// Check if it implements Marshaler
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(Marshaler); ok {
			m.MarshalBCS(ser)
			return ser.err
		}
	}
	if m, ok := v.Interface().(Marshaler); ok {
		m.MarshalBCS(ser)
		return ser.err
	}

	switch v.Kind() {
	case reflect.Bool:
		ser.Bool(v.Bool())
	case reflect.Uint8:
		ser.U8(uint8(v.Uint()))
	case reflect.Uint16:
		ser.U16(uint16(v.Uint()))
	case reflect.Uint32:
		ser.U32(uint32(v.Uint()))
	case reflect.Uint64, reflect.Uint:
		ser.U64(v.Uint())
	case reflect.Int8:
		ser.I8(int8(v.Int()))
	case reflect.Int16:
		ser.I16(int16(v.Int()))
	case reflect.Int32:
		ser.I32(int32(v.Int()))
	case reflect.Int64, reflect.Int:
		ser.I64(v.Int())
	case reflect.String:
		ser.WriteString(v.String())
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// []byte - serialize as bytes
			ser.WriteBytes(v.Bytes())
		} else {
			// Other slices - serialize as sequence
			if err := marshalSlice(ser, v); err != nil {
				return err
			}
		}
	case reflect.Array:
		// Fixed-size arrays are serialized without length prefix
		for i := 0; i < v.Len(); i++ {
			if err := marshalValue(ser, v.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Struct:
		return marshalStruct(ser, v)
	default:
		return fmt.Errorf("%w: unsupported kind %s", ErrInvalidType, v.Kind())
	}

	return ser.err
}

func marshalSlice(ser *Serializer, v reflect.Value) error {
	length := v.Len()
	if length > 0xFFFFFFFF {
		return ErrOverflow
	}
	ser.Uleb128(uint32(length))
	for i := 0; i < length; i++ {
		if err := marshalValue(ser, v.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

type fieldInfo struct {
	index    int
	order    int
	optional bool
}

func marshalStruct(ser *Serializer, v reflect.Value) error {
	t := v.Type()

	// Handle special types
	switch v.Interface().(type) {
	case big.Int:
		// big.Int should be handled by the caller as U128/U256
		return fmt.Errorf("%w: big.Int must be serialized with specific bit width", ErrInvalidType)
	}

	fields := getStructFields(t)

	for _, fi := range fields {
		field := v.Field(fi.index)

		if fi.optional {
			// Handle optional field
			if isZero(field) {
				ser.Uleb128(0) // None
			} else {
				ser.Uleb128(1) // Some
				if err := marshalValue(ser, field); err != nil {
					return err
				}
			}
		} else {
			if err := marshalValue(ser, field); err != nil {
				return err
			}
		}
	}

	return ser.err
}

func getStructFields(t reflect.Type) []fieldInfo {
	var fields []fieldInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("bcs")
		if tag == "-" {
			continue
		}

		fi := fieldInfo{index: i, order: i}

		// Parse tag
		parts := strings.Split(tag, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "optional" {
				fi.optional = true
			} else if order, err := strconv.Atoi(part); err == nil {
				fi.order = order
			}
		}

		fields = append(fields, fi)
	}

	// Sort by order
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].order < fields[j].order
	})

	return fields
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	default:
		return v.IsZero()
	}
}

// unmarshalReflect uses reflection to deserialize a value.
func unmarshalReflect(des *Deserializer, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("%w: must be a non-nil pointer", ErrInvalidType)
	}
	return unmarshalValue(des, rv.Elem())
}

func unmarshalValue(des *Deserializer, v reflect.Value) error {
	if des.err != nil {
		return des.err
	}

	// Check if it implements Unmarshaler
	if v.CanAddr() {
		if u, ok := v.Addr().Interface().(Unmarshaler); ok {
			u.UnmarshalBCS(des)
			return des.err
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(des.Bool())
	case reflect.Uint8:
		v.SetUint(uint64(des.U8()))
	case reflect.Uint16:
		v.SetUint(uint64(des.U16()))
	case reflect.Uint32:
		v.SetUint(uint64(des.U32()))
	case reflect.Uint64, reflect.Uint:
		v.SetUint(des.U64())
	case reflect.Int8:
		v.SetInt(int64(des.I8()))
	case reflect.Int16:
		v.SetInt(int64(des.I16()))
	case reflect.Int32:
		v.SetInt(int64(des.I32()))
	case reflect.Int64, reflect.Int:
		v.SetInt(des.I64())
	case reflect.String:
		v.SetString(des.ReadString())
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			v.SetBytes(des.ReadBytes())
		} else {
			if err := unmarshalSlice(des, v); err != nil {
				return err
			}
		}
	case reflect.Array:
		// Fixed-size arrays
		for i := 0; i < v.Len(); i++ {
			if err := unmarshalValue(des, v.Index(i)); err != nil {
				return err
			}
		}
	case reflect.Struct:
		return unmarshalStruct(des, v)
	case reflect.Ptr:
		// Allocate and deserialize
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return unmarshalValue(des, v.Elem())
	default:
		return fmt.Errorf("%w: unsupported kind %s", ErrInvalidType, v.Kind())
	}

	return des.err
}

func unmarshalSlice(des *Deserializer, v reflect.Value) error {
	length := des.Uleb128()
	if des.err != nil {
		return des.err
	}

	slice := reflect.MakeSlice(v.Type(), int(length), int(length))
	for i := 0; i < int(length); i++ {
		if err := unmarshalValue(des, slice.Index(i)); err != nil {
			return err
		}
	}
	v.Set(slice)
	return nil
}

func unmarshalStruct(des *Deserializer, v reflect.Value) error {
	t := v.Type()
	fields := getStructFields(t)

	for _, fi := range fields {
		field := v.Field(fi.index)

		if fi.optional {
			// Handle optional field
			length := des.Uleb128()
			if des.err != nil {
				return des.err
			}
			switch length {
			case 0:
				// None - leave as zero value
			case 1:
				// Some - deserialize value
				if err := unmarshalValue(des, field); err != nil {
					return err
				}
			default:
				return fmt.Errorf("%w: got %d elements", ErrInvalidOptionLen, length)
			}
		} else {
			if err := unmarshalValue(des, field); err != nil {
				return err
			}
		}
	}

	return des.err
}
