package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
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

func TestPlayersCommandPrintsLocalProfileSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	createSyntheticArchive(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

	var out bytes.Buffer
	err := run([]string{"players", path}, &out)
	if err != nil {
		t.Fatalf("run(players) error = %v", err)
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

func TestTribesCommandKeepsMetadataWhenSummaryMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	createSyntheticArchive(t, path, "/Script/ShooterGame.PrimalTribeData")

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
