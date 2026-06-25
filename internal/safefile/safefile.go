package safefile

import (
	"errors"
	"fmt"
	"os"
)

var ErrFileTooLarge = errors.New("file exceeds maximum allowed size")

func ReadFile(path string, maxBytes int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file")
	}
	if maxBytes > 0 && info.Size() > maxBytes {
		return nil, fmt.Errorf("%w: size %d > limit %d", ErrFileTooLarge, info.Size(), maxBytes)
	}
	return os.ReadFile(path)
}
