package arkapi

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

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

func savePlayerObjectBytes(t *testing.T, opts testfixtures.PlayerArchiveOptions) []byte {
	t.Helper()
	var myData bytes.Buffer
	testfixtures.WriteNameIntProperty(&myData, "PlayerDataID", opts.PlayerDataID)
	testfixtures.WriteNameStringProperty(&myData, "PlayerCharacterName", opts.CharacterName)
	testfixtures.WriteNameStringProperty(&myData, "PlayerName", opts.PlayerName)
	if opts.UniqueID != "" {
		testfixtures.WriteNameStringProperty(&myData, "UniqueID", opts.UniqueID)
	}
	testfixtures.WriteNameIntProperty(&myData, "TribeID", opts.TribeID)
	if opts.NumDeaths != 0 {
		testfixtures.WriteNameIntProperty(&myData, "NumOfDeaths", opts.NumDeaths)
	}
	testfixtures.WriteArkString(&myData, "None")

	var props bytes.Buffer
	testfixtures.WriteNameIntProperty(&props, "SavedPlayerDataVersion", 17)
	if opts.ExtraCharacterLevel != 0 {
		testfixtures.WriteNameIntProperty(&props, "CharacterStatusComponent_ExtraCharacterLevel", opts.ExtraCharacterLevel)
	}
	if opts.ExperiencePoints != 0 {
		testfixtures.WriteNameFloatProperty(&props, "CharacterStatusComponent_ExperiencePoints", opts.ExperiencePoints)
	}
	if opts.TotalEngramPoints != 0 {
		testfixtures.WriteNameIntProperty(&props, "PlayerState_TotalEngramPoints", opts.TotalEngramPoints)
	}
	testfixtures.WriteNameStructProperty(&props, "MyData", "PlayerDataStruct", myData.Bytes())
	testfixtures.WriteArkString(&props, "None")
	return saveObjectBytes("/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C", []string{"PlayerData_0"}, props.Bytes())
}

func openSyntheticPlayerTribeSave(t *testing.T) *arksave.Save {
	t.Helper()
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	playerID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	tribeID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	pawnID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			playerID: savePlayerObjectBytes(t, testfixtures.PlayerArchiveOptions{
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
			tribeID: saveTribeObjectBytes(t, testfixtures.TribeArchiveOptions{
				Name:      "Porters",
				TribeID:   12345,
				OwnerID:   42,
				NumDinos:  7,
				Members:   []string{"Survivor"},
				MemberIDs: []int32{42},
			}),
			pawnID: savePlayerPawnObjectBytes(42),
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open(save) error = %v", err)
	}
	return save
}

func savePlayerPawnObjectBytes(playerDataID int32) []byte {
	var props bytes.Buffer
	testfixtures.WriteNameIntProperty(&props, "LinkedPlayerDataID", playerDataID)
	testfixtures.WriteArkString(&props, "None")
	return saveObjectBytes("Blueprint'/Game/PrimalEarth/CoreBlueprints/PlayerPawnTest.PlayerPawnTest_C'", []string{"PlayerPawn_0"}, props.Bytes())
}

func saveTribeObjectBytes(t *testing.T, opts testfixtures.TribeArchiveOptions) []byte {
	t.Helper()
	var tribeData bytes.Buffer
	testfixtures.WriteNameStringProperty(&tribeData, "TribeName", opts.Name)
	testfixtures.WriteNameIntProperty(&tribeData, "TribeID", opts.TribeID)
	testfixtures.WriteNameIntProperty(&tribeData, "OwnerPlayerDataId", opts.OwnerID)
	testfixtures.WriteNameIntProperty(&tribeData, "NumTribeDinos", opts.NumDinos)
	if len(opts.Members) > 0 {
		testfixtures.WriteNameStringArrayProperty(&tribeData, "MembersPlayerName", opts.Members)
	}
	if len(opts.MemberIDs) > 0 {
		testfixtures.WriteNameIntArrayProperty(&tribeData, "MembersPlayerDataID", opts.MemberIDs)
	}
	testfixtures.WriteArkString(&tribeData, "None")

	var props bytes.Buffer
	testfixtures.WriteNameStructProperty(&props, "TribeData", "TribeDataStruct", tribeData.Bytes())
	testfixtures.WriteArkString(&props, "None")
	return saveObjectBytes("/Script/ShooterGame.PrimalTribeData", []string{"TribeData_0"}, props.Bytes())
}

func saveObjectBytes(className string, names []string, properties []byte) []byte {
	var buf bytes.Buffer
	testfixtures.WriteArkString(&buf, className)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(len(names)))
	for _, name := range names {
		testfixtures.WriteArkString(&buf, name)
	}
	_ = binary.Write(&buf, binary.LittleEndian, int32(-1))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	buf.Write(properties)
	return buf.Bytes()
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
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout", "Inactive"},
		MemberIDs: []int32{42, 43, 99},
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
