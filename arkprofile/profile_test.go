package arkprofile

import (
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
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
