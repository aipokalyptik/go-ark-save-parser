package arkapi

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestPlayerAPIIndexesLocalProfileAndTribeFiles(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "123.arkprofile")
	tribePath := filepath.Join(dir, "456.arktribe")
	clusterPath := filepath.Join(dir, "EOS_abc123")
	ignoredPath := filepath.Join(dir, "ignore.txt")
	unrelatedExtensionlessPath := filepath.Join(dir, "README")
	nestedDir := filepath.Join(dir, "nested")
	createSyntheticArchive(t, profilePath, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")
	createSyntheticArchive(t, tribePath, "/Script/ShooterGame.PrimalTribeData")
	createSyntheticArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(ignoredPath, []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}
	if err := os.WriteFile(unrelatedExtensionlessPath, []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write unrelated extensionless file: %v", err)
	}
	if err := os.Mkdir(nestedDir, 0o700); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	createSyntheticArchive(t, filepath.Join(nestedDir, "EOS_nested"), "/Script/ShooterGame.ArkCloudInventoryData")

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	if len(api.ProfilePaths()) != 1 || api.ProfilePaths()[0] != profilePath {
		t.Fatalf("ProfilePaths() = %#v, want [%s]", api.ProfilePaths(), profilePath)
	}
	if len(api.TribePaths()) != 1 || api.TribePaths()[0] != tribePath {
		t.Fatalf("TribePaths() = %#v, want [%s]", api.TribePaths(), tribePath)
	}
	if len(api.ClusterPaths()) != 1 || api.ClusterPaths()[0] != clusterPath {
		t.Fatalf("ClusterPaths() = %#v, want [%s]", api.ClusterPaths(), clusterPath)
	}
}

func TestPlayerAPILoadsLocalProfileAndTribeArchives(t *testing.T) {
	dir := t.TempDir()
	createSyntheticArchive(t, filepath.Join(dir, "123.arkprofile"), "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")
	createSyntheticArchive(t, filepath.Join(dir, "456.arktribe"), "/Script/ShooterGame.PrimalTribeData")

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	profiles, err := api.Profiles()
	if err != nil {
		t.Fatalf("Profiles() error = %v", err)
	}
	if len(profiles) != 1 || len(profiles[0].Archive.Objects) != 1 {
		t.Fatalf("Profiles() = %#v", profiles)
	}
	tribes, err := api.Tribes()
	if err != nil {
		t.Fatalf("Tribes() error = %v", err)
	}
	if len(tribes) != 1 || len(tribes[0].Archive.Objects) != 1 {
		t.Fatalf("Tribes() = %#v", tribes)
	}
}

func TestPlayerAPIPlayersParsesLocalProfiles(t *testing.T) {
	dir := t.TempDir()
	writePlayerArchiveFile(t, filepath.Join(dir, "123.arkprofile"))

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	players, err := api.Players()
	if err != nil {
		t.Fatalf("Players() error = %v", err)
	}
	if len(players) != 1 {
		t.Fatalf("Players() length = %d, want 1", len(players))
	}
	if players[0].PlayerDataID != 42 || players[0].CharacterName != "Survivor" || players[0].TribeID != 777 {
		t.Fatalf("Players()[0] = %#v", players[0])
	}
}

func TestPlayerAPILoadsLocalClusterArchives(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	createSyntheticArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	clusters, err := api.Clusters()
	if err != nil {
		t.Fatalf("Clusters() error = %v", err)
	}
	if len(clusters) != 1 || clusters[0].ID != "EOS_abc123" || len(clusters[0].Archive.Objects) != 1 {
		t.Fatalf("Clusters() = %#v", clusters)
	}
}

func writePlayerArchiveFile(t *testing.T, path string) {
	t.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var myData bytes.Buffer
	writeNameIntProperty(&myData, "PlayerDataID", 42)
	writeNameStringProperty(&myData, "PlayerCharacterName", "Survivor")
	writeNameStringProperty(&myData, "PlayerName", "PlatformName")
	writeNameIntProperty(&myData, "TribeID", 777)
	writeArkString(&myData, "None")

	var buf bytes.Buffer
	writeInt32(&buf, 7)
	writeInt32(&buf, 0)
	writeInt32(&buf, 0)
	writeInt32(&buf, 1)
	buf.Write(id[:])
	writeArkString(&buf, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")
	writeUInt32(&buf, 0)
	writeStringArray(&buf, []string{"PlayerData_0"})
	writeUInt32(&buf, 0)
	writeInt32(&buf, -1)
	writeUInt32(&buf, 0)
	offsetPos := buf.Len()
	writeInt32(&buf, 0)
	writeUInt32(&buf, 0)
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	writeNameIntProperty(&buf, "SavedPlayerDataVersion", 17)
	writeNameStructProperty(&buf, "MyData", "PlayerDataStruct", myData.Bytes())
	writeArkString(&buf, "None")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write player archive fixture: %v", err)
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

func writeInt32(buf *bytes.Buffer, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeUInt32(buf *bytes.Buffer, value uint32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}
