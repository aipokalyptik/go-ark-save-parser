package examples_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestExamplesRunAgainstLocalSyntheticFixtures(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	copyPath := filepath.Join(dir, "copy.ark")
	clusterPath := filepath.Join(dir, "EOS_abc123")
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	dinoID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000000: "None",
			0x10000001: "Blueprint'/Game/Test.Test_C'",
			0x10000002: "None",
			0x10000003: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		}),
		Objects: map[uuid.UUID][]byte{
			objectID: testfixtures.GenericObjectBytes(0x10000001, 0x10000002),
			dinoID:   testfixtures.GenericObjectBytes(0x10000003, 0x10000002),
		},
	})
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteTributeFile(t, tributePath, []uint64{11, 22}, []uint64{33})

	runExample(t, "map_summary", "map=Valguero_WP", savePath)
	runExample(t, "object_classes", "Blueprint'/Game/Test.Test_C'", savePath)
	runExample(t, "property_filter", "objects=2 classes=2", savePath, "None")
	runExample(t, "dino_filter", "dinos=1 tamed=0 wild=1 classes=1", savePath)
	runExample(t, "local_profiles", "clusters=1 tributes=1", dir)
	runExample(t, "cluster_json", `"id": "EOS_abc123"`, clusterPath)
	runExample(t, "local_tribute", "player_data_ids=2", tributePath)
	runExample(t, "mutation_copy", "wrote copy:", savePath, copyPath)
	if _, err := os.Stat(copyPath); err != nil {
		t.Fatalf("mutation_copy output missing: %v", err)
	}
}

func runExample(t *testing.T, name string, want string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"run", "./" + name}, args...)
	cmd := exec.Command("go", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run ./%s %v error = %v\n%s", name, args, err, out)
	}
	if !strings.Contains(string(out), want) {
		t.Fatalf("go run ./%s output %q does not contain %q", name, out, want)
	}
}
