package arkapi

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	"github.com/google/uuid"
)

func createSyntheticArchive(t *testing.T, path string, className string) {
	t.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int32(7))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(1))
	buf.Write(id[:])
	writeArkString(&buf, className)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	writeStringArray(&buf, []string{"Object_0"})
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(-1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(128))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write archive fixture: %v", err)
	}
}

func writeStringArray(buf *bytes.Buffer, values []string) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		writeArkString(buf, value)
	}
}
