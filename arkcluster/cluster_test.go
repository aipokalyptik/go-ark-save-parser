package arkcluster

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/internal/safefile"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestDiscoverFindsLocalClusterFilesOnly(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteArchive(t, filepath.Join(dir, "123.arkprofile"), "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")
	testfixtures.WriteArchive(t, filepath.Join(dir, "456.arktribe"), "/Script/ShooterGame.PrimalTribeData")
	if err := os.WriteFile(filepath.Join(dir, "Valguero_WP.ark"), []byte("not sqlite"), 0o600); err != nil {
		t.Fatalf("write map placeholder: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write unrelated extensionless file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".hidden_cluster"), []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write hidden file: %v", err)
	}

	files, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(files) != 1 || files[0].Path != clusterPath || files[0].ID != "EOS_abc123" {
		t.Fatalf("Discover() = %#v, want only cluster file", files)
	}
}

func TestOpenLoadsLocalClusterArchiveMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	testfixtures.WriteArchive(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

	data, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if data.ID != "EOS_abc123" || data.Archive.Version != 7 || len(data.Archive.Objects) != 1 {
		t.Fatalf("ClusterData = %#v", data)
	}
	if data.Archive.Objects[0].ClassName != "/Script/ShooterGame.ArkCloudInventoryData" {
		t.Fatalf("ClassName = %q", data.Archive.Objects[0].ClassName)
	}
}

func TestOpenRejectsClusterArchiveAboveConfiguredLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	testfixtures.WriteSparseFile(t, path, 1024)

	_, err := OpenWithOptions(path, Options{MaxFileSize: 16})
	if !errors.Is(err, safefile.ErrFileTooLarge) {
		t.Fatalf("OpenWithOptions() error = %v, want ErrFileTooLarge", err)
	}
}

func TestOpenDirectoryRejectsClusterArchiveAboveConfiguredLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "EOS_abc123")
	testfixtures.WriteSparseFile(t, path, 1024)

	_, err := OpenDirectoryWithOptions(dir, Options{MaxFileSize: 16})
	if !errors.Is(err, safefile.ErrFileTooLarge) {
		t.Fatalf("OpenDirectoryWithOptions() error = %v, want ErrFileTooLarge", err)
	}
}

func TestOpenDirectoryWithFaultsKeepsValidClusterFiles(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "EOS_valid")
	brokenPath := filepath.Join(dir, "EOS_broken")
	testfixtures.WriteArchive(t, validPath, "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(brokenPath, []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken cluster file: %v", err)
	}

	entries, faults, err := OpenDirectoryWithFaults(dir)
	if err != nil {
		t.Fatalf("OpenDirectoryWithFaults() error = %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "EOS_valid" {
		t.Fatalf("OpenDirectoryWithFaults() entries = %#v, want valid cluster only", entries)
	}
	if len(faults) != 1 || faults[0].Path != brokenPath || faults[0].Err == nil {
		t.Fatalf("OpenDirectoryWithFaults() faults = %#v, want broken path fault", faults)
	}
}

func TestOpenDirectoryWithFaultsReportsLegacyClusterMetadata(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "EOS_valid")
	legacyPath := filepath.Join(dir, "EOS_legacy")
	testfixtures.WriteArchive(t, validPath, "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(legacyPath, syntheticClusterLegacyArchiveBytes(t), 0o600); err != nil {
		t.Fatalf("write legacy cluster file: %v", err)
	}

	entries, faults, err := OpenDirectoryWithFaults(dir)
	if err != nil {
		t.Fatalf("OpenDirectoryWithFaults() error = %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "EOS_valid" {
		t.Fatalf("OpenDirectoryWithFaults() entries = %#v, want valid cluster only", entries)
	}
	if len(faults) != 1 || faults[0].Path != legacyPath {
		t.Fatalf("OpenDirectoryWithFaults() faults = %#v, want legacy path fault", faults)
	}
	var legacyErr *arkarchive.LegacyArchiveError
	if !errors.As(faults[0].Err, &legacyErr) {
		t.Fatalf("OpenDirectoryWithFaults() fault error = %T %[1]v, want LegacyArchiveError", faults[0].Err)
	}
	if legacyErr.Version != 6 || legacyErr.ObjectCount != 1 || len(legacyErr.ClassNames) != 1 || legacyErr.ClassNames[0] != "/Script/ShooterGame.ArkCloudInventoryData" {
		t.Fatalf("LegacyArchiveError = %#v, want cluster metadata", legacyErr)
	}
}

func TestOpenParsesLocalClusterItemsFromMyArkData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	var item bytes.Buffer
	testfixtures.WriteNameDoubleProperty(&item, "Version", 7)
	testfixtures.WriteNameDoubleProperty(&item, "UploadTime", 12345)
	testfixtures.WriteNameObjectPathProperty(&item, "ItemArchetype", "BlueprintGeneratedClass /Game/Test/PrimalItem_Test.PrimalItem_Test_C")
	testfixtures.WriteNameIntProperty(&item, "ItemQuantity", 3)
	testfixtures.WriteNameFloatProperty(&item, "ItemRating", 7.5)
	testfixtures.WriteNameIntProperty(&item, "ItemQualityIndex", 2)
	testfixtures.WriteNameStringProperty(&item, "CrafterCharacterName", "Survivor")
	testfixtures.WriteNameStringProperty(&item, "CrafterTribeName", "Porters")
	testfixtures.WriteArkString(&item, "None")

	var payload bytes.Buffer
	testfixtures.WriteNameStructArrayProperty(&payload, "ArkItems", "ArkTributeInventoryItem", [][]byte{item.Bytes()})
	testfixtures.WriteArkString(&payload, "None")

	var props bytes.Buffer
	testfixtures.WriteNameStructProperty(&props, "MyArkData", "ArkInventoryData", payload.Bytes())
	testfixtures.WriteArkString(&props, "None")

	testfixtures.WriteArchiveWithProperties(t, path, "/Script/ShooterGame.ArkCloudInventoryData", props.Bytes())

	data, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if len(data.Items) != 1 {
		t.Fatalf("Items length = %d, want 1", len(data.Items))
	}
	itemData := data.Items[0]
	if itemData.Version != 7 || itemData.UploadTime != 12345 {
		t.Fatalf("Item timing = version %.3f upload %.3f, want 7 and 12345", itemData.Version, itemData.UploadTime)
	}
	if itemData.Blueprint != "/Game/Test/PrimalItem_Test.PrimalItem_Test_C" {
		t.Fatalf("Item blueprint = %q", itemData.Blueprint)
	}
	if itemData.Quantity != 3 {
		t.Fatalf("Item quantity = %d, want 3", itemData.Quantity)
	}
	if itemData.Rating != 7.5 || itemData.Quality != 2 {
		t.Fatalf("Item rating/quality = %.2f/%d, want 7.5/2", itemData.Rating, itemData.Quality)
	}
	if itemData.CrafterCharacterName != "Survivor" || itemData.CrafterTribeName != "Porters" {
		t.Fatalf("Item crafter = %q/%q, want Survivor/Porters", itemData.CrafterCharacterName, itemData.CrafterTribeName)
	}
}

func syntheticClusterLegacyArchiveBytes(t *testing.T) []byte {
	t.Helper()

	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	testfixtures.WriteInt32(&buf, 6)
	testfixtures.WriteInt32(&buf, 11)
	testfixtures.WriteInt32(&buf, 22)
	testfixtures.WriteInt32(&buf, 1)
	buf.Write(id[:])
	testfixtures.WriteArkString(&buf, "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteStringArray(&buf, []string{"Legacy_0"})
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, -1)
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, 128)
	testfixtures.WriteUInt32(&buf, 0)
	return buf.Bytes()
}

func TestOpenRecordsLocalClusterDinoArchiveParseErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	var dino bytes.Buffer
	testfixtures.WriteNameDoubleProperty(&dino, "Version", 7)
	testfixtures.WriteNameDoubleProperty(&dino, "UploadTime", 12345)
	testfixtures.WriteNameByteArrayProperty(&dino, "DinoData", []byte("not an archive"))
	testfixtures.WriteArkString(&dino, "None")

	var payload bytes.Buffer
	testfixtures.WriteNameStructArrayProperty(&payload, "ArkTamedDinosData", "ArkTributeDinoData", [][]byte{dino.Bytes()})
	testfixtures.WriteArkString(&payload, "None")

	var props bytes.Buffer
	testfixtures.WriteNameStructProperty(&props, "MyArkData", "ArkInventoryData", payload.Bytes())
	testfixtures.WriteArkString(&props, "None")

	testfixtures.WriteArchiveWithProperties(t, path, "/Script/ShooterGame.ArkCloudInventoryData", props.Bytes())

	data, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if len(data.Dinos) != 1 {
		t.Fatalf("Dinos length = %d, want 1", len(data.Dinos))
	}
	dinoData := data.Dinos[0]
	if dinoData.RawSize != len("not an archive") {
		t.Fatalf("Dino RawSize = %d, want %d", dinoData.RawSize, len("not an archive"))
	}
	if dinoData.Archive != nil {
		t.Fatalf("Dino Archive = %#v, want nil for invalid archive", dinoData.Archive)
	}
	if dinoData.ParseError == "" {
		t.Fatalf("Dino ParseError is empty, want invalid archive parse error")
	}
}
