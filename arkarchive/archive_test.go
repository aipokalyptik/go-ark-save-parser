package arkarchive

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestParseArchiveReadsVersionAndObjectHeaders(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	testfixtures.WriteInt32(&buf, 7)
	testfixtures.WriteInt32(&buf, 11)
	testfixtures.WriteInt32(&buf, 22)
	testfixtures.WriteInt32(&buf, 1)
	buf.Write(id[:])
	testfixtures.WriteArkString(&buf, "/Script/ShooterGame.PrimalTribeData")
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteStringArray(&buf, []string{"TribeData_0"})
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, -1)
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, 128)
	testfixtures.WriteUInt32(&buf, 0)

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
	testfixtures.WriteInt32(&buf, 6)

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
	testfixtures.WriteInt32(&buf, 6)
	testfixtures.WriteInt32(&buf, 0)

	_, err := Parse(buf.Bytes(), Options{Format: FormatAuto})
	if !errors.Is(err, ErrLegacyArchiveUnsupported) {
		t.Fatalf("Parse() error = %v, want ErrLegacyArchiveUnsupported", err)
	}
}

func TestParseArchiveRejectsModernFormatMismatch(t *testing.T) {
	var buf bytes.Buffer
	testfixtures.WriteInt32(&buf, 6)

	_, err := Parse(buf.Bytes(), Options{Format: FormatModern, HeaderOnly: true})
	if err == nil {
		t.Fatalf("Parse() error = nil, want modern format mismatch")
	}
}

func TestParseArchiveReadsClusterDinoWithoutVersionPrefix(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	testfixtures.WriteInt32(&buf, 11)
	testfixtures.WriteInt32(&buf, 22)
	testfixtures.WriteInt32(&buf, 1)
	buf.Write(id[:])
	testfixtures.WriteArkString(&buf, "/Game/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C")
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteStringArray(&buf, []string{"Raptor_0"})
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, -1)
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, 96)
	testfixtures.WriteUInt32(&buf, 0)

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

func TestParseEmbeddedCryopodPayloadReadsObjectProperties(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	payload := syntheticEmbeddedCryopodPayload(t, dinoID, statusID)

	archive, err := ParseEmbeddedCryopodPayload(payload, 1<<20)
	if err != nil {
		t.Fatalf("ParseEmbeddedCryopodPayload() error = %v", err)
	}
	if archive.Version != 0x0407 || archive.Legacy {
		t.Fatalf("archive = %#v, want modern embedded archive version 0x0407", archive)
	}
	if len(archive.Objects) != 2 {
		t.Fatalf("Objects length = %d, want 2", len(archive.Objects))
	}
	if archive.Objects[0].UUID != dinoID || archive.Objects[0].ClassName != "Dino" {
		t.Fatalf("dino object = %#v", archive.Objects[0])
	}
	if len(archive.Objects[0].Properties) != 1 || archive.Objects[0].Properties[0].Name != "DinoID1" || archive.Objects[0].Properties[0].Value != int32(1001) {
		t.Fatalf("dino properties = %#v, want DinoID1 1001", archive.Objects[0].Properties)
	}
	if archive.Objects[1].UUID != statusID {
		t.Fatalf("status UUID = %s, want %s", archive.Objects[1].UUID, statusID)
	}
	if len(archive.Objects[1].Properties) != 1 || archive.Objects[1].Properties[0].Name != "BaseCharacterLevel" || archive.Objects[1].Properties[0].Value != int32(12) {
		t.Fatalf("status properties = %#v, want BaseCharacterLevel 12", archive.Objects[1].Properties)
	}
}

func TestParseArchiveReadsObjectProperties(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	testfixtures.WriteArchivePrefix(&buf, id, "/Script/ShooterGame.PrimalTribeData", []string{"TribeData_0"})
	offsetPos := buf.Len()
	testfixtures.WriteInt32(&buf, 0)
	testfixtures.WriteUInt32(&buf, 0)
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	testfixtures.WriteNameIntProperty(&buf, "TribeID", 12345)
	testfixtures.WriteArkString(&buf, "None")

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

func syntheticEmbeddedCryopodPayload(t *testing.T, dinoID uuid.UUID, statusID uuid.UUID) []byte {
	t.Helper()

	var decoded bytes.Buffer
	testfixtures.WriteInt32(&decoded, 0)
	testfixtures.WriteInt32(&decoded, 0)
	testfixtures.WriteUInt32(&decoded, 2)
	dinoOffsetPos := writeEmbeddedObjectHeader(&decoded, dinoID, "Dino", []string{"D0"})
	statusOffsetPos := writeEmbeddedObjectHeader(&decoded, statusID, "Status", []string{"S0"})

	dinoPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeEmbeddedNameIntProperty(&decoded, 0x10000001, 1001)
	writeEmbeddedNone(&decoded)
	statusPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeEmbeddedNameIntProperty(&decoded, 0x10000003, 12)
	writeEmbeddedNone(&decoded)

	binary.LittleEndian.PutUint32(decoded.Bytes()[dinoOffsetPos:dinoOffsetPos+4], uint32(dinoPropsOffset))
	binary.LittleEndian.PutUint32(decoded.Bytes()[statusOffsetPos:statusOffsetPos+4], uint32(statusPropsOffset))

	namesOffset := decoded.Len()
	testfixtures.WriteUInt32(&decoded, 4)
	testfixtures.WriteArkString(&decoded, "None")
	testfixtures.WriteArkString(&decoded, "DinoID1")
	testfixtures.WriteArkString(&decoded, "IntProperty")
	testfixtures.WriteArkString(&decoded, "BaseCharacterLevel")

	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(decoded.Bytes()); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}

	var payload bytes.Buffer
	testfixtures.WriteUInt32(&payload, 0x0407)
	testfixtures.WriteUInt32(&payload, uint32(decoded.Len()))
	testfixtures.WriteUInt32(&payload, uint32(namesOffset))
	payload.Write(compressed.Bytes())
	return payload.Bytes()
}

func writeEmbeddedObjectHeader(buf *bytes.Buffer, id uuid.UUID, className string, names []string) int {
	buf.Write(id[:])
	testfixtures.WriteArkString(buf, className)
	testfixtures.WriteUInt32(buf, 0)
	testfixtures.WriteStringArray(buf, names)
	testfixtures.WriteUInt32(buf, 0)
	testfixtures.WriteInt32(buf, 0)
	testfixtures.WriteUInt32(buf, 0)
	offsetPos := buf.Len()
	testfixtures.WriteInt32(buf, 0)
	testfixtures.WriteUInt32(buf, 0)
	return offsetPos
}

func writeEmbeddedNameIntProperty(buf *bytes.Buffer, nameID uint32, value int32) {
	writeEmbeddedName(buf, nameID)
	writeEmbeddedName(buf, 0x10000002)
	testfixtures.WriteInt32(buf, 4)
	testfixtures.WriteInt32(buf, 0)
	buf.WriteByte(0)
	testfixtures.WriteInt32(buf, value)
}

func writeEmbeddedNone(buf *bytes.Buffer) {
	writeEmbeddedName(buf, 0x10000000)
}

func writeEmbeddedName(buf *bytes.Buffer, nameID uint32) {
	testfixtures.WriteUInt32(buf, nameID)
	testfixtures.WriteInt32(buf, 0)
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
	if !archive.HasPropertyErrors() {
		t.Fatalf("HasPropertyErrors() = false, want true")
	}
	errors := archive.PropertyErrors()
	if len(errors) != 1 || errors[0].UUID != archive.Objects[0].UUID || errors[0].ClassName != archive.Objects[0].ClassName || errors[0].Err == nil {
		t.Fatalf("PropertyErrors() = %#v, want one object property error", errors)
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
	testfixtures.WriteArchivePrefix(&buf, id, "/Script/ShooterGame.PrimalTribeData", []string{"TribeData_0"})
	offsetPos := buf.Len()
	testfixtures.WriteInt32(&buf, 0)
	testfixtures.WriteUInt32(&buf, 0)
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	testfixtures.WriteNameIntProperty(&buf, "TribeID", 12345)
	testfixtures.WriteArkString(&buf, "Owner")
	testfixtures.WriteArkString(&buf, "ObjectProperty")
	testfixtures.WriteInt32(&buf, 2)
	testfixtures.WriteInt32(&buf, 0)
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, int16(5))
	testfixtures.WriteArkString(&buf, "None")
	return buf.Bytes()
}

func archiveWithBrokenObjectProperty() []byte {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	testfixtures.WriteArchivePrefix(&buf, id, "/Script/ShooterGame.PrimalLocalProfile", []string{"Profile_0"})
	offsetPos := buf.Len()
	testfixtures.WriteInt32(&buf, 0)
	testfixtures.WriteUInt32(&buf, 0)
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	testfixtures.WriteArkString(&buf, "Owner")
	testfixtures.WriteArkString(&buf, "ObjectProperty")
	testfixtures.WriteInt32(&buf, 2)
	testfixtures.WriteInt32(&buf, 0)
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, int16(5))
	testfixtures.WriteArkString(&buf, "None")
	return buf.Bytes()
}
