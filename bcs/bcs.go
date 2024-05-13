package bcs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
)

type Marshaler interface {
	MarshalBCS(*Serializer)
}
type Unmarshaler interface {
	UnmarshalBCS(*Deserializer)
}

type Struct interface {
	Marshaler
	Unmarshaler
}

type Serializer struct {
	out bytes.Buffer
	err error
}

func (ser *Serializer) Error() error {
	return ser.err
}

// If the data is well formed but nonsense, MarshalBCS() code can set error
func (ser *Serializer) SetError(err error) {
	ser.err = err
}

func (ser *Serializer) U8(v uint8) {
	ser.out.WriteByte(v)
}

func (ser *Serializer) U16(v uint16) {
	var ub [2]byte
	binary.LittleEndian.PutUint16(ub[:], v)
	ser.out.Write(ub[:])
}

func (ser *Serializer) U32(v uint32) {
	var ub [4]byte
	binary.LittleEndian.PutUint32(ub[:], v)
	ser.out.Write(ub[:])
}

func (ser *Serializer) U64(v uint64) {
	var ub [8]byte
	binary.LittleEndian.PutUint64(ub[:], v)
	ser.out.Write(ub[:])
}

func reverse(ub []byte) {
	lo := 0
	hi := len(ub) - 1
	for hi > lo {
		t := ub[lo]
		ub[lo] = ub[hi]
		ub[hi] = t
		lo++
		hi--
	}
}

func (ser *Serializer) U128(v big.Int) {
	var ub [16]byte
	v.FillBytes(ub[:])
	reverse(ub[:])
	ser.out.Write(ub[:])
}

func (ser *Serializer) U256(v big.Int) {
	var ub [32]byte
	v.FillBytes(ub[:])
	reverse(ub[:])
	ser.out.Write(ub[:])
}

func (ser *Serializer) Uleb128(v uint32) {
	for v > 0x80 {
		nb := uint8(v & 0x7f)
		ser.out.WriteByte(0x80 | nb)
		v = v >> 7
	}
	ser.out.WriteByte(uint8(v & 0x7f))
}

func (ser *Serializer) WriteBytes(v []byte) {
	ser.Uleb128(uint32(len(v)))
	ser.out.Write(v)
}

func (ser *Serializer) WriteString(v string) {
	ser.WriteBytes([]byte(v))
}

// Something somewhere already knows how long this byte string will be
func (ser *Serializer) FixedBytes(v []byte) {
	ser.out.Write(v)
}

func (ser *Serializer) Bool(v bool) {
	if v {
		ser.out.WriteByte(1)
	} else {
		ser.out.WriteByte(0)
	}
}

func (ser *Serializer) Struct(x Marshaler) {
	x.MarshalBCS(ser)
}

func (ser *Serializer) ToBytes() []byte {
	return ser.out.Bytes()
}

func SerializeSequence[AT []T, T any](x AT, bcs *Serializer) {
	bcs.Uleb128(uint32(len(x)))
	for i, v := range x {
		mv, ok := any(v).(Marshaler)
		if ok {
			mv.MarshalBCS(bcs)
			continue
		}
		mv, ok = any(&v).(Marshaler)
		if ok {
			mv.MarshalBCS(bcs)
			continue
		}
		bcs.SetError(fmt.Errorf("could not serialize sequence[%d] member of %T", i, v))
		return
	}
}

func DeserializeSequence[T any](bcs *Deserializer) []T {
	slen := bcs.Uleb128()
	if bcs.Error() != nil {
		return nil
	}
	out := make([]T, slen)
	for i := 0; i < int(slen); i++ {
		v := &(out[i])
		mv, ok := any(v).(Unmarshaler)
		if ok {
			mv.UnmarshalBCS(bcs)
		} else {
			bcs.SetError(fmt.Errorf("could not deserialize sequence[%d] member of %T", i, v))
			return nil
		}
	}
	return out
}

// DeserializeMapToSlices returns two slices []K and []V of equal length that are equivalent to map[K]V but may represent types that are not valid Go map keys.
func DeserializeMapToSlices[K, V any](bcs *Deserializer) (keys []K, values []V) {
	count := bcs.Uleb128()
	keys = make([]K, 0, count)
	values = make([]V, 0, count)
	for _ = range count {
		var nextk K
		var nextv V
		switch sv := any(&nextk).(type) {
		case Unmarshaler:
			sv.UnmarshalBCS(bcs)
		case *string:
			*sv = bcs.ReadString()
		}
		switch sv := any(&nextv).(type) {
		case Unmarshaler:
			sv.UnmarshalBCS(bcs)
		case *string:
			*sv = bcs.ReadString()
		case *[]byte:
			*sv = bcs.ReadBytes()
		}
		keys = append(keys, nextk)
		values = append(values, nextv)
	}
	return
}

// Serialize serializes a single item
func Serialize(value Marshaler) (bcsBlob []byte, err error) {
	var bcs Serializer
	value.MarshalBCS(&bcs)
	err = bcs.Error()
	if err != nil {
		return
	}
	bcsBlob = bcs.ToBytes()
	return
}

// Deserialize deserializes a single item
func Deserialize(dest Unmarshaler, bcsBlob []byte) error {
	bcs := Deserializer{
		source: bcsBlob,
		pos:    0,
		err:    nil,
	}
	dest.UnmarshalBCS(&bcs)
	return bcs.err
}

func NewDeserializer(bcsBlob []byte) *Deserializer {
	return &Deserializer{
		source: bcsBlob,
		pos:    0,
		err:    nil,
	}
}

type Deserializer struct {
	source []byte
	pos    int

	err error
}

// If there has been any error, return it
func (d *Deserializer) Error() error {
	return d.err
}

// If the data is well formed but nonsense, UnmarshalBCS() code can set error
func (d *Deserializer) SetError(err error) {
	d.err = err
}

func (d *Deserializer) Remaining() int {
	return len(d.source) - d.pos
}

func (d *Deserializer) setError(msg string, args ...any) {
	if d.err != nil {
		return
	}
	d.err = fmt.Errorf(msg, args...)
}

func (d *Deserializer) Bool() bool {
	v := false
	switch d.source[d.pos] {
	case 0:
		v = false
	case 1:
		v = true
	default:
		d.setError("bad bool at [%d]: %x", d.pos, d.source[d.pos])
	}
	return v
}

func (d *Deserializer) Uleb128() uint32 {
	var value uint32 = 0
	shift := 0

	for {
		b := d.source[d.pos]
		value = value | (uint32(b&0x7f) << shift)
		d.pos++
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		// TODO: if shift is too much, error
	}

	return value
}

func (d *Deserializer) ReadBytes() []byte {
	blen := d.Uleb128()
	if d.err != nil {
		return nil
	}
	out := make([]byte, blen)
	copy(out, d.source[d.pos:d.pos+int(blen)])
	d.pos += int(blen)
	return out
}

func (d *Deserializer) ReadFixedBytes(blen int) []byte {
	out := make([]byte, blen)
	copy(out, d.source[d.pos:d.pos+blen])
	d.pos += blen
	return out
}

func (d *Deserializer) ReadFixedBytesInto(dest []byte) {
	blen := len(dest)
	copy(dest, d.source[d.pos:d.pos+blen])
	d.pos += blen
}

func (d *Deserializer) ReadString() string {
	b := d.ReadBytes()
	return string(b)
}

func (d *Deserializer) U8() uint8 {
	out := d.source[d.pos]
	d.pos++
	return out
}

func (d *Deserializer) U16() uint16 {
	out := binary.LittleEndian.Uint16(d.source[d.pos : d.pos+2])
	d.pos += 2
	return out
}

func (d *Deserializer) U32() uint32 {
	out := binary.LittleEndian.Uint32(d.source[d.pos : d.pos+4])
	d.pos += 4
	return out
}

func (d *Deserializer) U64() uint64 {
	out := binary.LittleEndian.Uint64(d.source[d.pos : d.pos+8])
	d.pos += 8
	return out
}

func (d *Deserializer) U128() big.Int {
	var bytesBigEndian [16]byte
	copy(bytesBigEndian[:], d.source[d.pos:d.pos+16])
	d.pos += 16
	reverse(bytesBigEndian[:])
	var out big.Int
	out.SetBytes(bytesBigEndian[:])
	return out
}

func (d *Deserializer) U256() big.Int {
	var bytesBigEndian [32]byte
	copy(bytesBigEndian[:], d.source[d.pos:d.pos+32])
	d.pos += 32
	reverse(bytesBigEndian[:])
	var out big.Int
	out.SetBytes(bytesBigEndian[:])
	return out
}

func (d *Deserializer) Struct(x Unmarshaler) {
	x.UnmarshalBCS(d)
}
