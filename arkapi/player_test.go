package arkapi

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
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
	testfixtures.WriteTributeFile(t, tributePath, []uint64{11}, []uint64{22})
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
	summary := api.LocalFileSummary()
	wantSummary := LocalFileSummary{Profiles: 1, Tribes: 1, Clusters: 1, Tributes: 1}
	if summary != wantSummary {
		t.Fatalf("LocalFileSummary() = %#v, want %#v", summary, wantSummary)
	}
}

func TestNewPlayerFromPathOpensDirectory(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchive(t, filepath.Join(dir, "123.arkprofile"))

	api, closeAPI, err := NewPlayerFromPath(dir, PlayerPathOptions{})
	if err != nil {
		t.Fatalf("NewPlayerFromPath(directory) error = %v", err)
	}
	defer closeAPI()

	players, err := api.Players()
	if err != nil {
		t.Fatalf("Players() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 {
		t.Fatalf("Players() = %#v, want directory profile player", players)
	}
}

func TestNewPlayerFromPathOpensSave(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	path := save.Path()
	if err := save.Close(); err != nil {
		t.Fatalf("Close(save) error = %v", err)
	}

	api, closeAPI, err := NewPlayerFromPath(path, PlayerPathOptions{})
	if err != nil {
		t.Fatalf("NewPlayerFromPath(save) error = %v", err)
	}
	defer closeAPI()

	players, err := api.Players()
	if err != nil {
		t.Fatalf("Players() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 {
		t.Fatalf("Players() = %#v, want save-contained player", players)
	}
}

func TestNewPlayerFromPathFallsBackToDirectoryProfilesWhenSaveHasNoPlayers(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "empty.ark")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{Header: testfixtures.Header("Valguero_WP", map[uint32]string{})})
	testfixtures.WritePlayerArchive(t, filepath.Join(dir, "123.arkprofile"))

	api, closeAPI, err := NewPlayerFromPath(savePath, PlayerPathOptions{Fallback: PlayerPathFallbackPlayers})
	if err != nil {
		t.Fatalf("NewPlayerFromPath(fallback players) error = %v", err)
	}
	defer closeAPI()

	players, err := api.Players()
	if err != nil {
		t.Fatalf("Players() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 {
		t.Fatalf("Players() = %#v, want directory fallback player", players)
	}
	if len(api.ProfilePaths()) != 1 {
		t.Fatalf("ProfilePaths() = %#v, want directory fallback API", api.ProfilePaths())
	}
}

func TestNewPlayerFromPathFallsBackToDirectoryTribesWhenSaveHasNoTribes(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "empty.ark")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{Header: testfixtures.Header("Valguero_WP", map[uint32]string{})})
	testfixtures.WriteTribeArchive(t, filepath.Join(dir, "456.arktribe"))

	api, closeAPI, err := NewPlayerFromPath(savePath, PlayerPathOptions{Fallback: PlayerPathFallbackTribes})
	if err != nil {
		t.Fatalf("NewPlayerFromPath(fallback tribes) error = %v", err)
	}
	defer closeAPI()

	tribes, err := api.TribeDetails()
	if err != nil {
		t.Fatalf("TribeDetails() error = %v", err)
	}
	if len(tribes) != 1 || tribes[0].TribeID != 12345 {
		t.Fatalf("TribeDetails() = %#v, want directory fallback tribe", tribes)
	}
	if len(api.TribePaths()) != 1 {
		t.Fatalf("TribePaths() = %#v, want directory fallback API", api.TribePaths())
	}
}

func TestNewPlayerFromPathDoesNotFallbackWhenSavePlayersHaveFaults(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "faulty-players.ark")
	faultyID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			faultyID: testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{
				PlayerDataID:  43,
				CharacterName: "Broken",
				PlayerName:    "BrokenPlatform",
				TribeID:       12345,
			})[:40],
		},
	})
	testfixtures.WritePlayerArchive(t, filepath.Join(dir, "123.arkprofile"))

	api, closeAPI, err := NewPlayerFromPath(savePath, PlayerPathOptions{Fallback: PlayerPathFallbackPlayers})
	if err == nil {
		defer closeAPI()
		t.Fatalf("NewPlayerFromPath(faulty player fallback) error = nil with api %#v, want save parse fault", api)
	}
}

func TestNewPlayerFromPathDoesNotFallbackWhenSaveTribesHaveFaults(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "faulty-tribes.ark")
	faultyID := uuid.MustParse("33112233-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			faultyID: testfixtures.TribeGameObjectBytes(testfixtures.TribeArchiveOptions{
				Name:     "Broken",
				TribeID:  67890,
				NumDinos: 3,
			})[:40],
		},
	})
	testfixtures.WriteTribeArchive(t, filepath.Join(dir, "456.arktribe"))

	api, closeAPI, err := NewPlayerFromPath(savePath, PlayerPathOptions{Fallback: PlayerPathFallbackTribes})
	if err == nil {
		defer closeAPI()
		t.Fatalf("NewPlayerFromPath(faulty tribe fallback) error = nil with api %#v, want save parse fault", api)
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

func TestPlayerAPIParsesSaveContainedPlayersAndTribes(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	api := NewPlayer(save)
	players, err := api.Players()
	if err != nil {
		t.Fatalf("Players() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 || players[0].Level != 5 || players[0].Experience != 123.5 {
		t.Fatalf("Players() = %#v, want save-contained player 42", players)
	}
	tribes, err := api.TribeDetails()
	if err != nil {
		t.Fatalf("TribeDetails() error = %v", err)
	}
	if len(tribes) != 1 || tribes[0].TribeID != 12345 || tribes[0].Name != "Porters" {
		t.Fatalf("TribeDetails() = %#v, want save-contained Porters tribe", tribes)
	}
	tribePlayers, err := api.TribePlayerMap()
	if err != nil {
		t.Fatalf("TribePlayerMap() error = %v", err)
	}
	if len(tribePlayers[12345]) != 1 || tribePlayers[12345][0].PlayerDataID != 42 {
		t.Fatalf("TribePlayerMap()[12345] = %#v, want save-contained player 42", tribePlayers[12345])
	}
}

func TestPlayerAPIParsesGameModeCustomBytesPlayersAndTribes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "embedded.ark")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Custom: map[string][]byte{
			"GameModeCustomBytes": gameModeCustomBytesFixture(t),
		},
		EmptyTables: true,
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open(save) error = %v", err)
	}
	defer save.Close()

	api := NewPlayer(save)
	players, err := api.Players()
	if err != nil {
		t.Fatalf("Players() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 || players[0].CharacterName != "Survivor" {
		t.Fatalf("Players() = %#v, want embedded player 42", players)
	}
	tribes, err := api.TribeDetails()
	if err != nil {
		t.Fatalf("TribeDetails() error = %v", err)
	}
	if len(tribes) != 1 || tribes[0].TribeID != 12345 || tribes[0].Name != "Porters" {
		t.Fatalf("TribeDetails() = %#v, want embedded Porters tribe", tribes)
	}
	tribePlayers, err := api.TribePlayerMap()
	if err != nil {
		t.Fatalf("TribePlayerMap() error = %v", err)
	}
	if len(tribePlayers[12345]) != 1 || tribePlayers[12345][0].PlayerDataID != 42 {
		t.Fatalf("TribePlayerMap()[12345] = %#v, want embedded player 42", tribePlayers[12345])
	}
}

func TestPlayerAPISaveContainedLookupAndOwnerHelpers(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	api := NewPlayer(save)
	player, ok, err := api.PlayerByDataID(42)
	if err != nil {
		t.Fatalf("PlayerByDataID() error = %v", err)
	}
	if !ok || player.UniqueID != "eos-survivor" {
		t.Fatalf("PlayerByDataID() = %#v, %v; want eos-survivor", player, ok)
	}
	if _, ok, err := api.PlayerByUniqueID("missing"); err != nil || ok {
		t.Fatalf("PlayerByUniqueID(missing) = ok %v err %v, want false nil", ok, err)
	}
	tribe, ok, err := api.TribeByID(12345)
	if err != nil {
		t.Fatalf("TribeByID() error = %v", err)
	}
	if !ok || tribe.Name != "Porters" {
		t.Fatalf("TribeByID() = %#v, %v; want Porters", tribe, ok)
	}
	owner, ok, err := api.ObjectOwnerByPlayerDataID(42)
	if err != nil {
		t.Fatalf("ObjectOwnerByPlayerDataID() error = %v", err)
	}
	if !ok || owner.PlayerID != 42 || owner.PlayerName != "PlatformName" || owner.TribeName != "Porters" {
		t.Fatalf("ObjectOwnerByPlayerDataID() = %#v, %v; want save-contained owner", owner, ok)
	}
	dinoOwner, ok, err := api.DinoOwnerByPlayerDataID(42)
	if err != nil {
		t.Fatalf("DinoOwnerByPlayerDataID() error = %v", err)
	}
	if !ok || dinoOwner.PlayerID != 42 || dinoOwner.ImprinterName != "Survivor" || dinoOwner.TamerString != "Porters" {
		t.Fatalf("DinoOwnerByPlayerDataID() = %#v, %v; want save-contained owner", dinoOwner, ok)
	}
}

func TestPlayerAPIPlayersWithFaultsKeepsValidSavePlayersAndReportsFaults(t *testing.T) {
	validID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	path := filepath.Join(t.TempDir(), "players.ark")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			validID: testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{
				PlayerDataID:  42,
				CharacterName: "Survivor",
				PlayerName:    "PlatformName",
				UniqueID:      "eos-survivor",
				TribeID:       12345,
			}),
			faultyID: testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{
				PlayerDataID:  43,
				CharacterName: "Broken",
				PlayerName:    "BrokenPlatform",
				TribeID:       12345,
			})[:40],
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	players, faults, err := NewPlayer(save).PlayersWithFaults()
	if err != nil {
		t.Fatalf("PlayersWithFaults() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 {
		t.Fatalf("PlayersWithFaults() players = %#v, want player 42", players)
	}
	if len(faults) != 1 || faults[0].UUID != faultyID || faults[0].Err == nil {
		t.Fatalf("PlayersWithFaults() faults = %#v, want one player parse fault", faults)
	}
}

func TestPlayerAPIPlayersWithFaultsKeepsValidLocalProfilesAndReportsFaults(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "123.arkprofile")
	brokenPath := filepath.Join(dir, "broken.arkprofile")
	testfixtures.WritePlayerArchiveWithOptions(t, validPath, testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
	})
	if err := os.WriteFile(brokenPath, []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken profile: %v", err)
	}

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	players, faults, err := api.PlayersWithFaults()
	if err != nil {
		t.Fatalf("PlayersWithFaults() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 {
		t.Fatalf("PlayersWithFaults() players = %#v, want valid local profile player", players)
	}
	if len(faults) != 1 || faults[0].ClassName != brokenPath || faults[0].Err == nil {
		t.Fatalf("PlayersWithFaults() faults = %#v, want broken profile path fault", faults)
	}
}

func TestPlayerAPIPlayersWithFaultsReportsLegacyProfileMetadataFaults(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "123.arkprofile")
	legacyPath := filepath.Join(dir, "legacy.arkprofile")
	testfixtures.WritePlayerArchiveWithOptions(t, validPath, testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
	})
	if err := os.WriteFile(legacyPath, syntheticLegacyArchiveBytes(t, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C"), 0o600); err != nil {
		t.Fatalf("write legacy profile: %v", err)
	}

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	players, faults, err := api.PlayersWithFaults()
	if err != nil {
		t.Fatalf("PlayersWithFaults() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 {
		t.Fatalf("PlayersWithFaults() players = %#v, want valid local profile player", players)
	}
	if len(faults) != 1 || faults[0].ClassName != legacyPath {
		t.Fatalf("PlayersWithFaults() faults = %#v, want legacy profile path fault", faults)
	}
	var legacyErr *arkarchive.LegacyArchiveError
	if !errors.As(faults[0].Err, &legacyErr) {
		t.Fatalf("PlayersWithFaults() fault error = %T %[1]v, want LegacyArchiveError", faults[0].Err)
	}
	if legacyErr.Version != 6 || legacyErr.ObjectCount != 1 || len(legacyErr.ClassNames) != 1 {
		t.Fatalf("LegacyArchiveError = %#v, want version/object/class metadata", legacyErr)
	}
}

func TestPlayerAPITribeDetailsWithFaultsKeepsValidSaveTribesAndReportsFaults(t *testing.T) {
	validID := uuid.MustParse("22112233-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("33112233-4455-6677-8899-aabbccddeeff")
	path := filepath.Join(t.TempDir(), "tribes.ark")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			validID: testfixtures.TribeGameObjectBytes(testfixtures.TribeArchiveOptions{
				Name:     "Porters",
				TribeID:  12345,
				NumDinos: 7,
			}),
			faultyID: testfixtures.TribeGameObjectBytes(testfixtures.TribeArchiveOptions{
				Name:     "Broken",
				TribeID:  67890,
				NumDinos: 3,
			})[:40],
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	tribes, faults, err := NewPlayer(save).TribeDetailsWithFaults()
	if err != nil {
		t.Fatalf("TribeDetailsWithFaults() error = %v", err)
	}
	if len(tribes) != 1 || tribes[0].TribeID != 12345 {
		t.Fatalf("TribeDetailsWithFaults() tribes = %#v, want tribe 12345", tribes)
	}
	if len(faults) != 1 || faults[0].UUID != faultyID || faults[0].Err == nil {
		t.Fatalf("TribeDetailsWithFaults() faults = %#v, want one tribe parse fault", faults)
	}
}

func TestPlayerAPITribeDetailsWithFaultsKeepsValidLocalTribesAndReportsFaults(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "456.arktribe")
	brokenPath := filepath.Join(dir, "broken.arktribe")
	testfixtures.WriteTribeArchiveWithOptions(t, validPath, testfixtures.TribeArchiveOptions{
		Name:     "Porters",
		TribeID:  12345,
		OwnerID:  42,
		NumDinos: 7,
	})
	if err := os.WriteFile(brokenPath, []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken tribe: %v", err)
	}

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	tribes, faults, err := api.TribeDetailsWithFaults()
	if err != nil {
		t.Fatalf("TribeDetailsWithFaults() error = %v", err)
	}
	if len(tribes) != 1 || tribes[0].TribeID != 12345 {
		t.Fatalf("TribeDetailsWithFaults() tribes = %#v, want valid local tribe", tribes)
	}
	if len(faults) != 1 || faults[0].ClassName != brokenPath || faults[0].Err == nil {
		t.Fatalf("TribeDetailsWithFaults() faults = %#v, want broken tribe path fault", faults)
	}
}

func TestPlayerAPIPlayerPawnByDataID(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	api := NewPlayer(save)
	pawn, ok, err := api.PlayerPawnByDataID(42)
	if err != nil {
		t.Fatalf("PlayerPawnByDataID() error = %v", err)
	}
	if !ok || pawn.UUID != uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff") {
		t.Fatalf("PlayerPawnByDataID() = %#v, %v; want synthetic pawn", pawn, ok)
	}
	if _, ok, err := api.PlayerPawnByDataID(999); err != nil || ok {
		t.Fatalf("PlayerPawnByDataID(missing) = ok %v err %v, want false nil", ok, err)
	}

	directoryAPI, err := NewPlayerFromDirectory(t.TempDir())
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	if _, ok, err := directoryAPI.PlayerPawnByDataID(42); err != nil || ok {
		t.Fatalf("directory PlayerPawnByDataID() = ok %v err %v, want false nil", ok, err)
	}
}

func TestPlayerAPIPlayerInventoryByDataID(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	api := NewPlayer(save)
	inventory, ok, err := api.PlayerInventoryByDataID(42)
	if err != nil {
		t.Fatalf("PlayerInventoryByDataID() error = %v", err)
	}
	if !ok || inventory.UUID != uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff") {
		t.Fatalf("PlayerInventoryByDataID() = %#v, %v; want synthetic inventory", inventory, ok)
	}
	if inventory.NumberOfItems() != 2 {
		t.Fatalf("PlayerInventoryByDataID().NumberOfItems() = %d, want 2", inventory.NumberOfItems())
	}
	if _, ok, err := api.PlayerInventoryByDataID(999); err != nil || ok {
		t.Fatalf("PlayerInventoryByDataID(missing) = ok %v err %v, want false nil", ok, err)
	}
}

func TestPlayerInventoryLookupFromPathFindsInventoryAndLocation(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	lookup, err := PlayerInventoryLookupFromPath(save.Path(), 42)
	if err != nil {
		t.Fatalf("PlayerInventoryLookupFromPath() error = %v", err)
	}
	if !lookup.Found || lookup.PlayerDataID != 42 {
		t.Fatalf("PlayerInventoryLookupFromPath() = %#v, want found player 42", lookup)
	}
	if lookup.InventoryUUID != uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff") || lookup.Items != 2 {
		t.Fatalf("PlayerInventoryLookupFromPath() inventory = %#v, want inventory with two items", lookup)
	}
	if !lookup.HasLocation || lookup.Location == nil || lookup.Location.X != 11 || lookup.Location.Y != 22 || lookup.Location.Z != 33 {
		t.Fatalf("PlayerInventoryLookupFromPath() location = %#v, has %v", lookup.Location, lookup.HasLocation)
	}

	missing, err := PlayerInventoryLookupFromPath(save.Path(), 999)
	if err != nil {
		t.Fatalf("PlayerInventoryLookupFromPath(missing) error = %v", err)
	}
	if missing.Found || missing.PlayerDataID != 999 || missing.Items != 0 || missing.HasLocation {
		t.Fatalf("PlayerInventoryLookupFromPath(missing) = %#v, want missing lookup", missing)
	}
}

func TestPlayerAPIPlayerInventoriesWithFaultsIndexesSavePawns(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	api := NewPlayer(save)
	inventories, faults, err := api.PlayerInventoriesWithFaults()
	if err != nil {
		t.Fatalf("PlayerInventoriesWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("PlayerInventoriesWithFaults() faults = %#v, want none", faults)
	}
	inventory, ok := inventories[42]
	if !ok {
		t.Fatalf("PlayerInventoriesWithFaults() missing player 42 inventory: %#v", inventories)
	}
	if inventory.UUID != uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff") {
		t.Fatalf("PlayerInventoriesWithFaults()[42].UUID = %s", inventory.UUID)
	}
	if inventory.NumberOfItems() != 2 {
		t.Fatalf("Inventory.NumberOfItems() = %d, want referenced item count 2", inventory.NumberOfItems())
	}
	if got := api.InventoryItemCount(inventory); got != 2 {
		t.Fatalf("InventoryItemCount() = %d, want upstream-style referenced item count 2", got)
	}
}

func TestPlayerAPIPlayerInventoriesWithFaultsKeepsMissingInventoryFault(t *testing.T) {
	pawnID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	missingInventoryID := uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff")
	path := filepath.Join(t.TempDir(), "missing-inventory.ark")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			pawnID: testfixtures.PlayerPawnGameObjectBytes(42, missingInventoryID),
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	inventories, faults, err := NewPlayer(save).PlayerInventoriesWithFaults()
	if err != nil {
		t.Fatalf("PlayerInventoriesWithFaults() error = %v", err)
	}
	if len(inventories) != 0 {
		t.Fatalf("PlayerInventoriesWithFaults() inventories = %#v, want empty on missing inventory row", inventories)
	}
	if len(faults) != 1 || faults[0].UUID != missingInventoryID || faults[0].Err == nil {
		t.Fatalf("PlayerInventoriesWithFaults() faults = %#v, want missing inventory fault", faults)
	}
}

func TestPlayerAPIPlayerInventorySummaryForPlayers(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	api := NewPlayer(save)
	players, err := api.Players()
	if err != nil {
		t.Fatalf("Players() error = %v", err)
	}
	summary, faults, err := api.PlayerInventorySummaryForPlayers(players)
	if err != nil {
		t.Fatalf("PlayerInventorySummaryForPlayers() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("PlayerInventorySummaryForPlayers() faults = %#v, want none", faults)
	}
	want := PlayerInventorySummary{
		Players:       1,
		WithInventory: 1,
		TotalItems:    2,
		MaxItems:      2,
		MinItems:      2,
		AverageItems:  2,
	}
	if summary != want {
		t.Fatalf("PlayerInventorySummaryForPlayers() = %#v, want %#v", summary, want)
	}
}

func TestPlayerAPIPlayerInventorySummaryForPlayersEmpty(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	summary, faults, err := NewPlayer(save).PlayerInventorySummaryForPlayers(nil)
	if err != nil {
		t.Fatalf("PlayerInventorySummaryForPlayers(nil) error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("PlayerInventorySummaryForPlayers(nil) faults = %#v, want none", faults)
	}
	if summary != (PlayerInventorySummary{}) {
		t.Fatalf("PlayerInventorySummaryForPlayers(nil) = %#v, want zero summary", summary)
	}
}

func TestPlayerAPIPlayerInventorySummaryForPlayersRequiresSave(t *testing.T) {
	api, err := NewPlayerFromDirectory(t.TempDir())
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	_, _, err = api.PlayerInventorySummaryForPlayers(nil)
	if err == nil || !strings.Contains(err.Error(), "save-backed") {
		t.Fatalf("PlayerInventorySummaryForPlayers(directory) error = %v, want save-backed error", err)
	}
}

func TestPlayerAPIPlayerRosterSummaryForPlayers(t *testing.T) {
	api := NewPlayer(nil)
	players := []arkobject.Player{
		{CharacterName: "Ada", Level: 5},
		{PlayerName: "Grace", Level: 12},
		{Level: 7},
	}

	summary := api.PlayerRosterSummaryForPlayers(players)
	want := PlayerRosterSummary{Players: 3, WithNames: 2, HighestLevel: 12}
	if summary != want {
		t.Fatalf("PlayerRosterSummaryForPlayers() = %#v, want %#v", summary, want)
	}
}

func TestPlayerAPIPlayerRosterSummaryWithFaultsKeepsValidLocalProfiles(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		ExtraCharacterLevel: 4,
	})
	if err := os.WriteFile(filepath.Join(dir, "broken.arkprofile"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken profile: %v", err)
	}

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	summary, faults, err := api.PlayerRosterSummaryWithFaults()
	if err != nil {
		t.Fatalf("PlayerRosterSummaryWithFaults() error = %v", err)
	}
	want := PlayerRosterSummary{Players: 1, WithNames: 1, HighestLevel: 5}
	if summary != want {
		t.Fatalf("PlayerRosterSummaryWithFaults() summary = %#v, want %#v", summary, want)
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("PlayerRosterSummaryWithFaults() faults = %#v, want one profile fault", faults)
	}
}

func TestPlayerRosterSummaryFromPathUsesDirectoryProfiles(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		ExtraCharacterLevel: 9,
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        43,
		ExtraCharacterLevel: 4,
	})

	summary, faults, err := PlayerRosterSummaryFromPath(dir)
	if err != nil {
		t.Fatalf("PlayerRosterSummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("PlayerRosterSummaryFromPath() faults = %#v, want none", faults)
	}
	want := PlayerRosterSummary{Players: 2, WithNames: 1, HighestLevel: 10}
	if summary != want {
		t.Fatalf("PlayerRosterSummaryFromPath() = %#v, want %#v", summary, want)
	}
}

func TestPlayersFromPathReturnsTypedPlayersAndFaults(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		ExtraCharacterLevel: 9,
	})
	if err := os.WriteFile(filepath.Join(dir, "broken.arkprofile"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken profile: %v", err)
	}

	players, faults, err := PlayersFromPath(dir, PlayerPathOptions{})
	if err != nil {
		t.Fatalf("PlayersFromPath() error = %v", err)
	}
	if len(players) != 1 || players[0].PlayerDataID != 42 || players[0].Level != 10 || !players[0].HasName() {
		t.Fatalf("PlayersFromPath() players = %#v, want typed local player", players)
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("PlayersFromPath() faults = %#v, want one local profile fault", faults)
	}
}

func TestPlayerAPIPlayerAllSummaryForData(t *testing.T) {
	api := NewPlayer(nil)
	players := []arkobject.Player{
		{
			CharacterName:   "Ada",
			Level:           5,
			NumDeaths:       2,
			UnlockedEngrams: []string{"EngramA", "EngramB", "EngramB"},
		},
		{
			PlayerName:      "Grace",
			Level:           12,
			NumDeaths:       3,
			UnlockedEngrams: []string{"EngramD", "EngramA"},
		},
	}
	tribes := []arkobject.Tribe{
		{Name: "Porters"},
		{},
	}

	summary := api.PlayerAllSummaryForData(players, tribes)
	want := PlayerAllSummary{
		Players:         2,
		Tribes:          2,
		HighestLevel:    12,
		TotalDeaths:     5,
		UnlockedEngrams: 3,
	}
	if summary != want {
		t.Fatalf("PlayerAllSummaryForData() = %#v, want %#v", summary, want)
	}
}

func TestPlayerAPIPlayerAllSummaryWithFaultsKeepsValidLocalPlayersAndTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:      42,
		CharacterName:     "Survivor",
		NumDeaths:         3,
		UnlockedEngrams:   []string{"EngramA", "EngramB"},
		ExperiencePoints:  10,
		TotalEngramPoints: 5,
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:    "Porters",
		TribeID: 12345,
	})
	if err := os.WriteFile(filepath.Join(dir, "broken.arkprofile"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken profile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.arktribe"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken tribe: %v", err)
	}

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	summary, faults, err := api.PlayerAllSummaryWithFaults()
	if err != nil {
		t.Fatalf("PlayerAllSummaryWithFaults() error = %v", err)
	}
	want := PlayerAllSummary{Players: 1, Tribes: 1, HighestLevel: 1, TotalDeaths: 3, UnlockedEngrams: 2}
	if summary != want {
		t.Fatalf("PlayerAllSummaryWithFaults() summary = %#v, want %#v", summary, want)
	}
	if len(faults) != 2 {
		t.Fatalf("PlayerAllSummaryWithFaults() faults = %#v, want two local faults", faults)
	}
}

func TestPlayerAllSummaryFromPathUsesDirectoryPlayersAndTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		ExtraCharacterLevel: 9,
		NumDeaths:           3,
		UnlockedEngrams:     []string{"EngramA", "EngramB"},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:     "Porters",
		TribeID:  12345,
		OwnerID:  42,
		NumDinos: 7,
	})

	summary, faults, err := PlayerAllSummaryFromPath(dir)
	if err != nil {
		t.Fatalf("PlayerAllSummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("PlayerAllSummaryFromPath() faults = %#v, want none", faults)
	}
	want := PlayerAllSummary{Players: 1, Tribes: 1, HighestLevel: 10, TotalDeaths: 3, UnlockedEngrams: 2}
	if summary != want {
		t.Fatalf("PlayerAllSummaryFromPath() = %#v, want %#v", summary, want)
	}
}

func TestPlayerUnlockedEngramsFromPathUsesDirectoryProfiles(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID: 42,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID: 43,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
			"Blueprint'/Game/Engrams/EngramC.EngramC_C'",
		},
	})

	engrams, err := PlayerUnlockedEngramsFromPath(dir)
	if err != nil {
		t.Fatalf("PlayerUnlockedEngramsFromPath() error = %v", err)
	}
	want := []string{
		"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
		"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
		"Blueprint'/Game/Engrams/EngramC.EngramC_C'",
	}
	if !reflect.DeepEqual(engrams, want) {
		t.Fatalf("PlayerUnlockedEngramsFromPath() = %#v, want %#v", engrams, want)
	}
}

func TestLocalProfileSummaryFromPathUsesDirectoryProfilesAndTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		TribeID:             12345,
		ExtraCharacterLevel: 9,
		ExperiencePoints:    123.5,
		NumDeaths:           4,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
		},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:     "Porters",
		TribeID:  12345,
		OwnerID:  42,
		NumDinos: 7,
	})
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_abc123"), "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "abc.arktributetribe"), []uint64{11}, []uint64{22})

	summary, faults, err := LocalProfileSummaryFromPath(dir)
	if err != nil {
		t.Fatalf("LocalProfileSummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("LocalProfileSummaryFromPath() faults = %#v, want none", faults)
	}
	if summary.Files != (LocalFileSummary{Profiles: 1, Tribes: 1, Clusters: 1, Tributes: 1}) {
		t.Fatalf("LocalProfileSummaryFromPath() files = %#v, want one of each local file type", summary.Files)
	}
	if !summary.HasParsedPlayers || !summary.HasParsedTribes || !summary.HasTribePlayerLinks || !summary.HasTotalDeaths {
		t.Fatalf("LocalProfileSummaryFromPath() count success flags missing: %#v", summary)
	}
	if summary.ParsedPlayers != 1 || summary.ParsedTribes != 1 || summary.TribePlayerLinks != 0 || summary.TotalDeaths != 4 {
		t.Fatalf("LocalProfileSummaryFromPath() counts = %#v, want parsed player/tribe, zero link rows, and four deaths", summary)
	}
	if !summary.HasHighestLevel || summary.HighestLevel != 10 {
		t.Fatalf("LocalProfileSummaryFromPath() highest level = %d, %v; want 10, true", summary.HighestLevel, summary.HasHighestLevel)
	}
	if !summary.HasHighestExperience || summary.HighestExperience != 123.5 {
		t.Fatalf("LocalProfileSummaryFromPath() highest experience = %f, %v; want 123.5, true", summary.HighestExperience, summary.HasHighestExperience)
	}
	if !summary.HasAverageDeaths || !summary.HasAverageLevel || !summary.HasAverageExperience {
		t.Fatalf("LocalProfileSummaryFromPath() averages missing: %#v", summary)
	}
	if !summary.HasUnlockedEngrams {
		t.Fatalf("LocalProfileSummaryFromPath() unlocked engram success flag missing: %#v", summary)
	}
	if summary.AverageDeaths != 4 || summary.AverageLevel != 10 || summary.AverageExperience != 123.5 || summary.UnlockedEngrams != 2 {
		t.Fatalf("LocalProfileSummaryFromPath() metrics = %#v, want expected averages and engram count", summary)
	}
}

func TestPlayerProfileFileSummaryFromPathReturnsArchiveAndPlayer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WritePlayerArchiveWithOptions(t, path, testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
		TribeID:       777,
		NumDeaths:     3,
	})

	summary, err := PlayerProfileFileSummaryFromPath(path)
	if err != nil {
		t.Fatalf("PlayerProfileFileSummaryFromPath() error = %v", err)
	}
	if summary.Archive.Path != path || summary.Archive.ArchiveVersion != 7 || summary.Archive.ObjectCount != 1 || len(summary.Archive.ClassNames) != 1 {
		t.Fatalf("PlayerProfileFileSummaryFromPath() archive = %#v", summary.Archive)
	}
	if summary.Player.PlayerDataID != 42 || summary.Player.CharacterName != "Survivor" || summary.Player.PlayerName != "PlatformName" || summary.Player.TribeID != 777 || summary.Player.NumDeaths != 3 {
		t.Fatalf("PlayerProfileFileSummaryFromPath() player = %#v", summary.Player)
	}
}

func TestPlayerProfileFileSummaryFromPathKeepsArchiveOnParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WriteArchive(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

	summary, err := PlayerProfileFileSummaryFromPath(path)
	if err == nil {
		t.Fatalf("PlayerProfileFileSummaryFromPath() error = nil, want missing player data")
	}
	if summary.Archive.Path != path || summary.Archive.ObjectCount != 1 || len(summary.Archive.ClassNames) != 1 {
		t.Fatalf("PlayerProfileFileSummaryFromPath() archive = %#v, want partial archive summary", summary.Archive)
	}
}

func TestTribeFileSummaryFromPathReturnsArchiveAndSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteTribeArchiveWithOptions(t, path, testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout"},
		MemberIDs: []int32{42, 43},
	})

	summary, err := TribeFileSummaryFromPath(path)
	if err != nil {
		t.Fatalf("TribeFileSummaryFromPath() error = %v", err)
	}
	if summary.Archive.Path != path || summary.Archive.ArchiveVersion != 7 || summary.Archive.ObjectCount != 1 || len(summary.Archive.ClassNames) != 1 {
		t.Fatalf("TribeFileSummaryFromPath() archive = %#v", summary.Archive)
	}
	if summary.Summary.Name != "Porters" || summary.Summary.TribeID != 12345 || summary.Summary.OwnerID != 42 || summary.Summary.NumDinos != 7 || len(summary.Summary.Members) != 2 {
		t.Fatalf("TribeFileSummaryFromPath() summary = %#v", summary.Summary)
	}
}

func TestPlayerAPITribeRosterSummaryForTribes(t *testing.T) {
	api := NewPlayer(nil)
	tribes := []arkobject.Tribe{
		{Name: "Porters", MemberIDs: []int32{42, 43}, NumDinos: 7},
		{MemberIDs: []int32{99}, NumDinos: 3},
	}

	summary := api.TribeRosterSummaryForTribes(tribes)
	want := TribeRosterSummary{Tribes: 2, WithNames: 1, Members: 3, Dinos: 10}
	if summary != want {
		t.Fatalf("TribeRosterSummaryForTribes() = %#v, want %#v", summary, want)
	}
}

func TestPlayerAPITribeRosterSummaryWithFaultsKeepsValidLocalTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		NumDinos:  7,
		MemberIDs: []int32{42, 43},
	})
	if err := os.WriteFile(filepath.Join(dir, "broken.arktribe"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken tribe: %v", err)
	}

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	summary, faults, err := api.TribeRosterSummaryWithFaults()
	if err != nil {
		t.Fatalf("TribeRosterSummaryWithFaults() error = %v", err)
	}
	want := TribeRosterSummary{Tribes: 1, WithNames: 1, Members: 2, Dinos: 7}
	if summary != want {
		t.Fatalf("TribeRosterSummaryWithFaults() summary = %#v, want %#v", summary, want)
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("TribeRosterSummaryWithFaults() faults = %#v, want one tribe fault", faults)
	}
}

func TestTribeRosterSummaryFromPathUsesDirectoryTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout"},
		MemberIDs: []int32{42, 43},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "789.arktribe"), testfixtures.TribeArchiveOptions{
		TribeID:  222,
		OwnerID:  43,
		NumDinos: 3,
		Members:  []string{"Builder"},
	})

	summary, faults, err := TribeRosterSummaryFromPath(dir)
	if err != nil {
		t.Fatalf("TribeRosterSummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("TribeRosterSummaryFromPath() faults = %#v, want none", faults)
	}
	want := TribeRosterSummary{Tribes: 2, WithNames: 1, Members: 3, Dinos: 10}
	if summary != want {
		t.Fatalf("TribeRosterSummaryFromPath() = %#v, want %#v", summary, want)
	}
}

func TestTribesFromPathReturnsTypedTribesAndFaults(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout"},
		MemberIDs: []int32{42, 43},
	})
	if err := os.WriteFile(filepath.Join(dir, "broken.arktribe"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken tribe: %v", err)
	}

	tribes, faults, err := TribesFromPath(dir, PlayerPathOptions{})
	if err != nil {
		t.Fatalf("TribesFromPath() error = %v", err)
	}
	if len(tribes) != 1 || tribes[0].TribeID != 12345 || tribes[0].Name != "Porters" || len(tribes[0].MemberIDs) != 2 {
		t.Fatalf("TribesFromPath() tribes = %#v, want typed local tribe", tribes)
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("TribesFromPath() faults = %#v, want one local tribe fault", faults)
	}
}

func TestTribeDirectorySummaryFromPathUsesDirectoryTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout"},
		MemberIDs: []int32{42, 43},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "789.arktribe"), testfixtures.TribeArchiveOptions{
		Name:     "Builders",
		TribeID:  222,
		OwnerID:  43,
		NumDinos: 3,
		Members:  []string{"Builder"},
	})

	summary, err := TribeDirectorySummaryFromPath(dir)
	if err != nil {
		t.Fatalf("TribeDirectorySummaryFromPath() error = %v", err)
	}
	if summary.Files != 2 || len(summary.Tribes) != 2 || summary.TotalMembers != 3 || summary.TotalDinos != 10 {
		t.Fatalf("TribeDirectorySummaryFromPath() = %#v, want two files, two tribes, three members, ten dinos", summary)
	}
	if !summary.HasAverageMembers || summary.AverageMembers != 1.5 {
		t.Fatalf("TribeDirectorySummaryFromPath() member average = %f, %v; want 1.5, true", summary.AverageMembers, summary.HasAverageMembers)
	}
	if !summary.HasAverageDinos || summary.AverageDinos != 5 {
		t.Fatalf("TribeDirectorySummaryFromPath() average = %f, %v; want 5, true", summary.AverageDinos, summary.HasAverageDinos)
	}
	ids := map[int32]bool{}
	for _, tribe := range summary.Tribes {
		ids[tribe.TribeID] = true
	}
	if !ids[222] || !ids[12345] {
		t.Fatalf("TribeDirectorySummaryFromPath() tribes = %#v, want both synthetic tribes", summary.Tribes)
	}
}

func TestPlayerAPIPlayerAndTribeDataSummaryForDataSortsPrivacySafeRows(t *testing.T) {
	api := NewPlayer(nil)
	players := []arkobject.Player{
		{CharacterName: "Ada", Level: 5, TribeID: 20},
		{PlayerName: "Grace", Level: 2, TribeID: 10},
		{Level: 5, TribeID: 99},
	}
	tribes := []arkobject.Tribe{
		{Name: "Porters", MemberIDs: []int32{1, 2}, NumDinos: 7, TribeID: 20},
		{MemberIDs: []int32{3}, NumDinos: 1, TribeID: 10},
	}
	relations := []TribePlayerRelation{
		{ActivePlayers: []arkobject.Player{players[0]}, InactiveMemberIDs: []int32{9}},
		{InactiveMemberIDs: []int32{3}},
	}

	summary := api.PlayerAndTribeDataSummaryForData(players, tribes, relations)
	if summary.Players != 3 || summary.Tribes != 2 || summary.PlayersWithNames != 2 ||
		summary.TribesWithNames != 1 || summary.ActiveLinks != 1 ||
		summary.InactiveMembers != 2 || summary.PlayersWithoutTribe != 1 ||
		summary.TribesWithInactive != 2 || summary.TribesWithoutActive != 1 {
		t.Fatalf("PlayerAndTribeDataSummaryForData() aggregate = %#v", summary)
	}
	wantPlayers := []PlayerDataRow{
		{HasCharacterName: false, HasPlayerName: true, Level: 2, TribeID: 10},
		{HasCharacterName: true, Level: 5, TribeID: 20},
		{Level: 5, TribeID: 99},
	}
	if !reflect.DeepEqual(summary.PlayerRows, wantPlayers) {
		t.Fatalf("PlayerRows = %#v, want %#v", summary.PlayerRows, wantPlayers)
	}
	wantTribes := []TribeDataRow{
		{Members: 1, Dinos: 1},
		{HasName: true, Members: 2, Dinos: 7},
	}
	if !reflect.DeepEqual(summary.TribeRows, wantTribes) {
		t.Fatalf("TribeRows = %#v, want %#v", summary.TribeRows, wantTribes)
	}
	wantRelations := []RelationDataRow{
		{InactiveMembers: 1},
		{ActiveMembers: 1, InactiveMembers: 1},
	}
	if !reflect.DeepEqual(summary.RelationRows, wantRelations) {
		t.Fatalf("RelationRows = %#v, want %#v", summary.RelationRows, wantRelations)
	}
}

func TestPlayerInventorySummaryFromPathFallsBackToDirectoryPlayers(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "inventory.ark")
	inventoryID := uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff")
	firstItemID := uuid.MustParse("44444444-4455-6677-8899-aabbccddeeff")
	secondItemID := uuid.MustParse("55555555-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff"): testfixtures.PlayerPawnGameObjectBytes(42, inventoryID),
			inventoryID: testfixtures.InventoryGameObjectBytes(
				inventoryID,
				firstItemID,
				secondItemID,
			),
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "42.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Fallback",
	})

	summary, faults, err := PlayerInventorySummaryFromPath(savePath)
	if err != nil {
		t.Fatalf("PlayerInventorySummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("PlayerInventorySummaryFromPath() faults = %#v, want none", faults)
	}
	want := PlayerInventorySummary{
		Players:       1,
		WithInventory: 1,
		TotalItems:    2,
		MaxItems:      2,
		MinItems:      2,
		AverageItems:  2,
	}
	if summary != want {
		t.Fatalf("PlayerInventorySummaryFromPath() = %#v, want %#v", summary, want)
	}
}

func TestPlayerAPIPlayerInventoryByDataIDIgnoresUnrelatedBrokenObjects(t *testing.T) {
	playerID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	pawnID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	inventoryID := uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff")
	firstItemID := uuid.MustParse("44444433-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("55555533-4455-6677-8899-aabbccddeeff")
	path := filepath.Join(t.TempDir(), "player-inventory.ark")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			playerID:    testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{PlayerDataID: 42, CharacterName: "Survivor", PlayerName: "PlatformName", TribeID: 12345}),
			pawnID:      testfixtures.PlayerPawnGameObjectBytes(42, inventoryID),
			inventoryID: testfixtures.InventoryGameObjectBytes(inventoryID, firstItemID),
			faultyID:    testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{PlayerDataID: 99, CharacterName: "Broken"})[:40],
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	inventory, ok, err := NewPlayer(save).PlayerInventoryByDataID(42)
	if err != nil {
		t.Fatalf("PlayerInventoryByDataID() error = %v", err)
	}
	if !ok || inventory.UUID != inventoryID || inventory.NumberOfItems() != 1 {
		t.Fatalf("PlayerInventoryByDataID() = %#v, %v; want inventory with one item", inventory, ok)
	}
}

func TestPlayerAPIPlayerLocationByDataID(t *testing.T) {
	save := openSyntheticPlayerTribeSave(t)
	defer save.Close()

	api := NewPlayer(save)
	location, ok, err := api.PlayerLocationByDataID(42)
	if err != nil {
		t.Fatalf("PlayerLocationByDataID() error = %v", err)
	}
	if !ok || location.X != 11 || location.Y != 22 || location.Z != 33 {
		t.Fatalf("PlayerLocationByDataID() = %#v, %v; want 11/22/33", location, ok)
	}
	if _, ok, err := api.PlayerLocationByDataID(999); err != nil || ok {
		t.Fatalf("PlayerLocationByDataID(missing) = ok %v err %v, want false nil", ok, err)
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

func gameModeCustomBytesFixture(t *testing.T) []byte {
	t.Helper()
	return testfixtures.GameModeCustomBytes(t, testfixtures.GameModeCustomBytesOptions{
		Player: testfixtures.PlayerArchiveOptions{
			PlayerDataID:        42,
			CharacterName:       "Survivor",
			PlayerName:          "PlatformName",
			UniqueID:            "eos-survivor",
			TribeID:             12345,
			NumDeaths:           3,
			ExtraCharacterLevel: 4,
			ExperiencePoints:    123.5,
			TotalEngramPoints:   12,
		},
		NextPlayer: testfixtures.PlayerArchiveOptions{
			PlayerDataID:  99,
			CharacterName: "Marker",
			PlayerName:    "MarkerPlatform",
			UniqueID:      "eos-marker",
			TribeID:       12345,
		},
		Tribe: testfixtures.TribeArchiveOptions{
			Name:      "Porters",
			TribeID:   12345,
			OwnerID:   42,
			NumDinos:  7,
			Members:   []string{"Survivor"},
			MemberIDs: []int32{42},
		},
	})
}

func openSyntheticPlayerTribeSave(t *testing.T) *arksave.Save {
	t.Helper()
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	playerID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	tribeID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	pawnID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	inventoryID := uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff")
	firstItemID := uuid.MustParse("44444444-4455-6677-8899-aabbccddeeff")
	secondItemID := uuid.MustParse("55555555-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			playerID: testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{
				PlayerDataID:        42,
				CharacterName:       "Survivor",
				PlayerName:          "PlatformName",
				UniqueID:            "eos-survivor",
				TribeID:             12345,
				NumDeaths:           3,
				ExtraCharacterLevel: 4,
				ExperiencePoints:    123.5,
				TotalEngramPoints:   12,
			}),
			tribeID: testfixtures.TribeGameObjectBytes(testfixtures.TribeArchiveOptions{
				Name:      "Porters",
				TribeID:   12345,
				OwnerID:   42,
				NumDinos:  7,
				Members:   []string{"Survivor"},
				MemberIDs: []int32{42},
			}),
			pawnID:      testfixtures.PlayerPawnGameObjectBytes(42, inventoryID),
			inventoryID: testfixtures.InventoryGameObjectBytes(inventoryID, firstItemID, secondItemID),
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open(save) error = %v", err)
	}
	return save
}

func TestPlayerAPIFiltersLocalPlayersByNamesAndUniqueID(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
		UniqueID:      "eos-survivor",
		TribeID:       777,
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  43,
		CharacterName: "Scout",
		PlayerName:    "OtherPlatform",
		UniqueID:      "eos-scout",
		TribeID:       888,
	})

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	byCharacter, err := api.PlayersByCharacterName("Survivor")
	if err != nil {
		t.Fatalf("PlayersByCharacterName() error = %v", err)
	}
	if len(byCharacter) != 1 || byCharacter[0].PlayerDataID != 42 {
		t.Fatalf("PlayersByCharacterName() = %#v, want player 42", byCharacter)
	}
	byPlatform, err := api.PlayersByPlayerName("OtherPlatform")
	if err != nil {
		t.Fatalf("PlayersByPlayerName() error = %v", err)
	}
	if len(byPlatform) != 1 || byPlatform[0].PlayerDataID != 43 {
		t.Fatalf("PlayersByPlayerName() = %#v, want player 43", byPlatform)
	}
	player, ok, err := api.PlayerByUniqueID("eos-survivor")
	if err != nil {
		t.Fatalf("PlayerByUniqueID() error = %v", err)
	}
	if !ok || player.PlayerDataID != 42 {
		t.Fatalf("PlayerByUniqueID() = %#v, %v; want player 42, true", player, ok)
	}
	if _, ok, err := api.PlayerByUniqueID("missing"); err != nil || ok {
		t.Fatalf("PlayerByUniqueID(missing) = ok %v err %v, want false nil", ok, err)
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

func TestPlayerAPIFiltersLocalTribesByNameOwnerAndMembers(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Ada", "Grace"},
		MemberIDs: []int32{42, 43},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "789.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Builders",
		TribeID:   67890,
		OwnerID:   99,
		NumDinos:  3,
		Members:   []string{"Linus"},
		MemberIDs: []int32{99},
	})

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	byName, err := api.TribesByName("Porters")
	if err != nil {
		t.Fatalf("TribesByName() error = %v", err)
	}
	if len(byName) != 1 || byName[0].TribeID != 12345 {
		t.Fatalf("TribesByName() = %#v, want tribe 12345", byName)
	}
	byOwner, err := api.TribesByOwnerID(99)
	if err != nil {
		t.Fatalf("TribesByOwnerID() error = %v", err)
	}
	if len(byOwner) != 1 || byOwner[0].TribeID != 67890 {
		t.Fatalf("TribesByOwnerID() = %#v, want tribe 67890", byOwner)
	}
	byMemberName, err := api.TribesByMemberName("Grace")
	if err != nil {
		t.Fatalf("TribesByMemberName() error = %v", err)
	}
	if len(byMemberName) != 1 || byMemberName[0].TribeID != 12345 {
		t.Fatalf("TribesByMemberName() = %#v, want tribe 12345", byMemberName)
	}
	byMemberID, err := api.TribesByMemberID(43)
	if err != nil {
		t.Fatalf("TribesByMemberID() error = %v", err)
	}
	if len(byMemberID) != 1 || byMemberID[0].TribeID != 12345 {
		t.Fatalf("TribesByMemberID() = %#v, want tribe 12345", byMemberID)
	}
}

func TestPlayerAPITribeDinoStatistics(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Ada", "Grace"},
		MemberIDs: []int32{42, 43},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "789.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Builders",
		TribeID:   67890,
		OwnerID:   99,
		NumDinos:  3,
		Members:   []string{"Linus"},
		MemberIDs: []int32{99},
	})

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	counts, err := api.TribeDinoCountsByID()
	if err != nil {
		t.Fatalf("TribeDinoCountsByID() error = %v", err)
	}
	if counts[12345] != 7 || counts[67890] != 3 {
		t.Fatalf("TribeDinoCountsByID() = %#v, want dino counts for tribes 12345 and 67890", counts)
	}
	total, err := api.TotalTribeDinos()
	if err != nil {
		t.Fatalf("TotalTribeDinos() error = %v", err)
	}
	if total != 10 {
		t.Fatalf("TotalTribeDinos() = %d, want 10", total)
	}
	average, ok, err := api.AverageTribeDinos()
	if err != nil {
		t.Fatalf("AverageTribeDinos() error = %v", err)
	}
	if !ok || average != 5 {
		t.Fatalf("AverageTribeDinos() = %f, %v; want 5, true", average, ok)
	}
	tribe, count, ok, err := api.TribeWithMostDinos()
	if err != nil {
		t.Fatalf("TribeWithMostDinos() error = %v", err)
	}
	if !ok || tribe.TribeID != 12345 || count != 7 {
		t.Fatalf("TribeWithMostDinos() = %#v, %d, %v; want tribe 12345, 7, true", tribe, count, ok)
	}
	tribe, count, ok, err = api.TribeWithFewestDinos()
	if err != nil {
		t.Fatalf("TribeWithFewestDinos() error = %v", err)
	}
	if !ok || tribe.TribeID != 67890 || count != 3 {
		t.Fatalf("TribeWithFewestDinos() = %#v, %d, %v; want tribe 67890, 3, true", tribe, count, ok)
	}
}

func TestPlayerAPIRelatesLocalPlayersTribesAndOwners(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
		UniqueID:      "eos-survivor",
		TribeID:       12345,
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  43,
		CharacterName: "Scout",
		PlayerName:    "OtherPlatform",
		UniqueID:      "eos-scout",
		TribeID:       12345,
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "789.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  77,
		CharacterName: "Nomad",
		PlayerName:    "NoTribe",
		UniqueID:      "eos-nomad",
		TribeID:       77777,
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout", "Inactive"},
		MemberIDs: []int32{42, 43, 99},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "789.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Sleepers",
		TribeID:   67890,
		OwnerID:   88,
		NumDinos:  0,
		Members:   []string{"Gone"},
		MemberIDs: []int32{88},
	})

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	tribePlayers, err := api.TribePlayerMap()
	if err != nil {
		t.Fatalf("TribePlayerMap() error = %v", err)
	}
	if len(tribePlayers[12345]) != 2 {
		t.Fatalf("TribePlayerMap()[12345] = %#v, want two active local players", tribePlayers[12345])
	}
	linkCount, err := api.TribePlayerLinkCount()
	if err != nil {
		t.Fatalf("TribePlayerLinkCount() error = %v", err)
	}
	if linkCount != 2 {
		t.Fatalf("TribePlayerLinkCount() = %d, want 2", linkCount)
	}
	relations, err := api.TribePlayerRelations()
	if err != nil {
		t.Fatalf("TribePlayerRelations() error = %v", err)
	}
	if len(relations) != 2 {
		t.Fatalf("TribePlayerRelations() length = %d, want 2", len(relations))
	}
	if len(relations[0].ActivePlayers) != 2 || len(relations[0].InactiveMemberIDs) != 1 || relations[0].InactiveMemberIDs[0] != 99 {
		t.Fatalf("TribePlayerRelations()[0] = %#v, want two active players and inactive member 99", relations[0])
	}
	if len(relations[0].InactiveMemberNames) != 1 || relations[0].InactiveMemberNames[0] != "Inactive" {
		t.Fatalf("TribePlayerRelations()[0].InactiveMemberNames = %#v, want Inactive", relations[0].InactiveMemberNames)
	}
	summary, err := api.TribePlayerRelationSummary()
	if err != nil {
		t.Fatalf("TribePlayerRelationSummary() error = %v", err)
	}
	wantSummary := TribePlayerRelationSummary{
		Players:             3,
		Tribes:              2,
		ActiveLinks:         2,
		InactiveMembers:     2,
		PlayersWithoutTribe: 1,
		TribesWithInactive:  2,
		TribesWithoutActive: 1,
	}
	if summary != wantSummary {
		t.Fatalf("TribePlayerRelationSummary() = %#v, want %#v", summary, wantSummary)
	}
	player, ok, err := api.PlayerByDataID(42)
	if err != nil || !ok {
		t.Fatalf("PlayerByDataID(42) = %#v, %v, %v; want player, true, nil", player, ok, err)
	}
	tribe, ok, err := api.TribeOfPlayer(player)
	if err != nil {
		t.Fatalf("TribeOfPlayer() error = %v", err)
	}
	if !ok || tribe.Name != "Porters" {
		t.Fatalf("TribeOfPlayer() = %#v, %v; want Porters, true", tribe, ok)
	}
	objectOwner, ok, err := api.ObjectOwnerByPlayerDataID(42)
	if err != nil {
		t.Fatalf("ObjectOwnerByPlayerDataID() error = %v", err)
	}
	if !ok || objectOwner.PlayerID != 42 || objectOwner.PlayerName != "PlatformName" || objectOwner.TribeName != "Porters" {
		t.Fatalf("ObjectOwnerByPlayerDataID() = %#v, %v; want profile-derived owner", objectOwner, ok)
	}
	dinoOwner, ok, err := api.DinoOwnerByPlayerDataID(42)
	if err != nil {
		t.Fatalf("DinoOwnerByPlayerDataID() error = %v", err)
	}
	if !ok || dinoOwner.PlayerID != 42 || dinoOwner.ImprinterName != "Survivor" || dinoOwner.TamerString != "Porters" {
		t.Fatalf("DinoOwnerByPlayerDataID() = %#v, %v; want profile-derived dino owner", dinoOwner, ok)
	}
}

func TestTribePlayerRelationSummaryFromPathUsesDirectoryPlayersAndTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerTribeRelationDirectory(t, dir)

	summary, err := TribePlayerRelationSummaryFromPath(dir)
	if err != nil {
		t.Fatalf("TribePlayerRelationSummaryFromPath() error = %v", err)
	}
	want := TribePlayerRelationSummary{
		Players:             3,
		Tribes:              2,
		ActiveLinks:         2,
		InactiveMembers:     2,
		PlayersWithoutTribe: 1,
		TribesWithInactive:  2,
		TribesWithoutActive: 1,
	}
	if summary != want {
		t.Fatalf("TribePlayerRelationSummaryFromPath() = %#v, want %#v", summary, want)
	}
}

func TestPlayerAndTribeDataSummaryFromPathUsesDirectoryPlayersAndTribes(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerTribeRelationDirectory(t, dir)

	summary, err := PlayerAndTribeDataSummaryFromPath(dir)
	if err != nil {
		t.Fatalf("PlayerAndTribeDataSummaryFromPath() error = %v", err)
	}
	if summary.Players != 3 || summary.Tribes != 2 || summary.ActiveLinks != 2 ||
		summary.InactiveMembers != 2 || summary.PlayersWithoutTribe != 1 {
		t.Fatalf("PlayerAndTribeDataSummaryFromPath() aggregate = %#v", summary)
	}
	if len(summary.PlayerRows) != 3 || len(summary.TribeRows) != 2 || len(summary.RelationRows) != 2 {
		t.Fatalf("PlayerAndTribeDataSummaryFromPath() rows = %#v", summary)
	}
}

func TestPlayerAPILocalDeathStatistics(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
		TribeID:       12345,
		NumDeaths:     3,
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  43,
		CharacterName: "Scout",
		PlayerName:    "OtherPlatform",
		TribeID:       12345,
		NumDeaths:     7,
	})

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	deaths, err := api.DeathsByPlayerID()
	if err != nil {
		t.Fatalf("DeathsByPlayerID() error = %v", err)
	}
	if deaths[42] != 3 || deaths[43] != 7 {
		t.Fatalf("DeathsByPlayerID() = %#v, want deaths for players 42 and 43", deaths)
	}
	total, err := api.TotalDeaths()
	if err != nil {
		t.Fatalf("TotalDeaths() error = %v", err)
	}
	if total != 10 {
		t.Fatalf("TotalDeaths() = %d, want 10", total)
	}
	averageDeaths, ok, err := api.AverageDeaths()
	if err != nil {
		t.Fatalf("AverageDeaths() error = %v", err)
	}
	if !ok || averageDeaths != 5 {
		t.Fatalf("AverageDeaths() = %f, %v; want 5, true", averageDeaths, ok)
	}
	player, deathsValue, ok, err := api.PlayerWithMostDeaths()
	if err != nil {
		t.Fatalf("PlayerWithMostDeaths() error = %v", err)
	}
	if !ok || player.PlayerDataID != 43 || deathsValue != 7 {
		t.Fatalf("PlayerWithMostDeaths() = %#v, %d, %v; want player 43, 7, true", player, deathsValue, ok)
	}
	player, deathsValue, ok, err = api.PlayerWithFewestDeaths()
	if err != nil {
		t.Fatalf("PlayerWithFewestDeaths() error = %v", err)
	}
	if !ok || player.PlayerDataID != 42 || deathsValue != 3 {
		t.Fatalf("PlayerWithFewestDeaths() = %#v, %d, %v; want player 42, 3, true", player, deathsValue, ok)
	}
}

func TestPlayerAPILocalLevelAndExperienceStatistics(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		TribeID:             12345,
		ExtraCharacterLevel: 4,
		ExperiencePoints:    123.5,
		TotalEngramPoints:   12,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        43,
		CharacterName:       "Scout",
		PlayerName:          "OtherPlatform",
		TribeID:             12345,
		ExtraCharacterLevel: 9,
		ExperiencePoints:    456.25,
		TotalEngramPoints:   30,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
			"Blueprint'/Game/Engrams/EngramC.EngramC_C'",
		},
	})

	api, err := NewPlayerFromDirectory(dir)
	if err != nil {
		t.Fatalf("NewPlayerFromDirectory() error = %v", err)
	}
	levels, err := api.LevelsByPlayerID()
	if err != nil {
		t.Fatalf("LevelsByPlayerID() error = %v", err)
	}
	if levels[42] != 5 || levels[43] != 10 {
		t.Fatalf("LevelsByPlayerID() = %#v, want levels 5 and 10", levels)
	}
	xp, err := api.ExperienceByPlayerID()
	if err != nil {
		t.Fatalf("ExperienceByPlayerID() error = %v", err)
	}
	if xp[42] != 123.5 || xp[43] != 456.25 {
		t.Fatalf("ExperienceByPlayerID() = %#v, want profile XP values", xp)
	}
	totalLevel, err := api.TotalLevel()
	if err != nil {
		t.Fatalf("TotalLevel() error = %v", err)
	}
	if totalLevel != 15 {
		t.Fatalf("TotalLevel() = %d, want 15", totalLevel)
	}
	totalExperience, err := api.TotalExperience()
	if err != nil {
		t.Fatalf("TotalExperience() error = %v", err)
	}
	if totalExperience != 579.75 {
		t.Fatalf("TotalExperience() = %f, want 579.75", totalExperience)
	}
	averageExperience, ok, err := api.AverageExperience()
	if err != nil {
		t.Fatalf("AverageExperience() error = %v", err)
	}
	if !ok || averageExperience != 289.875 {
		t.Fatalf("AverageExperience() = %f, %v; want 289.875, true", averageExperience, ok)
	}
	averageLevel, ok, err := api.AverageLevel()
	if err != nil {
		t.Fatalf("AverageLevel() error = %v", err)
	}
	if !ok || averageLevel != 7.5 {
		t.Fatalf("AverageLevel() = %f, %v; want 7.5, true", averageLevel, ok)
	}
	engramPoints, err := api.EngramPointsByPlayerID()
	if err != nil {
		t.Fatalf("EngramPointsByPlayerID() error = %v", err)
	}
	if engramPoints[42] != 12 || engramPoints[43] != 30 {
		t.Fatalf("EngramPointsByPlayerID() = %#v, want profile engram point values", engramPoints)
	}
	totalEngramPoints, err := api.TotalEngramPoints()
	if err != nil {
		t.Fatalf("TotalEngramPoints() error = %v", err)
	}
	if totalEngramPoints != 42 {
		t.Fatalf("TotalEngramPoints() = %d, want 42", totalEngramPoints)
	}
	unlockedEngrams, err := api.UnlockedEngrams()
	if err != nil {
		t.Fatalf("UnlockedEngrams() error = %v", err)
	}
	wantEngrams := []string{
		"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
		"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
		"Blueprint'/Game/Engrams/EngramC.EngramC_C'",
	}
	if len(unlockedEngrams) != len(wantEngrams) {
		t.Fatalf("UnlockedEngrams() = %#v, want %#v", unlockedEngrams, wantEngrams)
	}
	for i := range wantEngrams {
		if unlockedEngrams[i] != wantEngrams[i] {
			t.Fatalf("UnlockedEngrams() = %#v, want %#v", unlockedEngrams, wantEngrams)
		}
	}
	player, level, ok, err := api.PlayerWithHighestLevel()
	if err != nil {
		t.Fatalf("PlayerWithHighestLevel() error = %v", err)
	}
	if !ok || player.PlayerDataID != 43 || level != 10 {
		t.Fatalf("PlayerWithHighestLevel() = %#v, %d, %v; want player 43, level 10, true", player, level, ok)
	}
	player, level, ok, err = api.PlayerWithLowestLevel()
	if err != nil {
		t.Fatalf("PlayerWithLowestLevel() error = %v", err)
	}
	if !ok || player.PlayerDataID != 42 || level != 5 {
		t.Fatalf("PlayerWithLowestLevel() = %#v, %d, %v; want player 42, level 5, true", player, level, ok)
	}
	player, experience, ok, err := api.PlayerWithHighestExperience()
	if err != nil {
		t.Fatalf("PlayerWithHighestExperience() error = %v", err)
	}
	if !ok || player.PlayerDataID != 43 || experience != 456.25 {
		t.Fatalf("PlayerWithHighestExperience() = %#v, %f, %v; want player 43, XP 456.25, true", player, experience, ok)
	}
	player, experience, ok, err = api.PlayerWithLowestExperience()
	if err != nil {
		t.Fatalf("PlayerWithLowestExperience() error = %v", err)
	}
	if !ok || player.PlayerDataID != 42 || experience != 123.5 {
		t.Fatalf("PlayerWithLowestExperience() = %#v, %f, %v; want player 42, XP 123.5, true", player, experience, ok)
	}
	player, points, ok, err := api.PlayerWithMostEngramPoints()
	if err != nil {
		t.Fatalf("PlayerWithMostEngramPoints() error = %v", err)
	}
	if !ok || player.PlayerDataID != 43 || points != 30 {
		t.Fatalf("PlayerWithMostEngramPoints() = %#v, %d, %v; want player 43, 30, true", player, points, ok)
	}
	player, points, ok, err = api.PlayerWithFewestEngramPoints()
	if err != nil {
		t.Fatalf("PlayerWithFewestEngramPoints() error = %v", err)
	}
	if !ok || player.PlayerDataID != 42 || points != 12 {
		t.Fatalf("PlayerWithFewestEngramPoints() = %#v, %d, %v; want player 42, 12, true", player, points, ok)
	}
}

func TestPlayerDirectorySummaryFromPathUsesDirectoryPlayers(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		TribeID:             777,
		NumDeaths:           3,
		ExtraCharacterLevel: 9,
		ExperiencePoints:    12.5,
		TotalEngramPoints:   14,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        43,
		CharacterName:       "Scout",
		PlayerName:          "OtherPlatform",
		TribeID:             888,
		NumDeaths:           1,
		ExtraCharacterLevel: 4,
		ExperiencePoints:    7.5,
		TotalEngramPoints:   6,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
			"Blueprint'/Game/Engrams/EngramC.EngramC_C'",
		},
	})

	summary, err := PlayerDirectorySummaryFromPath(dir)
	if err != nil {
		t.Fatalf("PlayerDirectorySummaryFromPath() error = %v", err)
	}
	if summary.Files != 2 || len(summary.Players) != 2 || summary.TotalDeaths != 4 || summary.TotalLevel != 15 {
		t.Fatalf("PlayerDirectorySummaryFromPath() = %#v, want two files/players, four deaths, level fifteen", summary)
	}
	if summary.HighestLevel != 10 || summary.HighestExperience != 12.5 {
		t.Fatalf("PlayerDirectorySummaryFromPath() maxima = level %d experience %f; want 10 and 12.5", summary.HighestLevel, summary.HighestExperience)
	}
	if !summary.HasAverageDeaths || summary.AverageDeaths != 2 || !summary.HasAverageLevel || summary.AverageLevel != 7.5 || !summary.HasAverageExperience || summary.AverageExperience != 10 {
		t.Fatalf("PlayerDirectorySummaryFromPath() averages = %#v", summary)
	}
	if summary.TotalExperience != 20 || summary.TotalEngramPoints != 20 || summary.UnlockedEngrams != 3 {
		t.Fatalf("PlayerDirectorySummaryFromPath() totals = %#v, want expected XP, engram points, and unlocked count", summary)
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
	testfixtures.WriteTributeFile(t, tributePath, []uint64{11, 22}, []uint64{33})

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

func syntheticLegacyArchiveBytes(t *testing.T, className string) []byte {
	t.Helper()

	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	testfixtures.WriteInt32(&buf, 6)
	testfixtures.WriteInt32(&buf, 11)
	testfixtures.WriteInt32(&buf, 22)
	testfixtures.WriteInt32(&buf, 1)
	buf.Write(id[:])
	testfixtures.WriteArkString(&buf, className)
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteStringArray(&buf, []string{"Legacy_0"})
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, -1)
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteInt32(&buf, 128)
	testfixtures.WriteUInt32(&buf, 0)
	return buf.Bytes()
}
