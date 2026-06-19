package arkapi

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
)

func TestPlayerAPIIndexesLocalProfileAndTribeFiles(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "123.arkprofile")
	tribePath := filepath.Join(dir, "456.arktribe")
	clusterPath := filepath.Join(dir, "EOS_abc123")
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	ignoredPath := filepath.Join(dir, "ignore.txt")
	unrelatedExtensionlessPath := filepath.Join(dir, "README")
	nestedDir := filepath.Join(dir, "nested")
	testfixtures.WriteArchive(t, profilePath, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")
	testfixtures.WriteArchive(t, tribePath, "/Script/ShooterGame.PrimalTribeData")
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	writeTributeFile(t, tributePath, []uint64{11}, []uint64{22})
	if err := os.WriteFile(ignoredPath, []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write ignored file: %v", err)
	}
	if err := os.WriteFile(unrelatedExtensionlessPath, []byte("ignore"), 0o600); err != nil {
		t.Fatalf("write unrelated extensionless file: %v", err)
	}
	if err := os.Mkdir(nestedDir, 0o700); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	testfixtures.WriteArchive(t, filepath.Join(nestedDir, "EOS_nested"), "/Script/ShooterGame.ArkCloudInventoryData")

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
	if len(api.TributePaths()) != 1 || api.TributePaths()[0] != tributePath {
		t.Fatalf("TributePaths() = %#v, want [%s]", api.TributePaths(), tributePath)
	}
}

func TestPlayerAPILoadsLocalProfileAndTribeArchives(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteArchive(t, filepath.Join(dir, "123.arkprofile"), "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")
	testfixtures.WriteArchive(t, filepath.Join(dir, "456.arktribe"), "/Script/ShooterGame.PrimalTribeData")

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
	testfixtures.WritePlayerArchive(t, filepath.Join(dir, "123.arkprofile"))

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

func TestPlayerAPIFindsLocalPlayersByDataAndTribeID(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchive(t, filepath.Join(dir, "123.arkprofile"))

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	player, ok, err := api.PlayerByDataID(42)
	if err != nil {
		t.Fatalf("PlayerByDataID() error = %v", err)
	}
	if !ok || player.CharacterName != "Survivor" {
		t.Fatalf("PlayerByDataID() = %#v, %v; want Survivor, true", player, ok)
	}
	if _, ok, err := api.PlayerByDataID(999); err != nil || ok {
		t.Fatalf("PlayerByDataID(missing) = ok %v err %v, want false nil", ok, err)
	}
	players, err := api.PlayersByTribeID(777)
	if err != nil {
		t.Fatalf("PlayersByTribeID() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 {
		t.Fatalf("PlayersByTribeID() = %#v, want player 42", players)
	}
}

func TestPlayerAPITribeSummariesParsesLocalTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchive(t, filepath.Join(dir, "456.arktribe"))

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	tribes, err := api.TribeSummaries()
	if err != nil {
		t.Fatalf("TribeSummaries() error = %v", err)
	}
	if len(tribes) != 1 {
		t.Fatalf("TribeSummaries() length = %d, want 1", len(tribes))
	}
	if tribes[0].Name != "Porters" || tribes[0].TribeID != 12345 {
		t.Fatalf("TribeSummaries()[0] = %#v", tribes[0])
	}
}

func TestPlayerAPITribeDetailsParsesAndFindsLocalTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchive(t, filepath.Join(dir, "456.arktribe"))

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	tribes, err := api.TribeDetails()
	if err != nil {
		t.Fatalf("TribeDetails() error = %v", err)
	}
	if len(tribes) != 1 || tribes[0].Name != "Porters" || tribes[0].OwnerID != 42 || tribes[0].NumDinos != 7 {
		t.Fatalf("TribeDetails() = %#v, want parsed Porters tribe", tribes)
	}
	tribe, ok, err := api.TribeByID(12345)
	if err != nil {
		t.Fatalf("TribeByID() error = %v", err)
	}
	if !ok || tribe.Name != "Porters" {
		t.Fatalf("TribeByID() = %#v, %v; want Porters, true", tribe, ok)
	}
	if _, ok, err := api.TribeByID(999); err != nil || ok {
		t.Fatalf("TribeByID(missing) = ok %v err %v, want false nil", ok, err)
	}
}

func TestPlayerAPILoadsLocalClusterArchives(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")

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

func TestPlayerAPILoadsLocalTributeIndexes(t *testing.T) {
	dir := t.TempDir()
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	writeTributeFile(t, tributePath, []uint64{11, 22}, []uint64{33})

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	tributes, err := api.Tributes()
	if err != nil {
		t.Fatalf("Tributes() error = %v", err)
	}
	if len(tributes) != 1 || tributes[0].ID != "abc" || len(tributes[0].PlayerDataIDs) != 2 || len(tributes[0].TribeDataIDs) != 1 {
		t.Fatalf("Tributes() = %#v", tributes)
	}
}

func writeTributeFile(t *testing.T, path string, playerIDs []uint64, tribeIDs []uint64) {
	t.Helper()
	var buf bytes.Buffer
	writeTributeIDs(t, &buf, playerIDs)
	writeTributeIDs(t, &buf, tribeIDs)
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write tribute fixture: %v", err)
	}
}

func writeTributeIDs(t *testing.T, buf *bytes.Buffer, ids []uint64) {
	t.Helper()
	if err := binary.Write(buf, binary.LittleEndian, int32(len(ids))); err != nil {
		t.Fatalf("write tribute count: %v", err)
	}
	for _, id := range ids {
		if err := binary.Write(buf, binary.LittleEndian, id); err != nil {
			t.Fatalf("write tribute id: %v", err)
		}
	}
}
