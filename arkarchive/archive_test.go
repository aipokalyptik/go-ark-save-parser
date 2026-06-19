package arkarchive

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/google/uuid"
)

func TestParseArchiveReadsVersionAndObjectHeaders(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	writeInt32(&buf, 7)
	writeInt32(&buf, 11)
	writeInt32(&buf, 22)
	writeInt32(&buf, 1)
	buf.Write(id[:])
	writeArkString(&buf, "/Script/ShooterGame.PrimalTribeData")
	writeUInt32(&buf, 0)
	writeStringArray(&buf, []string{"TribeData_0"})
	writeUInt32(&buf, 0)
	writeInt32(&buf, -1)
	writeUInt32(&buf, 0)
	writeInt32(&buf, 128)
	writeUInt32(&buf, 0)

	archive, err := Parse(buf.Bytes(), Options{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if archive.Version != 7 {
		t.Fatalf("Version = %d, want 7", archive.Version)
	}
	if len(archive.Objects) != 1 {
		t.Fatalf("Objects length = %d, want 1", len(archive.Objects))
	}
	obj := archive.Objects[0]
	if obj.UUID != id {
		t.Fatalf("UUID = %s, want %s", obj.UUID, id)
	}
	if obj.ClassName != "/Script/ShooterGame.PrimalTribeData" {
		t.Fatalf("ClassName = %q", obj.ClassName)
	}
	if len(obj.Names) != 1 || obj.Names[0] != "TribeData_0" {
		t.Fatalf("Names = %#v, want TribeData_0", obj.Names)
	}
	if obj.PropertiesOffset != 128 {
		t.Fatalf("PropertiesOffset = %d, want 128", obj.PropertiesOffset)
	}
}

func TestParseArchiveDetectsLegacyFormat(t *testing.T) {
	var buf bytes.Buffer
	writeInt32(&buf, 6)

	archive, err := Parse(buf.Bytes(), Options{HeaderOnly: true})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !archive.Legacy {
		t.Fatalf("Legacy = false, want true")
	}
}

func writeInt32(buf *bytes.Buffer, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeUInt32(buf *bytes.Buffer, value uint32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}

func writeStringArray(buf *bytes.Buffer, values []string) {
	writeUInt32(buf, uint32(len(values)))
	for _, value := range values {
		writeArkString(buf, value)
	}
}
