package arkapi

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestEquipmentAPIClassifiesBlueprints(t *testing.T) {
	api := EquipmentAPI{}
	for blueprint, want := range map[string]arkobject.EquipmentKind{
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'":                         arkobject.EquipmentWeapon,
		"Blueprint'/Game/ScorchedEarth/WeaponChainsaw/PrimalItem_ChainSaw.PrimalItem_ChainSaw_C'":                                 arkobject.EquipmentWeapon,
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_Saddle.PrimalItemArmor_Saddle_C'":         arkobject.EquipmentSaddle,
		"Blueprint'/Game/ASA/Dinos/YiLing/PrimalItemArmor_YiLingSaddle.PrimalItemArmor_YiLingSaddle_C'":                           arkobject.EquipmentSaddle,
		"Blueprint'/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C'":       arkobject.EquipmentSaddle,
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'":   arkobject.EquipmentArmor,
		"Blueprint'/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothShirt.PrimalItemArmor_DesertClothShirt_C'":              arkobject.EquipmentArmor,
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_WoodShield.PrimalItemArmor_WoodShield_C'": arkobject.EquipmentShield,
	} {
		got := api.KindForBlueprint(blueprint)
		if got != want {
			t.Fatalf("KindForBlueprint(%q) = %s, want %s", blueprint, got, want)
		}
	}
	if api.IsApplicableBlueprint("Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItemAmmo_ArrowStone.PrimalItemAmmo_ArrowStone_C'") {
		t.Fatalf("IsApplicableBlueprint(ammo) = true, want false")
	}
}

func sortedStrings(values []string) bool {
	for i := 1; i < len(values); i++ {
		if values[i-1] > values[i] {
			return false
		}
	}
	return true
}

func stringSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func TestUpstreamEquipmentBlueprintListsExposeCanonicalClasses(t *testing.T) {
	weapons := UpstreamWeaponBlueprints()
	armor := UpstreamArmorBlueprints()
	shields := UpstreamShieldBlueprints()
	if len(weapons) != 45 {
		t.Fatalf("UpstreamWeaponBlueprints() length = %d, want 45", len(weapons))
	}
	if len(armor) != 77 {
		t.Fatalf("UpstreamArmorBlueprints() length = %d, want 77", len(armor))
	}
	if len(shields) != 6 {
		t.Fatalf("UpstreamShieldBlueprints() length = %d, want 6", len(shields))
	}
	if !sortedStrings(weapons) || !sortedStrings(armor) || !sortedStrings(shields) {
		t.Fatalf("upstream equipment blueprint lists must be sorted")
	}
	if _, ok := stringSet(weapons)["/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponShotgun.PrimalItem_WeaponShotgun_C"]; !ok {
		t.Fatalf("UpstreamWeaponBlueprints() missing shotgun")
	}
	if _, ok := stringSet(armor)["/Game/PrimalEarth/CoreBlueprints/Items/Armor/Metal/PrimalItemArmor_MetalShirt.PrimalItemArmor_MetalShirt_C"]; !ok {
		t.Fatalf("UpstreamArmorBlueprints() missing metal shirt")
	}
	if _, ok := stringSet(shields)["/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_WoodShield.PrimalItemArmor_WoodShield_C"]; !ok {
		t.Fatalf("UpstreamShieldBlueprints() missing wood shield")
	}
}

func TestEquipmentAPIAllAndByKindReadLocalSaveItems(t *testing.T) {
	save := openSyntheticEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	items, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("All() length = %d, want 1", len(items))
	}
	for _, item := range items {
		if item.Kind != arkobject.EquipmentWeapon || item.Rating != 7.5 || item.Quality != 3 {
			t.Fatalf("Equipment item = %#v", item)
		}
		if item.Stats.Internal[arkobject.EquipmentStatDamage] != 1234 || item.Stats.Damage != 112.3 {
			t.Fatalf("Equipment stats = %#v", item.Stats)
		}
	}
	weapons, err := api.ByKind(arkobject.EquipmentWeapon)
	if err != nil {
		t.Fatalf("ByKind(weapon) error = %v", err)
	}
	if len(weapons) != 1 {
		t.Fatalf("ByKind(weapon) length = %d, want 1", len(weapons))
	}
	weapons, err = api.Weapons()
	if err != nil {
		t.Fatalf("Weapons() error = %v", err)
	}
	if len(weapons) != 1 {
		t.Fatalf("Weapons() length = %d, want 1", len(weapons))
	}
}

func TestEquipmentAPIExportBinaryWritesEquipmentRows(t *testing.T) {
	save := openSyntheticEquipmentSave(t)
	defer save.Close()

	itemID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	outDir := filepath.Join(t.TempDir(), "equipment-export")
	exported, err := NewEquipment(save).ExportBinary(outDir)
	if err != nil {
		t.Fatalf("ExportBinary() error = %v", err)
	}
	if exported.ItemCount != 1 || exported.RowCount != 1 || exported.FaultCount != 0 {
		t.Fatalf("ExportBinary() = %#v, want one equipment row", exported)
	}
	got, err := os.ReadFile(filepath.Join(outDir, "item_"+itemID.String()+".bin"))
	if err != nil {
		t.Fatalf("read exported item row: %v", err)
	}
	want, err := save.ObjectBinary(itemID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", itemID, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("exported item row differs from save row")
	}
	if _, err := os.Stat(filepath.Join(outDir, "manifest.json")); err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
}

func TestEquipmentAPIByKindClassFiltersBlueprintsWithinKind(t *testing.T) {
	save := openSyntheticMixedEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	weaponBlueprint := "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'"
	armorBlueprint := "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'"
	weapons, err := api.ByKindClass(arkobject.EquipmentWeapon, []string{weaponBlueprint, armorBlueprint})
	if err != nil {
		t.Fatalf("ByKindClass(weapon) error = %v", err)
	}
	if len(weapons) != 1 {
		t.Fatalf("ByKindClass(weapon) length = %d, want 1", len(weapons))
	}
	for _, item := range weapons {
		if item.Kind != arkobject.EquipmentWeapon || item.Blueprint != weaponBlueprint {
			t.Fatalf("ByKindClass(weapon) item = %#v", item)
		}
	}

	armor, err := api.ByKindClass(arkobject.EquipmentArmor, []string{weaponBlueprint, armorBlueprint})
	if err != nil {
		t.Fatalf("ByKindClass(armor) error = %v", err)
	}
	if len(armor) != 1 {
		t.Fatalf("ByKindClass(armor) length = %d, want 1", len(armor))
	}
	for _, item := range armor {
		if item.Kind != arkobject.EquipmentArmor || item.Blueprint != armorBlueprint {
			t.Fatalf("ByKindClass(armor) item = %#v", item)
		}
	}
}

func TestEquipmentAPIAllMatchingBlueprintsComposesClassFilterBeforeParsing(t *testing.T) {
	save := openSyntheticMixedEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	weapons, err := api.AllMatchingBlueprints(func(blueprint string) bool {
		return strings.Contains(blueprint, "/Weapons/")
	})
	if err != nil {
		t.Fatalf("AllMatchingBlueprints(weapons) error = %v", err)
	}
	if len(weapons) != 1 {
		t.Fatalf("AllMatchingBlueprints(weapons) length = %d, want 1", len(weapons))
	}
	for _, item := range weapons {
		if item.Kind != arkobject.EquipmentWeapon {
			t.Fatalf("AllMatchingBlueprints(weapons) item = %#v", item)
		}
	}

	armor, err := api.AllMatchingBlueprints(func(blueprint string) bool {
		return strings.Contains(blueprint, "/Armor/Cloth/")
	})
	if err != nil {
		t.Fatalf("AllMatchingBlueprints(armor) error = %v", err)
	}
	if len(armor) != 1 {
		t.Fatalf("AllMatchingBlueprints(armor) length = %d, want 1", len(armor))
	}
	for _, item := range armor {
		if item.Kind != arkobject.EquipmentArmor {
			t.Fatalf("AllMatchingBlueprints(armor) item = %#v", item)
		}
	}
}

func TestEquipmentAPIAllWithFaultsKeepsValidItemsAndReportsParseFaults(t *testing.T) {
	save := openSyntheticEquipmentSaveWithFault(t)
	defer save.Close()

	api := NewEquipment(save)
	items, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("AllWithFaults() items length = %d, want 1", len(items))
	}
	for _, item := range items {
		if item.Kind != arkobject.EquipmentWeapon || item.Rating != 7.5 || item.Quality != 3 {
			t.Fatalf("Equipment item = %#v", item)
		}
	}
	if len(faults) != 1 || faults[0].ClassName != "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'" || faults[0].Err == nil {
		t.Fatalf("AllWithFaults() faults = %#v, want one equipment parse fault", faults)
	}
}

func TestEquipmentAPIFilteredWithFaultsKeepsValidFilteredItemsAndReportsParseFaults(t *testing.T) {
	save := openSyntheticEquipmentSaveWithFault(t)
	defer save.Close()

	api := NewEquipment(save)
	items, faults, err := api.FilteredWithFaults(EquipmentFilterOptions{
		Kinds:        []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		NoBlueprints: true,
	})
	if err != nil {
		t.Fatalf("FilteredWithFaults() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("FilteredWithFaults() items length = %d, want 1", len(items))
	}
	for _, item := range items {
		if item.Kind != arkobject.EquipmentWeapon || item.IsBlueprint {
			t.Fatalf("FilteredWithFaults() item = %#v, want non-blueprint weapon", item)
		}
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("FilteredWithFaults() faults = %#v, want one parse fault", faults)
	}

	_, _, err = api.FilteredWithFaults(EquipmentFilterOptions{NoBlueprints: true, OnlyBlueprints: true})
	if err == nil {
		t.Fatalf("FilteredWithFaults(conflicting blueprint filters) error = nil, want error")
	}
}

func TestEquipmentAPIReadsArmorStatValues(t *testing.T) {
	save := openSyntheticArmorEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	armor, err := api.ByKind(arkobject.EquipmentArmor)
	if err != nil {
		t.Fatalf("ByKind(armor) error = %v", err)
	}
	if len(armor) != 1 {
		t.Fatalf("ByKind(armor) length = %d, want 1", len(armor))
	}
	armor, err = api.Armor()
	if err != nil {
		t.Fatalf("Armor() error = %v", err)
	}
	if len(armor) != 1 {
		t.Fatalf("Armor() length = %d, want 1", len(armor))
	}
	saddles, err := api.Saddles()
	if err != nil {
		t.Fatalf("Saddles() error = %v", err)
	}
	if len(saddles) != 0 {
		t.Fatalf("Saddles() length = %d, want 0", len(saddles))
	}
	shields, err := api.Shields()
	if err != nil {
		t.Fatalf("Shields() error = %v", err)
	}
	if len(shields) != 0 {
		t.Fatalf("Shields() length = %d, want 0", len(shields))
	}
	for _, item := range armor {
		if item.Stats.Armor != 12 || item.Stats.HypothermalResistance != 8.8 || item.Stats.HyperthermalResistance != 15.6 {
			t.Fatalf("Armor equipment stats = %#v", item.Stats)
		}
	}
}

func TestEquipmentAPICountsQuantities(t *testing.T) {
	api := NewEquipment(nil)
	items := map[uuid.UUID]arkobject.EquipmentItem{
		uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): {
			InventoryItem: arkobject.InventoryItem{Quantity: 2},
		},
		uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): {
			InventoryItem: arkobject.InventoryItem{Quantity: 3},
		},
	}
	if got := api.Count(items); got != 5 {
		t.Fatalf("Count() = %d, want 5", got)
	}
}

func TestEquipmentAPIFiltersByParsedEquipmentStats(t *testing.T) {
	weaponSave := openSyntheticEquipmentSave(t)
	defer weaponSave.Close()

	weaponAPI := NewEquipment(weaponSave)
	highDamage, err := weaponAPI.WithMinDamage(112)
	if err != nil {
		t.Fatalf("WithMinDamage() error = %v", err)
	}
	if len(highDamage) != 1 {
		t.Fatalf("WithMinDamage(112) length = %d, want 1", len(highDamage))
	}
	tooMuchDamage, err := weaponAPI.WithMinDamage(113)
	if err != nil {
		t.Fatalf("WithMinDamage(113) error = %v", err)
	}
	if len(tooMuchDamage) != 0 {
		t.Fatalf("WithMinDamage(113) length = %d, want 0", len(tooMuchDamage))
	}
	actualDurability, err := weaponAPI.WithMinActualDurability(62.5)
	if err != nil {
		t.Fatalf("WithMinActualDurability() error = %v", err)
	}
	if len(actualDurability) != 1 {
		t.Fatalf("WithMinActualDurability(62.5) length = %d, want 1", len(actualDurability))
	}
	tooMuchActualDurability, err := weaponAPI.WithMinActualDurability(62.6)
	if err != nil {
		t.Fatalf("WithMinActualDurability(62.6) error = %v", err)
	}
	if len(tooMuchActualDurability) != 0 {
		t.Fatalf("WithMinActualDurability(62.6) length = %d, want 0", len(tooMuchActualDurability))
	}

	armorSave := openSyntheticArmorEquipmentSave(t)
	defer armorSave.Close()

	armorAPI := NewEquipment(armorSave)
	armored, err := armorAPI.WithMinArmor(12)
	if err != nil {
		t.Fatalf("WithMinArmor() error = %v", err)
	}
	if len(armored) != 1 {
		t.Fatalf("WithMinArmor(12) length = %d, want 1", len(armored))
	}
	tooMuchArmor, err := armorAPI.WithMinArmor(13)
	if err != nil {
		t.Fatalf("WithMinArmor(13) error = %v", err)
	}
	if len(tooMuchArmor) != 0 {
		t.Fatalf("WithMinArmor(13) length = %d, want 0", len(tooMuchArmor))
	}
	cold, err := armorAPI.WithMinHypothermalResistance(8.8)
	if err != nil {
		t.Fatalf("WithMinHypothermalResistance() error = %v", err)
	}
	if len(cold) != 1 {
		t.Fatalf("WithMinHypothermalResistance() length = %d, want 1", len(cold))
	}
	heat, err := armorAPI.WithMinHyperthermalResistance(15.6)
	if err != nil {
		t.Fatalf("WithMinHyperthermalResistance() error = %v", err)
	}
	if len(heat) != 1 {
		t.Fatalf("WithMinHyperthermalResistance() length = %d, want 1", len(heat))
	}
}

func TestEquipmentAPIByCrafterFiltersLocalSaveItems(t *testing.T) {
	save := openSyntheticEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	items, err := api.ByCrafter(arkobject.ObjectCrafter{CharacterName: "Survivor", TribeName: "Porters"})
	if err != nil {
		t.Fatalf("ByCrafter() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ByCrafter() length = %d, want 1", len(items))
	}
	items, err = api.ByCrafter(arkobject.ObjectCrafter{CharacterName: "Other", TribeName: "Porters"})
	if err != nil {
		t.Fatalf("ByCrafter(other) error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("ByCrafter(other) length = %d, want 0", len(items))
	}
}

func TestEquipmentAPIFilterOwnedByFindsItemsThroughOwnerInventory(t *testing.T) {
	save := openSyntheticEquipmentOwnedByStructureSave(t)
	defer save.Close()

	api := NewEquipment(save)
	weapons, err := api.Weapons()
	if err != nil {
		t.Fatalf("Weapons() error = %v", err)
	}
	owned, err := api.FilterOwnedBy(weapons, arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("FilterOwnedBy() error = %v", err)
	}
	if len(owned) != 1 || api.Count(owned) != 1 {
		t.Fatalf("FilterOwnedBy() = len %d count %d, want one owned weapon", len(owned), api.Count(owned))
	}
	for _, item := range owned {
		if item.OwnerInventory == nil || *item.OwnerInventory != uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff") {
			t.Fatalf("owned item owner inventory = %v", item.OwnerInventory)
		}
	}

	allOwned, err := api.OwnedBy(arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("OwnedBy() error = %v", err)
	}
	if len(allOwned) != 1 || api.Count(allOwned) != 1 {
		t.Fatalf("OwnedBy() = len %d count %d, want one owned weapon", len(allOwned), api.Count(allOwned))
	}
}

func TestEquipmentAPIFilterOwnedByIgnoresMalformedUnrelatedContainers(t *testing.T) {
	save := openSyntheticEquipmentOwnedByStructureSaveWithFault(t)
	defer save.Close()

	api := NewEquipment(save)
	weapons, err := api.Weapons()
	if err != nil {
		t.Fatalf("Weapons() error = %v", err)
	}
	owned, err := api.FilterOwnedBy(weapons, arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("FilterOwnedBy() error = %v", err)
	}
	if len(owned) != 1 || api.Count(owned) != 1 {
		t.Fatalf("FilterOwnedBy() = len %d count %d, want one owned weapon", len(owned), api.Count(owned))
	}
}

func TestEquipmentAPIFiltersByStateAndStats(t *testing.T) {
	save := openSyntheticEquipmentFilterSave(t)
	defer save.Close()

	api := NewEquipment(save)
	equipped, err := api.Equipped()
	if err != nil {
		t.Fatalf("Equipped() error = %v", err)
	}
	if len(equipped) != 1 {
		t.Fatalf("Equipped() length = %d, want 1", len(equipped))
	}
	blueprints, err := api.Blueprints()
	if err != nil {
		t.Fatalf("Blueprints() error = %v", err)
	}
	if len(blueprints) != 1 {
		t.Fatalf("Blueprints() length = %d, want 1", len(blueprints))
	}
	byQuality, err := api.ByQuality(5)
	if err != nil {
		t.Fatalf("ByQuality() error = %v", err)
	}
	if len(byQuality) != 1 {
		t.Fatalf("ByQuality() length = %d, want 1", len(byQuality))
	}
	rated, err := api.WithMinRating(7)
	if err != nil {
		t.Fatalf("WithMinRating() error = %v", err)
	}
	if len(rated) != 1 {
		t.Fatalf("WithMinRating() length = %d, want 1", len(rated))
	}
	durable, err := api.WithMinDurability(0.5)
	if err != nil {
		t.Fatalf("WithMinDurability() error = %v", err)
	}
	if len(durable) != 1 {
		t.Fatalf("WithMinDurability() length = %d, want 1", len(durable))
	}
}

func TestEquipmentAPIFilteredCombinesUpstreamReadOnlyCriteria(t *testing.T) {
	save := openSyntheticEquipmentFilterSave(t)
	defer save.Close()

	api := NewEquipment(save)
	weaponBlueprint := "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'"
	items, err := api.Filtered(EquipmentFilterOptions{
		Kinds:         []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		Blueprints:    []string{weaponBlueprint},
		NoBlueprints:  true,
		MinQuality:    5,
		MinRating:     7,
		MinDurability: 0.5,
		Equipped:      boolPtr(true),
	})
	if err != nil {
		t.Fatalf("Filtered() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Filtered() length = %d, want 1", len(items))
	}
	for _, item := range items {
		if item.Kind != arkobject.EquipmentWeapon || item.Blueprint != weaponBlueprint || item.IsBlueprint || !item.IsEquipped {
			t.Fatalf("Filtered() item = %#v", item)
		}
	}

	blueprints, err := api.Filtered(EquipmentFilterOptions{OnlyBlueprints: true})
	if err != nil {
		t.Fatalf("Filtered(blueprints) error = %v", err)
	}
	if len(blueprints) != 1 {
		t.Fatalf("Filtered(blueprints) length = %d, want 1", len(blueprints))
	}
	for _, item := range blueprints {
		if !item.IsBlueprint {
			t.Fatalf("Filtered(blueprints) item = %#v", item)
		}
	}

	_, err = api.Filtered(EquipmentFilterOptions{NoBlueprints: true, OnlyBlueprints: true})
	if err == nil {
		t.Fatalf("Filtered(conflicting blueprint filters) error = nil, want error")
	}
}

func TestEquipmentAPITopItemHelpersMirrorExampleSelections(t *testing.T) {
	api := EquipmentAPI{}
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	thirdID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	items := map[uuid.UUID]arkobject.EquipmentItem{
		firstID: {
			Kind:  arkobject.EquipmentWeapon,
			Stats: arkobject.EquipmentStats{Damage: 112.3, Durability: 20},
		},
		secondID: {
			Kind:  arkobject.EquipmentWeapon,
			Stats: arkobject.EquipmentStats{Damage: 175.0, Durability: 10},
		},
		thirdID: {
			Kind:  arkobject.EquipmentArmor,
			Stats: arkobject.EquipmentStats{Armor: 45, Durability: 50},
		},
	}

	id, item, ok := api.BestWeaponDamage(items)
	if !ok {
		t.Fatalf("BestWeaponDamage() ok = false, want true")
	}
	if id != secondID || item.Stats.Damage != 175 {
		t.Fatalf("BestWeaponDamage() = %s %#v, want second weapon", id, item)
	}

	id, item, ok = api.BestActualDurability(items)
	if !ok {
		t.Fatalf("BestActualDurability() ok = false, want true")
	}
	if id != thirdID || item.Stats.Durability != 50 {
		t.Fatalf("BestActualDurability() = %s %#v, want armor with durability 50", id, item)
	}
}

func TestEquipmentAPIFilterAscendantWeaponBlueprints(t *testing.T) {
	api := EquipmentAPI{}
	ascendantID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	primitiveID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	nonBlueprintID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	armorID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	items := map[uuid.UUID]arkobject.EquipmentItem{
		ascendantID: {
			Kind:        arkobject.EquipmentWeapon,
			IsBlueprint: true,
			Quality:     5,
		},
		primitiveID: {
			Kind:        arkobject.EquipmentWeapon,
			IsBlueprint: true,
			Quality:     0,
		},
		nonBlueprintID: {
			Kind:        arkobject.EquipmentWeapon,
			IsBlueprint: false,
			Quality:     5,
		},
		armorID: {
			Kind:        arkobject.EquipmentArmor,
			IsBlueprint: true,
			Quality:     5,
		},
	}

	filtered := api.FilterAscendantWeaponBlueprints(items)
	if len(filtered) != 1 {
		t.Fatalf("FilterAscendantWeaponBlueprints() length = %d, want 1", len(filtered))
	}
	if _, ok := filtered[ascendantID]; !ok {
		t.Fatalf("FilterAscendantWeaponBlueprints() missing ascendant weapon blueprint: %#v", filtered)
	}
}

func openSyntheticEquipmentSave(t *testing.T) *arksave.Save {
	t.Helper()

	weaponID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	engramID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		weaponID: syntheticEquipmentObjectBytes(false),
		engramID: syntheticEquipmentObjectBytes(true),
	})
}

func openSyntheticEquipmentSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	weaponID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		weaponID: syntheticEquipmentObjectBytes(false),
		faultyID: truncatedEquipmentObjectBytes(),
	})
}

func openSyntheticEquipmentFilterSave(t *testing.T) *arksave.Save {
	t.Helper()

	equippedID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	blueprintID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		equippedID:  syntheticEquipmentObjectBytesWithFlags(false, true, false, 7.5, 5, 0.75),
		blueprintID: syntheticEquipmentObjectBytesWithFlags(false, false, true, 2.5, 1, 0.25),
	})
}

func openSyntheticMixedEquipmentSave(t *testing.T) *arksave.Save {
	t.Helper()

	weaponID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	armorID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		weaponID: syntheticEquipmentObjectBytes(false),
		armorID:  syntheticArmorEquipmentObjectBytes(),
	})
}

func openSyntheticArmorEquipmentSave(t *testing.T) *arksave.Save {
	t.Helper()

	armorID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		armorID: syntheticArmorEquipmentObjectBytes(),
	})
}

func openSyntheticEquipmentOwnedByStructureSave(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	ownedItemID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	otherItemID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	otherInventoryID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		structureID: syntheticStructureWithInventoryObjectBytes(inventoryID),
		ownedItemID: syntheticEquipmentObjectBytesWithOwnerInventory(inventoryID),
		otherItemID: syntheticEquipmentObjectBytesWithOwnerInventory(otherInventoryID),
	})
}

func openSyntheticEquipmentOwnedByStructureSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyStructureID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	ownedItemID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	otherItemID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	otherInventoryID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		structureID:       syntheticStructureWithInventoryObjectBytes(inventoryID),
		faultyStructureID: truncatedStructureObjectBytes(),
		ownedItemID:       syntheticEquipmentObjectBytesWithOwnerInventory(inventoryID),
		otherItemID:       syntheticEquipmentObjectBytesWithOwnerInventory(otherInventoryID),
	})
}

func syntheticEquipmentObjectBytes(isEngram bool) []byte {
	return syntheticEquipmentObjectBytesWithFlags(isEngram, false, false, 7.5, 3, 0.75)
}

func syntheticEquipmentObjectBytesWithFlags(isEngram bool, isEquipped bool, isBlueprint bool, rating float32, quality int32, durability float32) []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x1000000c, 1)
	writeFloatProperty(&props, 0x10000010, rating)
	writeIntProperty(&props, 0x10000011, quality)
	writeFloatProperty(&props, 0x10000012, durability)
	writePositionedUInt16Property(&props, 0x10000040, 2, 1000)
	writePositionedUInt16Property(&props, 0x10000040, 3, 1234)
	writeStringProperty(&props, 0x1000001b, "Survivor")
	writeStringProperty(&props, 0x1000001c, "Porters")
	if isEngram {
		writeBoolProperty(&props, 0x10000013, true)
	}
	if isEquipped {
		writeBoolProperty(&props, 0x10000022, true)
	}
	if isBlueprint {
		writeBoolProperty(&props, 0x1000000d, true)
	}
	return testfixtures.ObjectBytesWithProperties(0x1000000f, 0x10000004, props.Bytes())
}

func syntheticEquipmentObjectBytesWithOwnerInventory(ownerInventory uuid.UUID) []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x1000000c, 1)
	writeFloatProperty(&props, 0x10000010, 7.5)
	writeIntProperty(&props, 0x10000011, 3)
	writeFloatProperty(&props, 0x10000012, 0.75)
	writePositionedUInt16Property(&props, 0x10000040, 2, 1000)
	writePositionedUInt16Property(&props, 0x10000040, 3, 1234)
	testfixtures.WriteObjectReferencePropertyID(&props, 0x10000044, 0x1000001f, ownerInventory)
	return testfixtures.ObjectBytesWithProperties(0x1000000f, 0x10000004, props.Bytes())
}

func truncatedEquipmentObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000f))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticArmorEquipmentObjectBytes() []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x1000000c, 1)
	writePositionedUInt16Property(&props, 0x10000040, 1, 1000)
	writePositionedUInt16Property(&props, 0x10000040, 5, 500)
	writePositionedUInt16Property(&props, 0x10000040, 7, 200)
	return testfixtures.ObjectBytesWithProperties(0x10000042, 0x10000004, props.Bytes())
}

func boolPtr(value bool) *bool {
	return &value
}

func writeStringProperty(buf *bytes.Buffer, name uint32, value string) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001a))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+5))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	writeArkString(buf, value)
}

func writePositionedUInt16Property(buf *bytes.Buffer, name uint32, position int32, value uint16) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000041))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(2))
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, value)
}
