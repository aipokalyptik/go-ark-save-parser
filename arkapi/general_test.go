package arkapi

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestGeneralObjectIDsReturnsSaveObjectIDs(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewGeneral(save)
	ids, err := api.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs() error = %v", err)
	}

	wantID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	if len(ids) != 1 || ids[0] != wantID {
		t.Fatalf("ObjectIDs() = %v, want [%s]", ids, wantID)
	}
}

func TestGeneralObjectReturnsParsedSaveObject(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewGeneral(save)
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	obj, err := api.Object(id)
	if err != nil {
		t.Fatalf("Object() error = %v", err)
	}

	if obj.UUID != id {
		t.Fatalf("Object().UUID = %s, want %s", obj.UUID, id)
	}
	if obj.Blueprint != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Object().Blueprint = %q", obj.Blueprint)
	}
	if len(obj.Properties) != 1 || obj.Properties[0].Name != "Health" || obj.Properties[0].Type != arkproperty.TypeInt {
		t.Fatalf("Object().Properties = %#v, want Health Int property", obj.Properties)
	}
}

func TestGeneralObjectsReturnsParsedSaveObjects(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewGeneral(save)
	objects, err := api.Objects()
	if err != nil {
		t.Fatalf("Objects() error = %v", err)
	}

	wantID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	if len(objects) != 1 {
		t.Fatalf("Objects() length = %d, want 1", len(objects))
	}
	if objects[0].UUID != wantID {
		t.Fatalf("Objects()[0].UUID = %s, want %s", objects[0].UUID, wantID)
	}
	if objects[0].Blueprint != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Objects()[0].Blueprint = %q", objects[0].Blueprint)
	}
}

func openSyntheticSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, objectID[:], syntheticObjectBytes(0x10000001))
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func syntheticObjectBytes(classNameID uint32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, classNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000002))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000003))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(4))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, int32(250))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticHeader() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int16(12))
	nameOffsetPosition := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, float64(1234.5))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(77))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	for buf.Len() < 30 {
		buf.WriteByte(0)
	}
	writeArkString(&buf, "Valguero_WP")
	nameOffset := int32(buf.Len())
	binary.LittleEndian.PutUint32(buf.Bytes()[nameOffsetPosition:nameOffsetPosition+4], uint32(nameOffset))
	_ = binary.Write(&buf, binary.LittleEndian, int32(11))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000000))
	writeArkString(&buf, "None")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000001))
	writeArkString(&buf, "Blueprint'/Game/Test.Test_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000002))
	writeArkString(&buf, "Health")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000003))
	writeArkString(&buf, "IntProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	writeArkString(&buf, "None")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000005))
	writeArkString(&buf, "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000006))
	writeArkString(&buf, "StructureID")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000007))
	writeArkString(&buf, "MaxHealth")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000008))
	writeArkString(&buf, "Health")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000009))
	writeArkString(&buf, "TargetingTeam")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000a))
	writeArkString(&buf, "FloatProperty")
	return buf.Bytes()
}

func writeArkString(buf *bytes.Buffer, s string) {
	if s == "" {
		_ = binary.Write(buf, binary.LittleEndian, int32(0))
		return
	}
	_ = binary.Write(buf, binary.LittleEndian, int32(len(s)+1))
	buf.WriteString(s)
	buf.WriteByte(0)
}
