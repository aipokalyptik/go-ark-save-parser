package arkprofile

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

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
	tribe.Properties = map[string]any{
		"TribeData": map[string]any{
			"TribeName":         "Porters",
			"OwnerPlayerDataId": uint32(42),
			"TribeID":           int32(12345),
			"MembersPlayerName": []any{"Ada", "Grace"},
			"NumTribeDinos":     int32(7),
		},
	}

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
