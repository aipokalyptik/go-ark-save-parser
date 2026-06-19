package arkarchive

import (
	"bytes"
	"encoding/binary"
	"errors"
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

func TestParseArchiveRejectsLegacyFormatWhenObjectsRequested(t *testing.T) {
	var buf bytes.Buffer
	writeInt32(&buf, 6)
	writeInt32(&buf, 0)

	_, err := Parse(buf.Bytes(), Options{Format: FormatAuto})
	if !errors.Is(err, ErrLegacyArchiveUnsupported) {
		t.Fatalf("Parse() error = %v, want ErrLegacyArchiveUnsupported", err)
	}
}

func TestParseArchiveRejectsModernFormatMismatch(t *testing.T) {
	var buf bytes.Buffer
	writeInt32(&buf, 6)

	_, err := Parse(buf.Bytes(), Options{Format: FormatModern, HeaderOnly: true})
	if err == nil {
		t.Fatalf("Parse() error = nil, want modern format mismatch")
	}
}

func TestParseArchiveReadsClusterDinoWithoutVersionPrefix(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	writeInt32(&buf, 11)
	writeInt32(&buf, 22)
	writeInt32(&buf, 1)
	buf.Write(id[:])
	writeArkString(&buf, "/Game/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C")
	writeUInt32(&buf, 0)
	writeStringArray(&buf, []string{"Raptor_0"})
	writeUInt32(&buf, 0)
	writeInt32(&buf, -1)
	writeUInt32(&buf, 0)
	writeInt32(&buf, 96)
	writeUInt32(&buf, 0)

	archive, err := Parse(buf.Bytes(), Options{Format: FormatClusterDino})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if archive.Version != 7 || archive.Legacy {
		t.Fatalf("Archive = %#v, want version 7 modern cluster dino", archive)
	}
	if len(archive.Objects) != 1 || archive.Objects[0].UUID != id {
		t.Fatalf("Objects = %#v, want one object with id %s", archive.Objects, id)
	}
}

func TestParseArchiveReadsObjectProperties(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	writeInt32(&buf, 7)
	writeInt32(&buf, 0)
	writeInt32(&buf, 0)
	writeInt32(&buf, 1)
	buf.Write(id[:])
	writeArkString(&buf, "/Script/ShooterGame.PrimalTribeData")
	writeUInt32(&buf, 0)
	writeStringArray(&buf, []string{"TribeData_0"})
	writeUInt32(&buf, 0)
	writeInt32(&buf, -1)
	writeUInt32(&buf, 0)
	offsetPos := buf.Len()
	writeInt32(&buf, 0)
	writeUInt32(&buf, 0)
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	writeArchiveIntProperty(&buf, "TribeID", 12345)
	writeArkString(&buf, "None")

	archive, err := Parse(buf.Bytes(), Options{Format: FormatModern})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(archive.Objects) != 1 {
		t.Fatalf("Objects length = %d, want 1", len(archive.Objects))
	}
	props := archive.Objects[0].Properties
	if len(props) != 1 {
		t.Fatalf("Properties length = %d, want 1", len(props))
	}
	if props[0].Name != "TribeID" || props[0].Value != int32(12345) {
		t.Fatalf("Property = %#v, want TribeID 12345", props[0])
	}
}

func TestParseArchiveRecordsPropertyErrorsByDefault(t *testing.T) {
	raw := archiveWithBrokenObjectProperty()

	archive, err := Parse(raw, Options{Format: FormatModern})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(archive.Objects) != 1 {
		t.Fatalf("Objects length = %d, want 1", len(archive.Objects))
	}
	if archive.Objects[0].PropertyError == nil {
		t.Fatalf("PropertyError = nil, want recorded parse error")
	}
}

func TestParseArchiveKeepsPartialPropertiesWhenPropertyErrorIsRecorded(t *testing.T) {
	raw := archiveWithValidThenBrokenProperties()

	archive, err := Parse(raw, Options{Format: FormatModern})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if archive.Objects[0].PropertyError == nil {
		t.Fatalf("PropertyError = nil, want recorded parse error")
	}
	props := archive.Objects[0].Properties
	if len(props) != 1 {
		t.Fatalf("Properties length = %d, want 1 partial property", len(props))
	}
	if props[0].Name != "TribeID" || props[0].Value != int32(12345) {
		t.Fatalf("Partial property = %#v, want TribeID 12345", props[0])
	}
}

func TestParseArchiveStrictPropertiesReturnsPropertyErrors(t *testing.T) {
	raw := archiveWithBrokenObjectProperty()

	_, err := Parse(raw, Options{Format: FormatModern, StrictProperties: true})
	if err == nil {
		t.Fatalf("Parse() error = nil, want property parse error")
	}
}

func archiveWithValidThenBrokenProperties() []byte {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	writeInt32(&buf, 7)
	writeInt32(&buf, 0)
	writeInt32(&buf, 0)
	writeInt32(&buf, 1)
	buf.Write(id[:])
	writeArkString(&buf, "/Script/ShooterGame.PrimalTribeData")
	writeUInt32(&buf, 0)
	writeStringArray(&buf, []string{"TribeData_0"})
	writeUInt32(&buf, 0)
	writeInt32(&buf, -1)
	writeUInt32(&buf, 0)
	offsetPos := buf.Len()
	writeInt32(&buf, 0)
	writeUInt32(&buf, 0)
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	writeArchiveIntProperty(&buf, "TribeID", 12345)
	writeArkString(&buf, "Owner")
	writeArkString(&buf, "ObjectProperty")
	writeInt32(&buf, 2)
	writeInt32(&buf, 0)
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, int16(5))
	writeArkString(&buf, "None")
	return buf.Bytes()
}

func archiveWithBrokenObjectProperty() []byte {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	writeInt32(&buf, 7)
	writeInt32(&buf, 0)
	writeInt32(&buf, 0)
	writeInt32(&buf, 1)
	buf.Write(id[:])
	writeArkString(&buf, "/Script/ShooterGame.PrimalLocalProfile")
	writeUInt32(&buf, 0)
	writeStringArray(&buf, []string{"Profile_0"})
	writeUInt32(&buf, 0)
	writeInt32(&buf, -1)
	writeUInt32(&buf, 0)
	offsetPos := buf.Len()
	writeInt32(&buf, 0)
	writeUInt32(&buf, 0)
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	writeArkString(&buf, "Owner")
	writeArkString(&buf, "ObjectProperty")
	writeInt32(&buf, 2)
	writeInt32(&buf, 0)
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, int16(5))
	writeArkString(&buf, "None")
	return buf.Bytes()
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

func writeArchiveIntProperty(buf *bytes.Buffer, name string, value int32) {
	writeArkString(buf, name)
	writeArkString(buf, "IntProperty")
	writeInt32(buf, 4)
	writeInt32(buf, 0)
	buf.WriteByte(0)
	writeInt32(buf, value)
}
