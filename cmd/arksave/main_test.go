package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestInspectCommandPrintsOfflineSaveSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)

	var out bytes.Buffer
	err := run([]string{"inspect", path}, &out)
	if err != nil {
		t.Fatalf("run(inspect) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Map: Valguero_WP",
		"Save version: 12",
		"Objects: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("inspect output %q does not contain %q", got, want)
		}
	}
}

func TestRunRejectsNetworkCommands(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"rcon"}, &out)
	if err == nil {
		t.Fatalf("run(rcon) error = nil, want unsupported command")
	}
	if !strings.Contains(err.Error(), "offline-only") {
		t.Fatalf("run(rcon) error = %v, want offline-only message", err)
	}
}

func TestRunRejectsUnknownOption(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"--verbose", "inspect", "save.ark"}, &out)
	if err == nil {
		t.Fatalf("run(unknown option) error = nil, want unknown option")
	}
	if !strings.Contains(err.Error(), "unknown option") {
		t.Fatalf("run(unknown option) error = %v, want unknown option message", err)
	}
}

func TestArchiveSummaryPrintsPropertyErrorCount(t *testing.T) {
	var out bytes.Buffer
	err := printArchiveSummary(&out, "Archive", "/tmp/profile.arkprofile", 7, []arkarchive.Object{
		{ClassName: "/Game/Valid.Valid_C"},
		{ClassName: "/Game/Broken.Broken_C", PropertyError: errors.New("bad property")},
	}, runOptions{})
	if err != nil {
		t.Fatalf("printArchiveSummary() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Property parse errors: 1") {
		t.Fatalf("archive summary %q does not contain property error count", got)
	}
}

func TestPlayersCommandReturnsErrorWhenParsedProfileSummaryMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	createSyntheticArchive(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

	var out bytes.Buffer
	err := run([]string{"players", path}, &out)
	if err == nil {
		t.Fatalf("run(players) error = nil, want missing normalized profile error")
	}
	if !strings.Contains(err.Error(), "parse player profile details") {
		t.Fatalf("run(players) error = %v, want normalized parse context", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player profile:",
		"Archive version: 7",
		"Objects: 1",
		"/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("players output %q does not contain %q", got, want)
		}
	}
}

func TestPlayersCommandPrintsParsedLocalProfileSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WritePlayerArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"players", path}, &out)
	if err != nil {
		t.Fatalf("run(players) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player profile:",
		"Character name: Survivor",
		"Player name: PlatformName",
		"Player data ID: 42",
		"Tribe ID: 777",
		"Deaths: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("players output %q does not contain %q", got, want)
		}
	}
}

func TestPlayersCommandRedactsLocalProfileDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WritePlayerArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"players", "--redact", path}, &out)
	if err != nil {
		t.Fatalf("run(players --redact) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player profile: [redacted]",
		"Character name: [redacted]",
		"Player name: [redacted]",
		"Player data ID: [redacted]",
		"Tribe ID: [redacted]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted players output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "Survivor", "PlatformName", "/Game/"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted players output %q contains private detail %q", got, leaked)
		}
	}
}

func TestPlayersCommandPrintsDirectorySummary(t *testing.T) {
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

	var out bytes.Buffer
	err := run([]string{"players", dir}, &out)
	if err != nil {
		t.Fatalf("run(players directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player directory:",
		"Profiles: 2",
		"Players: 2",
		"Total deaths: 4",
		"Average deaths: 2.00",
		"Total level: 15",
		"Average level: 7.50",
		"Total experience: 20.00",
		"Total engram points: 20",
		"Unlocked engrams: 3",
		"  player id=42 character=Survivor platform=PlatformName tribe=777 level=10 deaths=3",
		"  player id=43 character=Scout platform=OtherPlatform tribe=888 level=5 deaths=1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("players directory output %q does not contain %q", got, want)
		}
	}
}

func TestTribesCommandPrintsLocalTribeSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteTribeArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"tribes", path}, &out)
	if err != nil {
		t.Fatalf("run(tribes) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save:",
		"Archive version: 7",
		"Objects: 1",
		"/Script/ShooterGame.PrimalTribeData",
		"Tribe name: Porters",
		"Tribe ID: 12345",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes output %q does not contain %q", got, want)
		}
	}
}

func TestTribesCommandRedactsLocalTribeDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteTribeArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"--redact", "tribes", path}, &out)
	if err != nil {
		t.Fatalf("run(--redact tribes) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save: [redacted]",
		"Tribe name: [redacted]",
		"Tribe ID: [redacted]",
		"Owner ID: [redacted]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted tribes output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "Porters", "/Script/ShooterGame.PrimalTribeData"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted tribes output %q contains private detail %q", got, leaked)
		}
	}
}

func TestTribesCommandPrintsDirectorySummary(t *testing.T) {
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

	var out bytes.Buffer
	err := run([]string{"tribes", dir}, &out)
	if err != nil {
		t.Fatalf("run(tribes directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe directory:",
		"Tribe files: 2",
		"Tribes: 2",
		"Total dinos: 10",
		"Average dinos: 5.00",
		"  tribe id=222 name=Builders owner=43 members=1 dinos=3",
		"  tribe id=12345 name=Porters owner=42 members=2 dinos=7",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes directory output %q does not contain %q", got, want)
		}
	}
}

func TestTribesCommandReturnsErrorWhenParsedSummaryMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	createSyntheticArchive(t, path, "/Script/ShooterGame.PrimalTribeData")

	var out bytes.Buffer
	err := run([]string{"tribes", path}, &out)
	if err == nil {
		t.Fatalf("run(tribes) error = nil, want missing normalized tribe error")
	}
	if !strings.Contains(err.Error(), "parse tribe details") {
		t.Fatalf("run(tribes) error = %v, want normalized parse context", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save:",
		"Archive version: 7",
		"Objects: 1",
		"/Script/ShooterGame.PrimalTribeData",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "Tribe name:") {
		t.Fatalf("tribes output %q includes summary despite missing TribeData", got)
	}
}

func TestClusterCommandPrintsLocalClusterSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	createSyntheticArchive(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"cluster", path}, &out)
	if err != nil {
		t.Fatalf("run(cluster) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster file:",
		"Archive version: 7",
		"Objects: 1",
		"Items: 0",
		"Dinos: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster output %q does not contain %q", got, want)
		}
	}
}

func TestClusterSummaryPrintsDinoParseErrors(t *testing.T) {
	var out bytes.Buffer
	err := printClusterSummary(&out, &arkcluster.Data{
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Dinos: []arkcluster.Dino{{
			Index:      0,
			UploadTime: 12345,
			RawSize:    32,
			ParseError: "unsupported embedded dino archive",
		}},
	}, runOptions{})
	if err != nil {
		t.Fatalf("printClusterSummary() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "parse_error=unsupported embedded dino archive") {
		t.Fatalf("cluster summary %q does not contain dino parse error", got)
	}
}

func TestClusterCommandRedactsPathAndUploadDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	createSyntheticArchive(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"--redact", "cluster", path}, &out)
	if err != nil {
		t.Fatalf("run(--redact cluster) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster file: [redacted]",
		"Archive version: 7",
		"Objects: 1",
		"Items: 0",
		"Dinos: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted cluster output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "EOS_abc123", "blueprint="} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted cluster output %q contains private detail %q", got, leaked)
		}
	}
}

func TestTributeCommandPrintsLocalTributeSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.arktributetribe")
	createSyntheticTribute(t, path, []uint64{11, 22}, []uint64{33})

	var out bytes.Buffer
	err := run([]string{"tribute", path}, &out)
	if err != nil {
		t.Fatalf("run(tribute) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribute file:",
		"Player data IDs: 2",
		"Tribe data IDs: 1",
		"player_data_id=11",
		"player_data_id=22",
		"tribe_data_id=33",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribute output %q does not contain %q", got, want)
		}
	}
}

func TestTributeCommandRedactsLocalTributeDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.arktributetribe")
	createSyntheticTribute(t, path, []uint64{11, 22}, []uint64{33})

	var out bytes.Buffer
	err := run([]string{"--redact", "tribute", path}, &out)
	if err != nil {
		t.Fatalf("run(--redact tribute) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribute file: [redacted]",
		"Player data IDs: 2",
		"Tribe data IDs: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted tribute output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "player_data_id=", "tribe_data_id=", "abc.arktributetribe"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted tribute output %q contains private detail %q", got, leaked)
		}
	}
}

func TestExportJSONWritesSaveInfoToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "save_info.json")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"export-json", savePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-json output %q does not mention %q", out.String(), outPath)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(exported json) error = %v", err)
	}
	var decoded struct {
		MapName     string `json:"map_name"`
		SaveVersion int16  `json:"save_version"`
		ObjectCount int    `json:"object_count"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, data)
	}
	if decoded.MapName != "Valguero_WP" || decoded.SaveVersion != 12 || decoded.ObjectCount != 1 {
		t.Fatalf("exported json = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportJSONRedactsObjectDetailsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "save_info.json")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"--redact", "export-json", savePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(--redact export-json) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-json output %q mentions output path %q", out.String(), outPath)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted json) error = %v", err)
	}
	var decoded arkapi.SaveInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, data)
	}
	if decoded.ObjectCount != 1 || len(decoded.Objects) != 0 {
		t.Fatalf("redacted SaveInfo = %#v, want object count without object details", decoded)
	}
	if strings.Contains(string(data), "00010203") || strings.Contains(string(data), "Blueprint'/Game/Test.Test_C'") {
		t.Fatalf("redacted save info contains object detail: %s", data)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONWritesClusterSummaryToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	outPath := filepath.Join(dir, "cluster.json")
	createSyntheticArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"export-cluster-json", clusterPath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-cluster-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-cluster-json output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(cluster json) error = %v", err)
	}
	var decoded arkapi.ClusterDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ID != "EOS_abc123" || decoded.ArchiveVersion != 7 || decoded.ObjectCount != 1 {
		t.Fatalf("decoded ClusterDataInfo = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONRedactsIdentifiersWhenRequested(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	outPath := filepath.Join(dir, "cluster.json")
	createSyntheticArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"export-cluster-json", "--redact", clusterPath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-cluster-json --redact) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-cluster-json output %q mentions output path %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted cluster json) error = %v", err)
	}
	var decoded arkapi.ClusterDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ID != "[redacted]" || decoded.Path != "[redacted]" || decoded.ObjectCount != 1 || len(decoded.Items) != 0 || len(decoded.Dinos) != 0 {
		t.Fatalf("redacted ClusterDataInfo = %#v", decoded)
	}
	if strings.Contains(string(raw), clusterPath) || strings.Contains(string(raw), "EOS_abc123") {
		t.Fatalf("redacted cluster json contains private detail: %s", raw)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONWritesDirectorySummary(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	outPath := filepath.Join(dir, "clusters.json")
	createSyntheticArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	createSyntheticArchive(t, filepath.Join(dir, "EOS_def456"), "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"export-cluster-json", dir, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-cluster-json directory) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-cluster-json directory output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(cluster directory json) error = %v", err)
	}
	var decoded arkapi.ClusterDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(cluster directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 2 || len(decoded.Files) != 2 {
		t.Fatalf("decoded ClusterDirectoryInfo = %#v, want two files", decoded)
	}
	if decoded.Files[0].ID != "EOS_abc123" || decoded.Files[1].ID != "EOS_def456" {
		t.Fatalf("decoded cluster IDs = %#v", decoded.Files)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportDomainJSONWritesDomainSummaryToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "stackables.json")
	createSyntheticEmptySave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"export-domain-json", savePath, "stackables", outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-domain-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-domain-json output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(domain json) error = %v", err)
	}
	var decoded arkapi.DomainExport
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.Domain != "stackables" {
		t.Fatalf("decoded DomainExport = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportDomainJSONRedactsItemsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "stackables.json")
	createSyntheticEmptySave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"--redact", "export-domain-json", savePath, "stackables", outPath}, &out)
	if err != nil {
		t.Fatalf("run(--redact export-domain-json) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-domain-json output %q mentions output path %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted domain json) error = %v", err)
	}
	var decoded arkapi.DomainExport
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.Domain != "stackables" || decoded.Count != 0 || decoded.Items != nil {
		t.Fatalf("redacted DomainExport = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestMutateCopyCommandWritesExplicitOutput(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "copy.ark")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"mutate", "copy", savePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(mutate copy) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("mutate copy output %q does not mention %q", out.String(), outPath)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("stat mutated copy: %v", err)
	}
}

func TestMutateCommandsRedactOutputDetailsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	copyPath := filepath.Join(dir, "copy.ark")
	removedPath := filepath.Join(dir, "removed.ark")
	objectPath := filepath.Join(dir, "object.ark")
	createSyntheticSave(t, savePath)
	objectID := "00010203-0405-0607-0809-0a0b0c0d0e0f"

	var copyOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "copy", savePath, copyPath}, &copyOut); err != nil {
		t.Fatalf("run(--redact mutate copy) error = %v", err)
	}
	if strings.Contains(copyOut.String(), copyPath) || !strings.Contains(copyOut.String(), "[redacted]") {
		t.Fatalf("redacted mutate copy output = %q", copyOut.String())
	}

	var removeOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "remove-object", savePath, removedPath, objectID}, &removeOut); err != nil {
		t.Fatalf("run(--redact mutate remove-object) error = %v", err)
	}
	got := removeOut.String()
	if strings.Contains(got, removedPath) || strings.Contains(got, objectID) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate remove-object output = %q", got)
	}

	var objectOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "put-object-hex", savePath, objectPath, objectID, "090807"}, &objectOut); err != nil {
		t.Fatalf("run(--redact mutate put-object-hex) error = %v", err)
	}
	got = objectOut.String()
	if strings.Contains(got, objectPath) || strings.Contains(got, objectID) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate put-object-hex output = %q", got)
	}
}

func TestMutateRemoveObjectCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "removed.ark")
	createSyntheticSave(t, savePath)
	objectID := "00010203-0405-0607-0809-0a0b0c0d0e0f"

	var out bytes.Buffer
	err := run([]string{"mutate", "remove-object", savePath, outPath, objectID}, &out)
	if err != nil {
		t.Fatalf("run(mutate remove-object) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), objectID) {
		t.Fatalf("mutate remove-object output %q missing path or uuid", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	ids, err := save.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs(mutated output) error = %v", err)
	}
	_ = save.Close()
	if len(ids) != 0 {
		t.Fatalf("mutated output ObjectIDs length = %d, want 0", len(ids))
	}
}

func TestMutatePutCustomCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "custom.ark")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"mutate", "put-custom", savePath, outPath, "Extra", "090807"}, &out)
	if err != nil {
		t.Fatalf("run(mutate put-custom) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), "Extra") {
		t.Fatalf("mutate put-custom output %q missing path or key", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.CustomValue("Extra")
	if err != nil {
		t.Fatalf("CustomValue(Extra) error = %v", err)
	}
	_ = save.Close()
	if !bytes.Equal(got, []byte{9, 8, 7}) {
		t.Fatalf("CustomValue(Extra) = % x, want 09 08 07", got)
	}
}

func TestMutatePutObjectHexCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "object.ark")
	createSyntheticSave(t, savePath)
	objectID := uuid.MustParse("00010203-0405-0607-0809-0a0b0c0d0e0f")

	var out bytes.Buffer
	err := run([]string{"mutate", "put-object-hex", savePath, outPath, objectID.String(), "090807"}, &out)
	if err != nil {
		t.Fatalf("run(mutate put-object-hex) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), objectID.String()) {
		t.Fatalf("mutate put-object-hex output %q missing path or uuid", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", objectID, err)
	}
	_ = save.Close()
	if !bytes.Equal(got, []byte{9, 8, 7}) {
		t.Fatalf("ObjectBinary(%s) = % x, want 09 08 07", objectID, got)
	}
}

func createSyntheticSave(t *testing.T, path string) {
	t.Helper()
	objectID := uuid.MustParse("00010203-0405-0607-0809-0a0b0c0d0e0f")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			1: "Blueprint'/Game/Test.Test_C'",
			2: "None",
		}),
		Objects: map[uuid.UUID][]byte{
			objectID: testfixtures.GenericObjectBytes(1, 2),
		},
	})
}

func createSyntheticEmptySave(t *testing.T, path string) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header:      testfixtures.Header("Valguero_WP", map[uint32]string{1: "Blueprint'/Game/Test.Test_C'"}),
		EmptyTables: true,
	})
}

func createSyntheticArchive(t *testing.T, path string, className string) {
	t.Helper()
	testfixtures.WriteArchive(t, path, className)
}

func createSyntheticTribute(t *testing.T, path string, playerIDs []uint64, tribeIDs []uint64) {
	t.Helper()
	testfixtures.WriteTributeFile(t, path, playerIDs, tribeIDs)
}

func assertPrivateFileMode(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat exported file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("exported file mode = %o, want 600", got)
	}
}
