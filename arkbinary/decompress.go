package arkbinary

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var ErrInflatedDataTooLarge = errors.New("inflated data exceeds maximum allowed size")
var ErrUnsupportedEmbeddedDataVersion = errors.New("unsupported embedded compressed data version")

type EmbeddedCompressedData struct {
	Version uint32
	Data    []byte
	Names   map[uint32]string
	Context *Context
}

func InflateZlib(data []byte) ([]byte, error) {
	return InflateZlibWithLimit(data, 0)
}

func InflateZlibWithLimit(data []byte, maxBytes int64) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if maxBytes <= 0 {
		return io.ReadAll(reader)
	}
	limited := &io.LimitedReader{R: reader, N: maxBytes + 1}
	out, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(out)) > maxBytes {
		return nil, fmt.Errorf("%w: limit %d", ErrInflatedDataTooLarge, maxBytes)
	}
	return out, nil
}

func DecodeEmbeddedCompressedData(data []byte, maxInflatedBytes int64) (EmbeddedCompressedData, error) {
	if len(data) < 12 {
		return EmbeddedCompressedData{}, fmt.Errorf("embedded compressed data header requires 12 bytes, got %d", len(data))
	}
	version := binary.LittleEndian.Uint32(data[0:4])
	if version < 0x0407 {
		return EmbeddedCompressedData{}, fmt.Errorf("%w: %#x", ErrUnsupportedEmbeddedDataVersion, version)
	}
	inflatedSize := binary.LittleEndian.Uint32(data[4:8])
	namesOffset := binary.LittleEndian.Uint32(data[8:12])
	limit := int64(inflatedSize)
	if maxInflatedBytes > 0 && maxInflatedBytes < limit {
		limit = maxInflatedBytes
	}
	inflated, err := InflateZlibWithLimit(data[12:], limit)
	if err != nil {
		return EmbeddedCompressedData{}, err
	}
	if uint32(len(inflated)) != inflatedSize {
		return EmbeddedCompressedData{}, fmt.Errorf("embedded compressed data inflated size = %d, want %d", len(inflated), inflatedSize)
	}
	decoded, err := WildcardDecompress(inflated)
	if err != nil {
		return EmbeddedCompressedData{}, err
	}
	if namesOffset > uint32(len(decoded)) {
		return EmbeddedCompressedData{}, fmt.Errorf("embedded compressed data names offset %d outside decoded size %d", namesOffset, len(decoded))
	}
	names, err := readEmbeddedNameTable(decoded, int(namesOffset))
	if err != nil {
		return EmbeddedCompressedData{}, err
	}
	ctx := NewContext()
	ctx.SetNames(names)
	return EmbeddedCompressedData{Version: version, Data: decoded, Names: names, Context: ctx}, nil
}

func readEmbeddedNameTable(data []byte, offset int) (map[uint32]string, error) {
	r := NewReader(data, nil)
	if err := r.SetPosition(offset); err != nil {
		return nil, err
	}
	count, err := r.ReadUInt32()
	if err != nil {
		return nil, err
	}
	names := make(map[uint32]string, count)
	for i := uint32(0); i < count; i++ {
		value, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		if value == nil {
			names[0x10000000|i] = ""
			continue
		}
		names[0x10000000|i] = *value
	}
	return names, nil
}

func WildcardDecompress(input []byte) ([]byte, error) {
	const (
		stateNone = iota
		stateEscape
		stateSwitch
	)

	state := stateNone
	out := make([]byte, 0, len(input))
	queue := make([]byte, 0)
	pos := 0

	for pos < len(input) || len(queue) > 0 {
		if len(queue) > 0 {
			out = append(out, queue[0])
			queue = queue[1:]
			continue
		}

		next := input[pos]
		pos++

		if state == stateSwitch {
			out = append(out, 0xf0|((next&0xf0)>>4))
			queue = append(queue, 0xf0|(next&0x0f))
			state = stateNone
			continue
		}

		if state == stateNone {
			switch {
			case next == 0xf0:
				state = stateEscape
				continue
			case next == 0xf1:
				state = stateSwitch
				continue
			case next >= 0xf2 && next < 0xff:
				count := int(next & 0x0f)
				for i := 0; i < count; i++ {
					queue = append(queue, 0)
				}
				continue
			case next == 0xff:
				if pos+2 > len(input) {
					return nil, fmt.Errorf("unexpected end of stream after 0xff")
				}
				b1 := input[pos]
				b2 := input[pos+1]
				pos += 2
				queue = append(queue, 0, 0, 0, b1, 0, 0, 0, b2, 0, 0, 0)
				continue
			}
		}

		state = stateNone
		out = append(out, next)
	}

	return out, nil
}
