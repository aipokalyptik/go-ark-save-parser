package arkcluster

import (
	"bytes"
	"encoding/binary"
	"math"
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

func TestOpenParsesLocalClusterItemsFromMyArkData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	var item bytes.Buffer
	writeDoubleProperty(&item, "Version", 7)
	writeDoubleProperty(&item, "UploadTime", 12345)
	writeObjectPathProperty(&item, "ItemArchetype", "BlueprintGeneratedClass /Game/Test/PrimalItem_Test.PrimalItem_Test_C")
	writeIntProperty(&item, "ItemQuantity", 3)
	writeArkString(&item, "None")

	var payload bytes.Buffer
	writeStructArrayProperty(&payload, "ArkItems", "ArkTributeInventoryItem", [][]byte{item.Bytes()})
	writeArkString(&payload, "None")

	var props bytes.Buffer
	writeStructProperty(&props, "MyArkData", "ArkInventoryData", payload.Bytes())
	writeArkString(&props, "None")

	writeArchiveFileWithProperties(t, path, "/Script/ShooterGame.ArkCloudInventoryData", props.Bytes())

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

func writeArchiveFile(t *testing.T, path string, className string) {
	t.Helper()
	writeArchiveFileWithProperties(t, path, className, nil)
}

func writeArchiveFileWithProperties(t *testing.T, path string, className string, props []byte) {
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
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	if len(props) > 0 {
		propertiesOffset := int32(buf.Len() - 1)
		binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
		buf.Write(props)
	} else {
		binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(128))
	}
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

func writeDoubleProperty(buf *bytes.Buffer, name string, value float64) {
	writeArkString(buf, name)
	writeArkString(buf, "DoubleProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, math.Float64bits(value))
}

func writeIntProperty(buf *bytes.Buffer, name string, value int32) {
	writeArkString(buf, name)
	writeArkString(buf, "IntProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeObjectPathProperty(buf *bytes.Buffer, name string, value string) {
	var body bytes.Buffer
	_ = binary.Write(&body, binary.LittleEndian, int32(1))
	writeArkString(&body, value)

	writeArkString(buf, name)
	writeArkString(buf, "ObjectProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(body.Len()))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	buf.Write(body.Bytes())
}

func writeStructProperty(buf *bytes.Buffer, name string, structType string, body []byte) {
	writeArkString(buf, name)
	writeArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	writeArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	writeArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
}

func writeStructArrayProperty(buf *bytes.Buffer, name string, structType string, elements [][]byte) {
	bodySize := 4
	for _, element := range elements {
		bodySize += len(element)
	}
	writeArkString(buf, name)
	writeArkString(buf, "ArrayProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	writeArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	writeArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	writeArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(bodySize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(elements)))
	for _, element := range elements {
		buf.Write(element)
	}
}
