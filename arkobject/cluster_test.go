package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestClusterItemShortNameUsesBlueprintClass(t *testing.T) {
	item := ClusterItem{Blueprint: "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C"}

	if got := item.ShortName(); got != "WeaponBow" {
		t.Fatalf("ShortName() = %q, want WeaponBow", got)
	}
}

func TestClusterDinoPrimaryClassAndShortNameSkipComponents(t *testing.T) {
	dino := ClusterDino{ClassNames: []string{
		"/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C",
		"/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
		"/Game/PrimalEarth/Dinos/Raptor/Raptor_AIController_BP.Raptor_AIController_BP_C",
	}}

	if got := dino.PrimaryClassName(); got != "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C" {
		t.Fatalf("PrimaryClassName() = %q, want raptor class", got)
	}
	if got := dino.ShortName(); got != "Raptor" {
		t.Fatalf("ShortName() = %q, want Raptor", got)
	}
}

func TestClusterDinoPrimaryClassNameFallsBackToFirstClass(t *testing.T) {
	dino := ClusterDino{ClassNames: []string{"/Game/Test/OnlyComponent_BP.OnlyComponent_BP_C"}}

	if got := dino.PrimaryClassName(); got != "/Game/Test/OnlyComponent_BP.OnlyComponent_BP_C" {
		t.Fatalf("PrimaryClassName() = %q, want first class fallback", got)
	}
}

func TestClusterDinoFromUploadProjectsEmbeddedDinoIdentityAndStats(t *testing.T) {
	dinoID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	statusID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	archive := &arkarchive.Archive{Objects: []arkarchive.Object{
		{
			UUID:      dinoID,
			ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
			Properties: []arkproperty.Property{
				{Name: "DinoID1", Type: arkproperty.TypeUInt32, Value: uint32(1001)},
				{Name: "DinoID2", Type: arkproperty.TypeUInt32, Value: uint32(2002)},
				{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)},
				{Name: "TamedName", Type: arkproperty.TypeString, Value: "Needle"},
				{Name: "bIsFemale", Type: arkproperty.TypeBool, Value: true},
				{Name: "MyCharacterStatusComponent", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{Type: arkproperty.ObjectReferenceUUID, Value: statusID}},
			},
		},
		{
			UUID:      statusID,
			ClassName: "/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C",
			Properties: []arkproperty.Property{
				{Name: "BaseCharacterLevel", Type: arkproperty.TypeInt, Value: int32(12)},
				{Name: "NumberOfLevelUpPointsApplied", Type: arkproperty.TypeInt, Position: 0, Value: int32(5)},
				{Name: "NumberOfLevelUpPointsApplied", Type: arkproperty.TypeInt, Position: 8, Value: int32(2)},
				{Name: "NumberOfLevelUpPointsAppliedTamed", Type: arkproperty.TypeInt, Position: 8, Value: int32(1)},
			},
		},
	}}

	got := ClusterDinoFromUpload(arkcluster.Dino{Index: 3, Version: 7, Archive: archive}, []string{
		archive.Objects[0].ClassName,
		archive.Objects[1].ClassName,
	})

	if got.DinoID1 != 1001 || got.DinoID2 != 2002 || !got.IsTamed || got.TamedName != "Needle" || !got.IsFemale {
		t.Fatalf("embedded dino identity = %#v, want parsed ID/name/tamed/female fields", got)
	}
	if !got.HasStats || got.BaseLevel != 12 || got.CurrentLevel != 9 {
		t.Fatalf("embedded dino stats = has=%v base=%d current=%d, want true/12/9", got.HasStats, got.BaseLevel, got.CurrentLevel)
	}
}
