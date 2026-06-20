package arkcluster

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/internal/safefile"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
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
	writeSparseFile(t, path, 1024)

	_, err := OpenWithOptions(path, Options{MaxFileSize: 16})
	if !errors.Is(err, safefile.ErrFileTooLarge) {
		t.Fatalf("OpenWithOptions() error = %v, want ErrFileTooLarge", err)
	}
}

func TestOpenDirectoryRejectsClusterArchiveAboveConfiguredLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "EOS_abc123")
	writeSparseFile(t, path, 1024)

	_, err := OpenDirectoryWithOptions(dir, Options{MaxFileSize: 16})
	if !errors.Is(err, safefile.ErrFileTooLarge) {
		t.Fatalf("OpenDirectoryWithOptions() error = %v, want ErrFileTooLarge", err)
	}
}

func TestOpenParsesLocalClusterItemsFromMyArkData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	var item bytes.Buffer
	testfixtures.WriteNameDoubleProperty(&item, "Version", 7)
	testfixtures.WriteNameDoubleProperty(&item, "UploadTime", 12345)
	testfixtures.WriteNameObjectPathProperty(&item, "ItemArchetype", "BlueprintGeneratedClass /Game/Test/PrimalItem_Test.PrimalItem_Test_C")
	testfixtures.WriteNameIntProperty(&item, "ItemQuantity", 3)
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
