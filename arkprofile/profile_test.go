package arkprofile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/safefile"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
)

func TestOpenPlayerProfileLoadsLocalArchiveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WriteArchive(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

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

func TestOpenPlayerProfilePlayerUsesParsedArchiveProperties(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WritePlayerArchive(t, path)

	profile, err := OpenPlayerProfile(path)
	if err != nil {
		t.Fatalf("OpenPlayerProfile() error = %v", err)
	}
	player, err := profile.Player()
	if err != nil {
		t.Fatalf("Player() error = %v", err)
	}
	if player.PlayerDataID != 42 || player.CharacterName != "Survivor" || player.TribeID != 777 {
		t.Fatalf("Player() = %#v", player)
	}
}

func TestOpenPlayerProfileExposesPropertyErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	writeBrokenArchive(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

	profile, err := OpenPlayerProfile(path)
	if err != nil {
		t.Fatalf("OpenPlayerProfile() error = %v", err)
	}
	if !profile.HasPropertyErrors() || profile.PropertyError() == nil {
		t.Fatalf("profile property error state = %v, %v; want recorded error", profile.HasPropertyErrors(), profile.PropertyError())
	}
}

func TestOpenPlayerProfileRejectsArchiveAboveConfiguredLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	writeSparseFile(t, path, 1024)

	_, err := OpenPlayerProfileWithOptions(path, Options{MaxFileSize: 16})
	if !errors.Is(err, safefile.ErrFileTooLarge) {
		t.Fatalf("OpenPlayerProfileWithOptions() error = %v, want ErrFileTooLarge", err)
	}
}

func TestOpenTribeSaveLoadsLocalArchiveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteArchive(t, path, "/Script/ShooterGame.PrimalTribeData")

	tribe, err := OpenTribeSave(path)
	if err != nil {
		t.Fatalf("OpenTribeSave() error = %v", err)
	}
	if len(tribe.Archive.Objects) != 1 || tribe.Archive.Objects[0].ClassName != "/Script/ShooterGame.PrimalTribeData" {
		t.Fatalf("Tribe archive objects = %#v", tribe.Archive.Objects)
	}
}

func TestOpenTribeSaveExposesPropertyErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	writeBrokenArchive(t, path, "/Script/ShooterGame.PrimalTribeData")

	tribe, err := OpenTribeSave(path)
	if err != nil {
		t.Fatalf("OpenTribeSave() error = %v", err)
	}
	if !tribe.HasPropertyErrors() || tribe.PropertyError() == nil {
		t.Fatalf("tribe property error state = %v, %v; want recorded error", tribe.HasPropertyErrors(), tribe.PropertyError())
	}
}

func TestOpenTribeSaveRejectsArchiveAboveConfiguredLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	writeSparseFile(t, path, 1024)

	_, err := OpenTribeSaveWithOptions(path, Options{MaxFileSize: 16})
	if !errors.Is(err, safefile.ErrFileTooLarge) {
		t.Fatalf("OpenTribeSaveWithOptions() error = %v, want ErrFileTooLarge", err)
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
	testfixtures.WriteTribeArchive(t, path)

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

func TestOpenTribeSaveTribeUsesParsedArchiveProperties(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteTribeArchive(t, path)

	save, err := OpenTribeSave(path)
	if err != nil {
		t.Fatalf("OpenTribeSave() error = %v", err)
	}
	tribe, err := save.Tribe()
	if err != nil {
		t.Fatalf("Tribe() error = %v", err)
	}
	if tribe.Name != "Porters" || tribe.TribeID != 12345 || tribe.OwnerID != 42 || tribe.NumDinos != 7 {
		t.Fatalf("Tribe() = %#v, want parsed tribe details", tribe)
	}
}

func writeSparseFile(t *testing.T, path string, size int64) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create sparse file: %v", err)
	}
	if err := file.Truncate(size); err != nil {
		_ = file.Close()
		t.Fatalf("truncate sparse file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close sparse file: %v", err)
	}
}

func writeBrokenArchive(t *testing.T, path string, className string) {
	t.Helper()
	var props bytes.Buffer
	writeProfileStringProperty(&props, "ValidName", "value")
	writeProfileArkString(&props, "Broken")
	writeProfileArkString(&props, "ObjectProperty")
	_ = binary.Write(&props, binary.LittleEndian, int32(2))
	_ = binary.Write(&props, binary.LittleEndian, int32(0))
	props.WriteByte(0)
	_ = binary.Write(&props, binary.LittleEndian, int16(5))
	writeProfileArkString(&props, "None")

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int32(7))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(1))
	buf.Write(make([]byte, 16))
	writeProfileArkString(&buf, className)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(-1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	buf.Write(props.Bytes())
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write broken archive: %v", err)
	}
}

func writeProfileStringProperty(buf *bytes.Buffer, name string, value string) {
	writeProfileArkString(buf, name)
	writeProfileArkString(buf, "StrProperty")
	bodySize := 4 + len(value) + 1
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	writeProfileArkString(buf, value)
}

func writeProfileArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}
