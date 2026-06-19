package arkprofile

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestOpenPlayerProfileLoadsLocalArchiveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	writeArchiveFile(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

	profile, err := OpenPlayerProfile(path)
	if err != nil {
		t.Fatalf("OpenPlayerProfile() error = %v", err)
	}
	if profile.Path != path {
		t.Fatalf("Path = %q, want %q", profile.Path, path)
	}
	if len(profile.Archive.Objects) != 1 {
		t.Fatalf("Archive objects = %d, want 1", len(profile.Archive.Objects))
	}
}

func TestOpenTribeSaveLoadsLocalArchiveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	writeArchiveFile(t, path, "/Script/ShooterGame.PrimalTribeData")

	tribe, err := OpenTribeSave(path)
	if err != nil {
		t.Fatalf("OpenTribeSave() error = %v", err)
	}
	if len(tribe.Archive.Objects) != 1 || tribe.Archive.Objects[0].ClassName != "/Script/ShooterGame.PrimalTribeData" {
		t.Fatalf("Tribe archive objects = %#v", tribe.Archive.Objects)
	}
}

func TestTribeSummaryReadsStructContainerFields(t *testing.T) {
	tribe := &TribeSave{}
	tribeData := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "TribeName", Type: arkproperty.TypeString, Value: "Porters"},
		{Name: "OwnerPlayerDataId", Type: arkproperty.TypeUInt32, Value: uint32(42)},
		{Name: "TribeID", Type: arkproperty.TypeInt, Value: int32(12345)},
		{Name: "MembersPlayerName", Type: arkproperty.TypeArray, Value: []any{"Ada", "Grace"}},
		{Name: "NumTribeDinos", Type: arkproperty.TypeInt, Value: int32(7)},
	}}
	tribe.Properties = arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "TribeData", Type: arkproperty.TypeStruct, Value: tribeData},
	}}

	summary, err := tribe.Summary()
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.Name != "Porters" || summary.OwnerID != 42 || summary.TribeID != 12345 || summary.NumDinos != 7 {
		t.Fatalf("Summary() = %#v", summary)
	}
	if len(summary.Members) != 2 || summary.Members[0] != "Ada" || summary.Members[1] != "Grace" {
		t.Fatalf("Summary().Members = %#v", summary.Members)
	}
}

func TestOpenTribeSaveSummaryUsesParsedArchiveProperties(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	writeTribeArchiveFile(t, path)

	tribe, err := OpenTribeSave(path)
	if err != nil {
		t.Fatalf("OpenTribeSave() error = %v", err)
	}
	summary, err := tribe.Summary()
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.Name != "Porters" || summary.TribeID != 12345 {
		t.Fatalf("Summary() = %#v, want Porters/12345", summary)
	}
}

func writeArchiveFile(t *testing.T, path string, className string) {
	t.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	writeInt32(&buf, 7)
	writeInt32(&buf, 0)
	writeInt32(&buf, 0)
	writeInt32(&buf, 1)
	buf.Write(id[:])
	writeArkString(&buf, className)
	writeUInt32(&buf, 0)
	writeStringArray(&buf, []string{"Object_0"})
	writeUInt32(&buf, 0)
	writeInt32(&buf, -1)
	writeUInt32(&buf, 0)
	writeInt32(&buf, 128)
	writeUInt32(&buf, 0)
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write archive fixture: %v", err)
	}
}

func writeTribeArchiveFile(t *testing.T, path string) {
	t.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var body bytes.Buffer
	writeNameStringProperty(&body, "TribeName", "Porters")
	writeNameIntProperty(&body, "TribeID", 12345)
	writeArkString(&body, "None")

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
	writeNameStructProperty(&buf, "TribeData", "TribeDataStruct", body.Bytes())
	writeArkString(&buf, "None")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write archive fixture: %v", err)
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

func writeNameIntProperty(buf *bytes.Buffer, name string, value int32) {
	writeArkString(buf, name)
	writeArkString(buf, "IntProperty")
	writeInt32(buf, 4)
	writeInt32(buf, 0)
	buf.WriteByte(0)
	writeInt32(buf, value)
}

func writeNameStringProperty(buf *bytes.Buffer, name string, value string) {
	writeArkString(buf, name)
	writeArkString(buf, "StrProperty")
	writeInt32(buf, int32(len(value)+5))
	writeInt32(buf, 0)
	buf.WriteByte(0)
	writeArkString(buf, value)
}

func writeNameStructProperty(buf *bytes.Buffer, name string, structType string, body []byte) {
	writeArkString(buf, name)
	writeArkString(buf, "StructProperty")
	writeUInt32(buf, 1)
	writeArkString(buf, structType)
	writeUInt32(buf, 1)
	writeArkString(buf, structType)
	writeUInt32(buf, 0)
	writeUInt32(buf, uint32(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
}
