package arkobject

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestParseGameObjectReadsHeaderNamesSectionAndProperties(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Blueprint'/Game/Test.Test_C'",
		2: "InstanceName_123",
		3: "Health",
		4: "IntProperty",
		5: "None",
	})
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeUInt32(stream, 0)
	writeInt32(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 0)
	_ = binary.Write(stream, binary.LittleEndian, int16(9))
	writeName(stream, 3)
	writeName(stream, 4)
	writeInt32(stream, 4)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeInt32(stream, 250)
	writeName(stream, 5)

	obj, err := ParseGameObject(id, stream.Bytes(), ctx, []string{"PersistentLevel"})
	if err != nil {
		t.Fatalf("ParseGameObject() error = %v", err)
	}
	if obj.UUID != id {
		t.Fatalf("UUID = %s, want %s", obj.UUID, id)
	}
	if obj.Blueprint != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Blueprint = %q", obj.Blueprint)
	}
	if len(obj.Names) != 1 || obj.Names[0] != "InstanceName_123" {
		t.Fatalf("Names = %#v, want InstanceName_123", obj.Names)
	}
	if obj.Section != "PersistentLevel" {
		t.Fatalf("Section = %q, want PersistentLevel", obj.Section)
	}
	if obj.Unknown != 9 {
		t.Fatalf("Unknown = %d, want 9", obj.Unknown)
	}
	if len(obj.Properties) != 1 {
		t.Fatalf("Properties length = %d, want 1", len(obj.Properties))
	}
	prop := obj.Properties[0]
	if prop.Name != "Health" || prop.Type != arkproperty.TypeInt || prop.Value != int32(250) {
		t.Fatalf("property = %#v, want Health Int 250", prop)
	}
}

func TestParseGameObjectReadsStringEncodedObjectNames(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Blueprint'/Game/Test.Test_C'",
		3: "Health",
		4: "IntProperty",
		5: "None",
	})
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeUInt32(stream, 0)
	writeInt32(stream, 1)
	testfixtures.WriteArkString(stream, "RuntimeObjectName_123")
	writeInt32(stream, 0)
	_ = binary.Write(stream, binary.LittleEndian, int16(9))
	writeName(stream, 3)
	writeName(stream, 4)
	writeInt32(stream, 4)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeInt32(stream, 250)
	writeName(stream, 5)

	obj, err := ParseGameObject(id, stream.Bytes(), ctx, []string{"PersistentLevel"})
	if err != nil {
		t.Fatalf("ParseGameObject() error = %v", err)
	}
	if len(obj.Names) != 1 || obj.Names[0] != "RuntimeObjectName_123" {
		t.Fatalf("Names = %#v, want RuntimeObjectName_123", obj.Names)
	}
	if obj.Section != "PersistentLevel" || obj.Unknown != 9 {
		t.Fatalf("header fields = section %q unknown %d", obj.Section, obj.Unknown)
	}
}

func TestGameObjectShortNameFollowsUpstreamBlueprintRules(t *testing.T) {
	tests := []struct {
		name      string
		blueprint string
		want      string
	}{
		{
			name:      "dino",
			blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			want:      "Raptor",
		},
		{
			name:      "resource",
			blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
			want:      "Resource_Stone",
		},
		{
			name:      "structure",
			blueprint: "Blueprint'/Game/Structures/Stone/PrimalItemStructure_Wall_Stone.PrimalItemStructure_Wall_Stone_C'",
			want:      "Wall_Stone",
		},
		{
			name:      "status",
			blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C'",
			want:      "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object := GameObject{Blueprint: tt.blueprint}
			if got := object.ShortName(); got != tt.want {
				t.Fatalf("ShortName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func writeName(buf *bytes.Buffer, id uint32) {
	writeUInt32(buf, id)
	writeInt32(buf, 0)
}

func writeUInt32(buf *bytes.Buffer, value uint32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeInt32(buf *bytes.Buffer, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}
