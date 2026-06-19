package arksave

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestOpenReadsHeaderCustomValuesAndGameObjects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	objectBytes := syntheticObjectBytes(0x10000001)
	header := syntheticHeader()
	actorTransforms := append(objectID[:], []byte{1, 2, 3, 4}...)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", header)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "ActorTransforms", actorTransforms)
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, objectID[:], objectBytes)
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

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

	ids, err := save.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != objectID {
		t.Fatalf("ObjectIDs() = %v, want [%s]", ids, objectID)
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
	writeArkString(&buf, "")
	for buf.Len() < 30 {
		buf.WriteByte(0)
	}
	writeArkString(&buf, "Valguero_WP")
	nameOffset := int32(buf.Len())
	binary.LittleEndian.PutUint32(buf.Bytes()[nameOffsetPosition:nameOffsetPosition+4], uint32(nameOffset))
	_ = binary.Write(&buf, binary.LittleEndian, int32(5))
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
