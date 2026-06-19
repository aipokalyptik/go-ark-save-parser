package arkcluster

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestDiscoverFindsLocalClusterFilesOnly(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	writeArchiveFile(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	writeArchiveFile(t, filepath.Join(dir, "123.arkprofile"), "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")
	writeArchiveFile(t, filepath.Join(dir, "456.arktribe"), "/Script/ShooterGame.PrimalTribeData")
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
	writeArchiveFile(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

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

func writeArchiveFile(t *testing.T, path string, className string) {
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

func writeArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}

func writeStringArray(buf *bytes.Buffer, values []string) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		writeArkString(buf, value)
	}
}
