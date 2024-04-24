package aptos

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
)

type BCSStruct interface {
	MarshalBCS(*Serializer)
	UnmarshalBCS(*Deserializer)
}

type Serializer struct {
	out bytes.Buffer
}

func (bcs *Serializer) U8(v uint8) {
	bcs.out.WriteByte(v)
}

func (bcs *Serializer) U16(v uint16) {
	var ub [2]byte
	binary.LittleEndian.PutUint16(ub[:], v)
	bcs.out.Write(ub[:])
}

func (bcs *Serializer) U32(v uint32) {
	var ub [4]byte
	binary.LittleEndian.PutUint32(ub[:], v)
	bcs.out.Write(ub[:])
}

func (bcs *Serializer) U64(v uint64) {
	var ub [8]byte
	binary.LittleEndian.PutUint64(ub[:], v)
	bcs.out.Write(ub[:])
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

func (bcs *Serializer) U128(v big.Int) {
	var ub [16]byte
	v.FillBytes(ub[:])
	reverse(ub[:])
	bcs.out.Write(ub[:])
}

func (bcs *Serializer) U256(v big.Int) {
	var ub [32]byte
	v.FillBytes(ub[:])
	reverse(ub[:])
	bcs.out.Write(ub[:])
}

func (bcs *Serializer) Uleb128(v uint64) {
	for v > 0x80 {
		nb := uint8(v & 0x7f)
		bcs.out.WriteByte(0x80 | nb)
		v = v >> 7
	}
	bcs.out.WriteByte(uint8(v & 0x7f))
}

func (bcs *Serializer) WriteBytes(v []byte) {
	bcs.Uleb128(uint64(len(v)))
	bcs.out.Write(v)
}

// Something somewhere already knows how long this byte string will be
func (bcs *Serializer) FixedBytes(v []byte) {
	bcs.out.Write(v)
}

func (bcs *Serializer) Bool(v bool) {
	if v {
		bcs.out.WriteByte(1)
	} else {
		bcs.out.WriteByte(0)
	}
}

func (bcs *Serializer) Struct(x BCSStruct) {
	x.MarshalBCS(bcs)
}

func (bcs *Serializer) ToBytes() []byte {
	return bcs.out.Bytes()
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

func (d *Deserializer) Uleb128() uint64 {
	var value uint64 = 0
	shift := 0

	for {
		b := d.source[d.pos]
		value = value | (uint64(b&0x7f) << shift)
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

func (d *Deserializer) Struct(x BCSStruct) {
	x.UnmarshalBCS(d)
}
