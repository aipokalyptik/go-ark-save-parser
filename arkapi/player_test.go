package arkapi

import (
	"os"
	"path/filepath"
	"testing"
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
