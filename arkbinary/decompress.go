package arkbinary

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
)

var ErrInflatedDataTooLarge = errors.New("inflated data exceeds maximum allowed size")

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
