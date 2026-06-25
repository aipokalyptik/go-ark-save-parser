package arksave

import (
	"bytes"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestOpenReadsHeaderCustomValuesAndGameObjects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	objectBytes := syntheticObjectBytes(0x10000001)
	secondObjectID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	secondObjectBytes := syntheticObjectBytes(0x10000005)
	header := syntheticHeader()
	actorTransforms := testfixtures.ActorTransforms(testfixtures.ActorTransform{
		UUID:       objectID,
		X:          1,
		Y:          2,
		Z:          3,
		Pitch:      4,
		Roll:       5,
		Yaw:        6,
		Quaternion: 7,
	})
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: header,
		Custom: map[string][]byte{
			"ActorTransforms": actorTransforms,
		},
		Objects: map[uuid.UUID][]byte{
			objectID:       objectBytes,
			secondObjectID: secondObjectBytes,
		},
	})

	save, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	if save.Context.SaveVersion != 12 {
		t.Fatalf("SaveVersion = %d, want 12", save.Context.SaveVersion)
	}
	if save.Context.MapName != "Valguero_WP" {
		t.Fatalf("MapName = %q, want Valguero_WP", save.Context.MapName)
	}
	if save.Context.GameTime != 1234.5 {
		t.Fatalf("GameTime = %f, want 1234.5", save.Context.GameTime)
	}
	if got := save.Context.Name(0x10000001); got != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("name table lookup = %q", got)
	}

	custom, err := save.CustomValue("ActorTransforms")
	if err != nil {
		t.Fatalf("CustomValue(ActorTransforms) error = %v", err)
	}
	if !bytes.Equal(custom, actorTransforms) {
		t.Fatalf("ActorTransforms bytes = % x, want % x", custom, actorTransforms)
	}
	transform, ok := save.Context.ActorTransforms[objectID]
	if !ok {
		t.Fatalf("Context.ActorTransforms missing %s", objectID)
	}
	if transform.X != 1 || transform.Y != 2 || transform.Z != 3 || transform.Pitch != 4 || transform.Roll != 5 || transform.Yaw != 6 || transform.Quaternion != 7 {
		t.Fatalf("ActorTransform = %#v", transform)
	}
	if save.Context.ActorTransformPositions[objectID] != 0 {
		t.Fatalf("ActorTransformPositions[%s] = %d, want 0", objectID, save.Context.ActorTransformPositions[objectID])
	}

	ids, err := save.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs() error = %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("ObjectIDs() length = %d, want 2", len(ids))
	}

	className, err := save.ClassOf(objectID)
	if err != nil {
		t.Fatalf("ClassOf() error = %v", err)
	}
	if className != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("ClassOf() = %q", className)
	}

	raw, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary() error = %v", err)
	}
	if !bytes.Equal(raw, objectBytes) {
		t.Fatalf("ObjectBinary() = % x, want % x", raw, objectBytes)
	}

	obj, err := save.Object(objectID)
	if err != nil {
		t.Fatalf("Object() error = %v", err)
	}
	if obj.Blueprint != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Object().Blueprint = %q", obj.Blueprint)
	}
	if obj.Section != "" {
		t.Fatalf("Object().Section = %q, want empty for synthetic header without sections", obj.Section)
	}
	if len(obj.Properties) != 1 || obj.Properties[0].Name != "Health" || obj.Properties[0].Type != arkproperty.TypeInt {
		t.Fatalf("Object().Properties = %#v, want Health Int property", obj.Properties)
	}

	classIDs, err := save.ObjectIDsByClassContains("/Game/Test")
	if err != nil {
		t.Fatalf("ObjectIDsByClassContains() error = %v", err)
	}
	if len(classIDs) != 1 || classIDs[0] != objectID {
		t.Fatalf("ObjectIDsByClassContains(/Game/Test) = %v, want [%s]", classIDs, objectID)
	}

	classes, err := save.Classes()
	if err != nil {
		t.Fatalf("Classes() error = %v", err)
	}
	if len(classes) != 2 || classes[0] != "Blueprint'/Game/Other.Other_C'" || classes[1] != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Classes() = %#v, want two sorted classes", classes)
	}

	infos, err := save.ObjectClassInfos()
	if err != nil {
		t.Fatalf("ObjectClassInfos() error = %v", err)
	}
	if len(infos) != 2 || infos[0].UUID != objectID || infos[0].ClassName != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("ObjectClassInfos() = %#v", infos)
	}

	parsed, err := save.ParsedObjects(nil)
	if err != nil {
		t.Fatalf("ParsedObjects(nil) error = %v", err)
	}
	if len(parsed) != 2 || parsed[0].UUID != objectID || parsed[0].Object.Blueprint != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("ParsedObjects(nil) = %#v", parsed)
	}

	filteredParsed, err := save.ParsedObjectsByClassContains("/Game/Test")
	if err != nil {
		t.Fatalf("ParsedObjectsByClassContains(/Game/Test) error = %v", err)
	}
	if len(filteredParsed) != 1 || filteredParsed[0].UUID != objectID {
		t.Fatalf("ParsedObjectsByClassContains(/Game/Test) = %#v, want [%s]", filteredParsed, objectID)
	}

	byProperty, err := save.ParsedObjectsWithAnyProperty([]string{"TamerString", "Health"})
	if err != nil {
		t.Fatalf("ParsedObjectsWithAnyProperty() error = %v", err)
	}
	if len(byProperty) != 2 || byProperty[0].UUID != objectID || byProperty[1].UUID != secondObjectID {
		t.Fatalf("ParsedObjectsWithAnyProperty(TamerString, Health) = %#v, want both synthetic objects", byProperty)
	}

	missingProperty, err := save.ParsedObjectsWithAnyProperty([]string{"TamerString"})
	if err != nil {
		t.Fatalf("ParsedObjectsWithAnyProperty(missing) error = %v", err)
	}
	if len(missingProperty) != 0 {
		t.Fatalf("ParsedObjectsWithAnyProperty(missing) = %#v, want empty", missingProperty)
	}

	classInfosByProperty, err := save.ObjectClassInfosWithAnyProperty([]string{"Health"})
	if err != nil {
		t.Fatalf("ObjectClassInfosWithAnyProperty() error = %v", err)
	}
	if len(classInfosByProperty) != 2 || classInfosByProperty[0].UUID != objectID || classInfosByProperty[1].UUID != secondObjectID {
		t.Fatalf("ObjectClassInfosWithAnyProperty(Health) = %#v, want both synthetic objects", classInfosByProperty)
	}
}

func TestObjectCacheControlsRawObjectRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	firstBytes := syntheticObjectBytes(0x10000001)
	secondBytes := syntheticObjectBytes(0x10000005)
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: syntheticHeader(),
		Objects: map[uuid.UUID][]byte{
			objectID: firstBytes,
		},
	})

	save, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	if save.ObjectCacheEnabled() {
		t.Fatalf("ObjectCacheEnabled() = true, want default false")
	}
	raw, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(first) error = %v", err)
	}
	if !bytes.Equal(raw, firstBytes) {
		t.Fatalf("ObjectBinary(first) = % x, want first bytes", raw)
	}
	if err := replaceGameRow(save, objectID, secondBytes); err != nil {
		t.Fatalf("replaceGameRow(second) error = %v", err)
	}
	raw, err = save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(second uncached) error = %v", err)
	}
	if !bytes.Equal(raw, secondBytes) {
		t.Fatalf("ObjectBinary(second uncached) = % x, want second bytes", raw)
	}

	save.SetObjectCacheEnabled(true)
	if !save.ObjectCacheEnabled() {
		t.Fatalf("ObjectCacheEnabled() = false, want true")
	}
	raw, err = save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(cache fill) error = %v", err)
	}
	if !bytes.Equal(raw, secondBytes) {
		t.Fatalf("ObjectBinary(cache fill) = % x, want second bytes", raw)
	}
	thirdBytes := syntheticObjectBytes(0x10000001)
	if err := replaceGameRow(save, objectID, thirdBytes); err != nil {
		t.Fatalf("replaceGameRow(third) error = %v", err)
	}
	raw, err = save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(cached) error = %v", err)
	}
	if !bytes.Equal(raw, secondBytes) {
		t.Fatalf("ObjectBinary(cached) = % x, want cached second bytes", raw)
	}

	save.ClearObjectCache()
	raw, err = save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(after clear) error = %v", err)
	}
	if !bytes.Equal(raw, thirdBytes) {
		t.Fatalf("ObjectBinary(after clear) = % x, want third bytes", raw)
	}

	save.SetObjectCacheEnabled(false)
	if save.ObjectCacheEnabled() {
		t.Fatalf("ObjectCacheEnabled() = true after disable")
	}
	if err := replaceGameRow(save, objectID, secondBytes); err != nil {
		t.Fatalf("replaceGameRow(second again) error = %v", err)
	}
	raw, err = save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(after disable) error = %v", err)
	}
	if !bytes.Equal(raw, secondBytes) {
		t.Fatalf("ObjectBinary(after disable) = % x, want live second bytes", raw)
	}
}

func TestObjectCacheReturnsDefensiveCopies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	objectBytes := syntheticObjectBytes(0x10000001)
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: syntheticHeader(),
		Objects: map[uuid.UUID][]byte{
			objectID: objectBytes,
		},
	})

	save, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()
	save.SetObjectCacheEnabled(true)

	raw, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(first) error = %v", err)
	}
	raw[0] = 0xff
	again, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(second) error = %v", err)
	}
	if !bytes.Equal(again, objectBytes) {
		t.Fatalf("ObjectBinary(second) = % x, want unmutated object bytes", again)
	}
}

func TestObjectCacheAllowsConcurrentReads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	objectBytes := syntheticObjectBytes(0x10000001)
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: syntheticHeader(),
		Objects: map[uuid.UUID][]byte{
			objectID: objectBytes,
		},
	})

	save, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()
	save.SetObjectCacheEnabled(true)

	if _, err := save.ObjectBinary(objectID); err != nil {
		t.Fatalf("ObjectBinary(cache fill) error = %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 64)
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				raw, err := save.ObjectBinary(objectID)
				if err != nil {
					errs <- err
					return
				}
				if !bytes.Equal(raw, objectBytes) {
					errs <- errors.New("cached object bytes changed")
					return
				}
				raw[0] = 0xff
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	raw, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(after concurrent reads) error = %v", err)
	}
	if !bytes.Equal(raw, objectBytes) {
		t.Fatalf("ObjectBinary(after concurrent reads) = % x, want unmutated object bytes", raw)
	}
}

func TestParsedObjectsWithFaultsCollectsObjectParseErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	header := syntheticHeader()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: header,
		Objects: map[uuid.UUID][]byte{
			objectID: syntheticObjectBytes(0x10000001),
			faultyID: truncatedObjectBytes(0x10000005),
		},
	})

	save, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	parsed, faults, err := save.ParsedObjectsWithFaults(nil)
	if err != nil {
		t.Fatalf("ParsedObjectsWithFaults(nil) error = %v", err)
	}
	if len(parsed) != 1 || parsed[0].UUID != objectID {
		t.Fatalf("ParsedObjectsWithFaults parsed = %#v, want object %s", parsed, objectID)
	}
	if len(faults) != 1 || faults[0].UUID != faultyID {
		t.Fatalf("ParsedObjectsWithFaults faults = %#v, want fault for %s", faults, faultyID)
	}
	if faults[0].ClassName != "Blueprint'/Game/Other.Other_C'" || faults[0].Err == nil {
		t.Fatalf("fault = %#v, want class name and parse error", faults[0])
	}
}

func TestParsedObjectsWithAnyPropertyWithFaultsKeepsValidMatchesAndReportsFaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	header := syntheticHeader()
	faultyBytes := syntheticObjectBytes(0x10000005)
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: header,
		Objects: map[uuid.UUID][]byte{
			objectID: syntheticObjectBytes(0x10000001),
			faultyID: faultyBytes[:len(faultyBytes)-2],
		},
	})

	save, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	parsed, faults, err := save.ParsedObjectsWithAnyPropertyWithFaults([]string{"Health"})
	if err != nil {
		t.Fatalf("ParsedObjectsWithAnyPropertyWithFaults() error = %v", err)
	}
	if len(parsed) != 1 || parsed[0].UUID != objectID {
		t.Fatalf("parsed = %#v, want valid object %s", parsed, objectID)
	}
	if len(faults) != 1 || faults[0].UUID != faultyID || faults[0].Err == nil {
		t.Fatalf("faults = %#v, want parse fault for %s", faults, faultyID)
	}
}

func syntheticObjectBytes(classNameID uint32) []byte {
	var buf bytes.Buffer
	testfixtures.WriteIntPropertyID(&buf, 0x10000002, 0x10000003, 250)
	return testfixtures.ObjectBytesWithProperties(classNameID, 0x10000004, buf.Bytes())
}

func truncatedObjectBytes(classNameID uint32) []byte {
	var buf bytes.Buffer
	buf.Write(testfixtures.ObjectBytesWithProperties(classNameID, 0x10000004, nil))
	buf.Truncate(buf.Len() - 10)
	return buf.Bytes()
}

func replaceGameRow(save *Save, id uuid.UUID, value []byte) error {
	_, err := save.db.Exec(`update game set value = ? where key = ?`, value, id[:])
	return err
}

func syntheticHeader() []byte {
	return testfixtures.Header("Valguero_WP", map[uint32]string{
		0x10000000: "None",
		0x10000001: "Blueprint'/Game/Test.Test_C'",
		0x10000002: "Health",
		0x10000003: "IntProperty",
		0x10000004: "None",
		0x10000005: "Blueprint'/Game/Other.Other_C'",
	})
}
