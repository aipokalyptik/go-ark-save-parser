package arkbinary

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"unicode/utf16"

	"github.com/google/uuid"
)

type Reader struct {
	data []byte
	pos  int
	ctx  *Context
}

func NewReader(data []byte, ctx *Context) *Reader {
	if ctx == nil {
		ctx = NewContext()
	}
	return &Reader{data: data, ctx: ctx}
}

func (r *Reader) Position() int {
	return r.pos
}

func (r *Reader) Size() int {
	return len(r.data)
}

func (r *Reader) HasMore() bool {
	return r.pos < len(r.data)
}

func (r *Reader) HasNameTable() bool {
	return r.ctx.HasNameTable()
}

func (r *Reader) SetPosition(pos int) error {
	if pos < 0 || pos > len(r.data) {
		return fmt.Errorf("position %d outside buffer size %d", pos, len(r.data))
	}
	r.pos = pos
	return nil
}

func (r *Reader) Skip(count int) error {
	return r.SetPosition(r.pos + count)
}

func (r *Reader) read(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("negative read size %d", n)
	}
	if r.pos+n > len(r.data) {
		return nil, fmt.Errorf("buffer underflow: need %d bytes, have %d", n, len(r.data)-r.pos)
	}
	out := r.data[r.pos : r.pos+n]
	r.pos += n
	return out, nil
}

func (r *Reader) ReadBytes(n int) ([]byte, error) {
	b, err := r.read(n)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out, nil
}

func (r *Reader) Slice(start int, end int) ([]byte, error) {
	if start < 0 || end < start || end > len(r.data) {
		return nil, fmt.Errorf("slice [%d:%d] outside buffer size %d", start, end, len(r.data))
	}
	out := make([]byte, end-start)
	copy(out, r.data[start:end])
	return out, nil
}

func (r *Reader) ReadByte() (byte, error) {
	b, err := r.read(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *Reader) ReadBool() (bool, error) {
	b, err := r.ReadByte()
	return b != 0, err
}

func (r *Reader) ReadInt16() (int16, error) {
	b, err := r.read(2)
	if err != nil {
		return 0, err
	}
	return int16(binary.LittleEndian.Uint16(b)), nil
}

func (r *Reader) ReadUInt16() (uint16, error) {
	b, err := r.read(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func (r *Reader) ReadInt32() (int32, error) {
	b, err := r.read(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(b)), nil
}

func (r *Reader) ReadUInt32() (uint32, error) {
	b, err := r.read(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (r *Reader) PeekUInt32() (uint32, error) {
	pos := r.pos
	v, err := r.ReadUInt32()
	r.pos = pos
	return v, err
}

func (r *Reader) ReadInt64() (int64, error) {
	b, err := r.read(8)
	if err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

func (r *Reader) ReadUInt64() (uint64, error) {
	b, err := r.read(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (r *Reader) ReadFloat32() (float32, error) {
	bits, err := r.ReadUInt32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

func (r *Reader) ReadFloat64() (float64, error) {
	bits, err := r.ReadUInt64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

func (r *Reader) ReadUUID() (uuid.UUID, error) {
	b, err := r.read(16)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.FromBytes(b)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *Reader) ReadString() (*string, error) {
	length, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, nil
	}

	if length < 0 {
		chars := int(-length)
		toRead := chars*2 - 2
		raw, err := r.read(toRead)
		if err != nil {
			return nil, err
		}
		terminator, err := r.ReadUInt16()
		if err != nil {
			return nil, err
		}
		if terminator != 0 {
			return nil, fmt.Errorf("string terminator is not zero: %d", terminator)
		}
		u16 := make([]uint16, 0, len(raw)/2)
		for i := 0; i < len(raw); i += 2 {
			u16 = append(u16, binary.LittleEndian.Uint16(raw[i:i+2]))
		}
		s := string(utf16.Decode(u16))
		return &s, nil
	}

	raw, err := r.read(int(length) - 1)
	if err != nil {
		return nil, err
	}
	terminator, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if terminator != 0 {
		return nil, fmt.Errorf("string terminator is not zero: %d", terminator)
	}
	s := string(raw)
	return &s, nil
}

func (r *Reader) ReadName(defaultValue string) (string, error) {
	if !r.ctx.HasNameTable() {
		s, err := r.ReadString()
		if err != nil {
			return "", err
		}
		if s == nil {
			return defaultValue, nil
		}
		return *s, nil
	}

	start := r.pos
	nameID, err := r.ReadUInt32()
	if err != nil {
		return "", err
	}
	name, ok := r.ctx.Name(nameID)
	if !ok {
		if defaultValue != "" {
			name = defaultValue
		} else {
			return "", fmt.Errorf("name is none for index %#x at position %d", nameID, start)
		}
	}

	if isSuffixedVolumeName(name) {
		suffix, err := r.ReadInt32()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s_%#x", name, suffix), nil
	}

	if _, err := r.ReadInt32(); err != nil {
		return "", err
	}
	return name, nil
}

func (r *Reader) ReadNameGeneratedUnknown(defaultValue string) (string, error) {
	if !r.ctx.HasNameTable() {
		return r.ReadName(defaultValue)
	}

	start := r.pos
	nameID, err := r.ReadUInt32()
	if err != nil {
		return "", err
	}
	name, ok := r.ctx.Name(nameID)
	if !ok {
		name = fmt.Sprintf("Unknown_%d", nameID)
	}
	if name == "" && defaultValue != "" {
		name = defaultValue
	}

	if isSuffixedVolumeName(name) {
		suffix, err := r.ReadInt32()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s_%#x", name, suffix), nil
	}

	if _, err := r.ReadInt32(); err != nil {
		return "", fmt.Errorf("read generated name suffix at position %d: %w", start, err)
	}
	return name, nil
}

func isSuffixedVolumeName(name string) bool {
	if strings.Contains(name, "/") || strings.Contains(name, "'") {
		return false
	}
	return name == "NPCZoneVolume" ||
		strings.Contains(name, "NPCZoneVolume_") ||
		strings.Contains(name, "_NPCZoneVolume") ||
		strings.Contains(name, "NPCCountVolume")
}
