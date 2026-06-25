package arkapi

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func BenchmarkOpenSave(b *testing.B) {
	path := createBenchmarkSave(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		save, err := arksave.Open(path)
		if err != nil {
			b.Fatal(err)
		}
		if err := save.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkObjectEnumeration(b *testing.B) {
	save := openBenchmarkSave(b)
	defer save.Close()
	api := NewGeneral(save)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := api.ObjectIDs(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkObjectParse(b *testing.B) {
	save := openBenchmarkSave(b)
	defer save.Close()
	api := NewGeneral(save)
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := api.Object(id); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkObjectParseWithObjectCache(b *testing.B) {
	save := openBenchmarkSave(b)
	defer save.Close()
	save.SetObjectCacheEnabled(true)
	api := NewGeneral(save)
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	if _, err := api.Object(id); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := api.Object(id); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFilteredAPICalls(b *testing.B) {
	save := openBenchmarkSave(b)
	defer save.Close()
	structures := NewStructure(save)
	stackables := NewStackable(save)
	equipment := NewEquipment(save)
	dinos := NewDino(save)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := structures.All(); err != nil {
			b.Fatal(err)
		}
		if _, err := stackables.All(); err != nil {
			b.Fatal(err)
		}
		if _, err := equipment.All(); err != nil {
			b.Fatal(err)
		}
		if _, err := dinos.All(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONExport(b *testing.B) {
	save := openBenchmarkSave(b)
	defer save.Close()
	api := NewJSON(save)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := api.ExportSaveInfoJSON(); err != nil {
			b.Fatal(err)
		}
	}
}

func openBenchmarkSave(b *testing.B) *arksave.Save {
	b.Helper()
	save, err := arksave.Open(createBenchmarkSave(b))
	if err != nil {
		b.Fatal(err)
	}
	return save
}

func createBenchmarkSave(b *testing.B) string {
	b.Helper()
	path := filepath.Join(b.TempDir(), "benchmark.ark")

	genericID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	structureID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	stackableID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	equipmentID := uuid.MustParse("33332233-4455-6677-8899-aabbccddeeff")
	dinoID := uuid.MustParse("44442233-4455-6677-8899-aabbccddeeff")

	testfixtures.WriteSave(b, path, testfixtures.SaveOptions{
		Header: syntheticHeader(),
		Custom: map[string][]byte{
			"ActorTransforms": benchmarkActorTransforms(structureID, dinoID),
		},
		Objects: map[uuid.UUID][]byte{
			genericID:   syntheticObjectBytes(0x10000001),
			structureID: syntheticStructureObjectBytes(),
			stackableID: syntheticStackableObjectBytes(false),
			equipmentID: syntheticEquipmentObjectBytes(false),
			dinoID:      syntheticDinoObjectBytes(),
		},
	})
	return path
}

func benchmarkActorTransforms(ids ...uuid.UUID) []byte {
	var buf bytes.Buffer
	for _, id := range ids {
		buf.Write(id[:])
		for _, value := range []float64{11, 22, 33, 0, 0, 0, 1} {
			_ = binary.Write(&buf, binary.LittleEndian, value)
		}
	}
	buf.Write(uuid.Nil[:])
	return buf.Bytes()
}
