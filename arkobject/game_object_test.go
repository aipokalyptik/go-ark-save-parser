package arkobject

import (
	"bytes"
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
	raw := testfixtures.ObjectBytesWithNamePayload(1, nameIDPayload(2), 9, gameObjectIntPropertyPayload(3, 4, 250), 5)

	obj, err := ParseGameObject(id, raw, ctx, []string{"PersistentLevel"})
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
	raw := testfixtures.ObjectBytesWithNamePayload(1, nameStringPayload("RuntimeObjectName_123"), 9, gameObjectIntPropertyPayload(3, 4, 250), 5)

	obj, err := ParseGameObject(id, raw, ctx, []string{"PersistentLevel"})
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

func TestParseGameObjectPartialKeepsPropertiesBeforePropertyError(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Blueprint'/Game/Test.Test_C'",
		2: "InstanceName_123",
		3: "Health",
		4: "IntProperty",
		5: "Owner",
		6: "ObjectProperty",
	})
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var props bytes.Buffer
	props.Write(gameObjectIntPropertyPayload(3, 4, 250))
	testfixtures.WriteNameID(&props, 5)
	testfixtures.WriteNameID(&props, 6)
	testfixtures.WriteInt32(&props, 2)
	testfixtures.WriteInt32(&props, 0)
	props.WriteByte(0)
	testfixtures.WriteInt16(&props, 5)
	raw := testfixtures.ObjectBytesWithNamePayload(1, nameIDPayload(2), 9, props.Bytes(), 0)

	obj, err := ParseGameObjectPartial(id, raw, ctx, []string{"PersistentLevel"})
	if err == nil {
		t.Fatalf("ParseGameObjectPartial() error = nil, want recorded property error")
	}
	if obj == nil {
		t.Fatalf("ParseGameObjectPartial() object = nil, want partial object")
	}
	if len(obj.Properties) != 1 {
		t.Fatalf("Properties length = %d, want 1", len(obj.Properties))
	}
	if obj.Properties[0].Name != "Health" || obj.Properties[0].Value != int32(250) {
		t.Fatalf("partial property = %#v, want Health 250", obj.Properties[0])
	}
}

func nameIDPayload(nameID uint32) []byte {
	var buf bytes.Buffer
	testfixtures.WriteNameID(&buf, nameID)
	return buf.Bytes()
}

func nameStringPayload(name string) []byte {
	var buf bytes.Buffer
	testfixtures.WriteArkString(&buf, name)
	return buf.Bytes()
}

func gameObjectIntPropertyPayload(nameID uint32, propertyTypeID uint32, value int32) []byte {
	var buf bytes.Buffer
	testfixtures.WriteIntPropertyID(&buf, nameID, propertyTypeID, value)
	return buf.Bytes()
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
