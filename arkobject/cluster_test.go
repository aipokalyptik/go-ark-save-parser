package arkobject

import "testing"

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
