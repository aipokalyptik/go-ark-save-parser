package arkapi

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

func TestDinoAPIRecognizesApplicableBlueprints(t *testing.T) {
	api := DinoAPI{}
	for _, blueprint := range []string{
		"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		"Blueprint'/Game/ASA/Creatures/Foo/Foo_Character_BP.Foo_Character_BP_C'",
		"Blueprint'/Game/Mods/SDinoVariants/Bar/Bar_Character_BP.Bar_Character_BP_C'",
		"Blueprint'/Game/Extinction/CoreBlueprints/Weapons/PrimalItem_WeaponEmptyCryopod.PrimalItem_WeaponEmptyCryopod_C'",
		"Blueprint'/Game/Mods/DinoDepot/Items/DinoBalls/ItemDinoball.ItemDinoball_C'",
	} {
		if !api.IsApplicableBlueprint(blueprint) {
			t.Fatalf("IsApplicableBlueprint(%q) = false, want true", blueprint)
		}
	}
	if api.IsApplicableBlueprint("Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'") {
		t.Fatalf("IsApplicableBlueprint(weapon) = true, want false")
	}
}

func TestDinoAPIAllIgnoresEmptyCryopodItems(t *testing.T) {
	save := openSyntheticDinoSaveWithEmptyCryopod(t)
	defer save.Close()

	api := NewDino(save)
	dinos, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("All() length = %d, want 1 active dino and no empty cryopod placeholder", len(dinos))
	}
	for id, dino := range dinos {
		if id != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") || dino.IsCryopodded {
			t.Fatalf("All() dino = %s %#v, want only active non-cryopodded dino", id, dino)
		}
	}
}

func TestDinoAPIAllIncludesModernCryopoddedDinos(t *testing.T) {
	save := openSyntheticCryopoddedDinoSave(t)
	defer save.Close()

	api := NewDino(save)
	dinos, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("All() length = %d, want one cryopodded dino", len(dinos))
	}
	podID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	dino, ok := dinos[podID]
	if !ok {
		t.Fatalf("All() missing cryopodded dino keyed by cryopod UUID %s: %#v", podID, dinos)
	}
	if dino.ID1 != 1001 || dino.ID2 != 2002 || !dino.IsTamed || !dino.IsCryopodded {
		t.Fatalf("cryopodded dino = %#v", dino)
	}
	if dino.Location == nil || !dino.Location.InCryopod {
		t.Fatalf("cryopodded dino location = %#v, want in cryopod", dino.Location)
	}
	if dino.Stats == nil || dino.Stats.BaseLevel != 12 {
		t.Fatalf("cryopodded dino stats = %#v, want base level 12", dino.Stats)
	}
}

func TestDinoAPISaddlesFromCryopodsParsesModernEmbeddedSaddles(t *testing.T) {
	save := openSyntheticCryopoddedDinoSaveWithSaddle(t)
	defer save.Close()

	api := NewDino(save)
	saddles, err := api.SaddlesFromCryopods()
	if err != nil {
		t.Fatalf("SaddlesFromCryopods() error = %v", err)
	}
	podID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	if len(saddles) != 1 {
		t.Fatalf("SaddlesFromCryopods() length = %d, want 1: %#v", len(saddles), saddles)
	}
	saddle, ok := saddles[podID]
	if !ok {
		t.Fatalf("SaddlesFromCryopods() missing saddle keyed by cryopod UUID %s: %#v", podID, saddles)
	}
	if saddle.UUID != podID {
		t.Fatalf("saddle UUID = %s, want containing cryopod UUID %s", saddle.UUID, podID)
	}
	if saddle.Kind != arkobject.EquipmentSaddle {
		t.Fatalf("saddle kind = %q, want saddle", saddle.Kind)
	}
	wantBlueprint := "/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C"
	if saddle.Blueprint != wantBlueprint {
		t.Fatalf("saddle blueprint = %q, want %q", saddle.Blueprint, wantBlueprint)
	}
}

func TestDinoAPIAllReturnsMalformedCryopodError(t *testing.T) {
	save := openSyntheticMalformedCryopodSave(t)
	defer save.Close()

	api := NewDino(save)
	_, err := api.All()
	if !errors.Is(err, arkbinary.ErrUnsupportedEmbeddedDataVersion) {
		t.Fatalf("All() error = %v, want ErrUnsupportedEmbeddedDataVersion", err)
	}
}

func TestDinoAPIAllWithFaultsReportsMalformedCryopodError(t *testing.T) {
	save := openSyntheticMalformedCryopodSave(t)
	defer save.Close()

	api := NewDino(save)
	dinos, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	if len(dinos) != 0 {
		t.Fatalf("AllWithFaults() dinos length = %d, want 0", len(dinos))
	}
	if len(faults) != 1 || !errors.Is(faults[0].Err, arkbinary.ErrUnsupportedEmbeddedDataVersion) {
		t.Fatalf("AllWithFaults() faults = %#v, want unsupported embedded data fault", faults)
	}
}

func TestDinoAPIAllAndByClassReadLocalSaveDinos(t *testing.T) {
	save := openSyntheticDinoSave(t)
	defer save.Close()

	api := NewDino(save)
	dinos, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("All() length = %d, want 1", len(dinos))
	}
	for _, dino := range dinos {
		if dino.ID1 != 1001 || !dino.IsTamed || !dino.IsFemale || dino.Location == nil || dino.Location.X != 11 {
			t.Fatalf("Dino = %#v", dino)
		}
	}
	filtered, err := api.ByClass([]string{"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"})
	if err != nil {
		t.Fatalf("ByClass() error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("ByClass() length = %d, want 1", len(filtered))
	}
}

func TestDinoAPIClassHelpersFilterWildAndTamedDinos(t *testing.T) {
	save := openSyntheticDinoFilterSave(t)
	defer save.Close()

	api := NewDino(save)
	blueprints := []string{"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"}
	tamed, err := api.TamedByClass(blueprints, true)
	if err != nil {
		t.Fatalf("TamedByClass() error = %v", err)
	}
	if len(tamed) != 1 {
		t.Fatalf("TamedByClass() length = %d, want 1", len(tamed))
	}
	for _, dino := range tamed {
		if !dino.IsTamed {
			t.Fatalf("TamedByClass() dino = %#v", dino)
		}
	}

	wild, err := api.WildByClass(blueprints)
	if err != nil {
		t.Fatalf("WildByClass() error = %v", err)
	}
	if len(wild) != 1 {
		t.Fatalf("WildByClass() length = %d, want 1", len(wild))
	}
	for _, dino := range wild {
		if dino.IsTamed {
			t.Fatalf("WildByClass() dino = %#v", dino)
		}
	}
}

func TestDinoAPIByDinoIDFindsMatchingDinoAndHonorsWildInclusion(t *testing.T) {
	save := openSyntheticDinoFilterSave(t)
	defer save.Close()

	api := NewDino(save)
	id, dino, ok, err := api.ByDinoID(arkobject.DinoID{ID1: 1001, ID2: 2002}, false)
	if err != nil {
		t.Fatalf("ByDinoID(tamed) error = %v", err)
	}
	if !ok {
		t.Fatalf("ByDinoID(tamed) ok = false, want true")
	}
	if id != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") || !dino.IsTamed || dino.ID1 != 1001 || dino.ID2 != 2002 {
		t.Fatalf("ByDinoID(tamed) = %s %#v", id, dino)
	}

	_, _, ok, err = api.ByDinoID(arkobject.DinoID{ID1: 3003, ID2: 4004}, false)
	if err != nil {
		t.Fatalf("ByDinoID(wild excluded) error = %v", err)
	}
	if ok {
		t.Fatalf("ByDinoID(wild excluded) ok = true, want false")
	}

	id, dino, ok, err = api.ByDinoID(arkobject.DinoID{ID1: 3003, ID2: 4004}, true)
	if err != nil {
		t.Fatalf("ByDinoID(wild included) error = %v", err)
	}
	if !ok {
		t.Fatalf("ByDinoID(wild included) ok = false, want true")
	}
	if id != uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff") || dino.IsTamed || dino.ID1 != 3003 || dino.ID2 != 4004 {
		t.Fatalf("ByDinoID(wild included) = %s %#v", id, dino)
	}
}

func TestDinoAPIAllWithFaultsKeepsValidDinosAndReportsParseFaults(t *testing.T) {
	save := openSyntheticDinoSaveWithFault(t)
	defer save.Close()

	api := NewDino(save)
	dinos, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("AllWithFaults() dinos length = %d, want 1", len(dinos))
	}
	for _, dino := range dinos {
		if dino.ID1 != 1001 || !dino.IsTamed || !dino.IsFemale || dino.Location == nil || dino.Location.X != 11 {
			t.Fatalf("Dino = %#v", dino)
		}
	}
	if len(faults) != 1 || faults[0].ClassName != "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'" || faults[0].Err == nil {
		t.Fatalf("AllWithFaults() faults = %#v, want one dino parse fault", faults)
	}
}

func TestDinoAPIFiltersBySexDeathAndBabyState(t *testing.T) {
	save := openSyntheticDinoFilterSave(t)
	defer save.Close()

	api := NewDino(save)
	females, err := api.Females()
	if err != nil {
		t.Fatalf("Females() error = %v", err)
	}
	if len(females) != 1 {
		t.Fatalf("Females() length = %d, want 1", len(females))
	}
	males, err := api.Males()
	if err != nil {
		t.Fatalf("Males() error = %v", err)
	}
	if len(males) != 1 {
		t.Fatalf("Males() length = %d, want 1", len(males))
	}
	dead, err := api.Dead()
	if err != nil {
		t.Fatalf("Dead() error = %v", err)
	}
	if len(dead) != 1 {
		t.Fatalf("Dead() length = %d, want 1", len(dead))
	}
	alive, err := api.Alive()
	if err != nil {
		t.Fatalf("Alive() error = %v", err)
	}
	if len(alive) != 1 {
		t.Fatalf("Alive() length = %d, want 1", len(alive))
	}
	babies, err := api.Babies()
	if err != nil {
		t.Fatalf("Babies() error = %v", err)
	}
	if len(babies) != 1 {
		t.Fatalf("Babies() length = %d, want 1", len(babies))
	}
}

func TestDinoAPIBabiesFilteredMatchesUpstreamInclusionFlags(t *testing.T) {
	save := openSyntheticDinoBabyFilterSave(t)
	defer save.Close()

	api := NewDino(save)
	tamedOnly, err := api.BabiesFiltered(BabyFilterOptions{
		IncludeTamed:      true,
		IncludeCryopodded: true,
		IncludeWild:       false,
	})
	if err != nil {
		t.Fatalf("BabiesFiltered(tamed) error = %v", err)
	}
	if len(tamedOnly) != 1 {
		t.Fatalf("BabiesFiltered(tamed) length = %d, want 1", len(tamedOnly))
	}
	for _, dino := range tamedOnly {
		if !dino.IsTamed || !dino.IsBaby {
			t.Fatalf("BabiesFiltered(tamed) dino = %#v", dino)
		}
	}

	all, err := api.BabiesFiltered(BabyFilterOptions{
		IncludeTamed:      true,
		IncludeCryopodded: true,
		IncludeWild:       true,
	})
	if err != nil {
		t.Fatalf("BabiesFiltered(all) error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("BabiesFiltered(all) length = %d, want 2", len(all))
	}

	wildOnly, err := api.BabiesFiltered(BabyFilterOptions{
		IncludeTamed:      false,
		IncludeCryopodded: true,
		IncludeWild:       true,
	})
	if err != nil {
		t.Fatalf("BabiesFiltered(wild) error = %v", err)
	}
	if len(wildOnly) != 1 {
		t.Fatalf("BabiesFiltered(wild) length = %d, want 1", len(wildOnly))
	}
	for _, dino := range wildOnly {
		if dino.IsTamed || !dino.IsBaby {
			t.Fatalf("BabiesFiltered(wild) dino = %#v", dino)
		}
	}
}

func TestDinoAPIReadsTamedDetailsAndOwner(t *testing.T) {
	save := openSyntheticDinoDetailSave(t)
	defer save.Close()

	api := NewDino(save)
	dinos, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("All() length = %d, want 1", len(dinos))
	}
	for _, dino := range dinos {
		if dino.TamedName != "Blue" || !dino.IsNeutered {
			t.Fatalf("dino tamed details = %#v", dino)
		}
		if dino.ColorSetIndices != [6]int{11, 0, 0, 44, 0, 0} {
			t.Fatalf("ColorSetIndices = %#v", dino.ColorSetIndices)
		}
		if dino.ColorSetNames != [6]string{"None", "Blue", "None", "None", "Black", "None"} {
			t.Fatalf("ColorSetNames = %#v", dino.ColorSetNames)
		}
		if dino.UploadedFromServerName != "TheIsland" {
			t.Fatalf("UploadedFromServerName = %q", dino.UploadedFromServerName)
		}
		if dino.InventoryUUID == nil || dino.InventoryUUID.String() != "99999999-aaaa-bbbb-cccc-ddddeeeeffff" {
			t.Fatalf("InventoryUUID = %v", dino.InventoryUUID)
		}
		if dino.Owner.TribeName != "Porters" || dino.Owner.TamerTribeID != 555 || dino.Owner.TargetTeam != 555 {
			t.Fatalf("dino owner tribe fields = %#v", dino.Owner)
		}
		if dino.Owner.PlayerName != "Survivor" || dino.Owner.PlayerID != 42 || dino.Owner.ImprinterUniqueID != "eos-survivor" {
			t.Fatalf("dino owner player fields = %#v", dino.Owner)
		}
		if len(dino.ParsedGeneTraits) != 2 || dino.ParsedGeneTraits[0].Name != "MutableMelee" || dino.ParsedGeneTraits[0].Level != 2 {
			t.Fatalf("dino parsed gene traits = %#v", dino.ParsedGeneTraits)
		}
	}
}

func TestDinoAPIOwnedByTribeFiltersTamedDinosByTargetTeam(t *testing.T) {
	save := openSyntheticDinoDetailSave(t)
	defer save.Close()

	api := NewDino(save)
	owned, err := api.OwnedByTribe(555, true)
	if err != nil {
		t.Fatalf("OwnedByTribe() error = %v", err)
	}
	if len(owned) != 1 {
		t.Fatalf("OwnedByTribe(555) length = %d, want 1", len(owned))
	}
	for _, dino := range owned {
		if !dino.IsTamed || dino.Owner.TargetTeam != 555 {
			t.Fatalf("OwnedByTribe(555) dino = %#v", dino)
		}
	}

	missing, err := api.OwnedByTribe(111, true)
	if err != nil {
		t.Fatalf("OwnedByTribe(missing) error = %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("OwnedByTribe(111) length = %d, want 0", len(missing))
	}
}

func TestDinoAPIWildTamedFiltersTamedDinosWithoutAncestors(t *testing.T) {
	api := DinoAPI{}
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	thirdID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	dinos := map[uuid.UUID]arkobject.Dino{
		firstID: {
			IsTamed: true,
		},
		secondID: {
			IsTamed:     true,
			AncestorIDs: []arkobject.DinoID{{ID1: 1, ID2: 2}},
		},
		thirdID: {
			IsTamed: false,
		},
	}

	wildTamed := api.FilterWildTamed(dinos)
	if len(wildTamed) != 1 {
		t.Fatalf("FilterWildTamed() length = %d, want 1", len(wildTamed))
	}
	if _, ok := wildTamed[firstID]; !ok {
		t.Fatalf("FilterWildTamed() missing ancestorless tamed dino: %#v", wildTamed)
	}
	if !dinos[firstID].IsWildTamed() {
		t.Fatalf("IsWildTamed() = false, want true")
	}
	if dinos[secondID].IsWildTamed() || dinos[thirdID].IsWildTamed() {
		t.Fatalf("IsWildTamed() classified bred or wild dino as wild-tamed")
	}
}

func TestDinoAPIFilterWildTamableDropsKnownNonTameableClasses(t *testing.T) {
	api := DinoAPI{}
	raptorID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	dragonflyID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	tamedID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	dinos := map[uuid.UUID]arkobject.Dino{
		raptorID: {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		},
		dragonflyID: {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Dragonfly/Dragonfly_Character_BP.Dragonfly_Character_BP_C'",
		},
		tamedID: {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			IsTamed:   true,
		},
	}

	tamables := api.FilterWildTamable(dinos)
	if len(tamables) != 1 {
		t.Fatalf("FilterWildTamable() length = %d, want 1", len(tamables))
	}
	if _, ok := tamables[raptorID]; !ok {
		t.Fatalf("FilterWildTamable() missing wild tameable raptor: %#v", tamables)
	}
	if api.IsNonTameableDino("Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'") {
		t.Fatalf("IsNonTameableDino(raptor) = true, want false")
	}
	if !api.IsNonTameableDino("Blueprint'/Game/PrimalEarth/Dinos/Dragonfly/Dragonfly_Character_BP.Dragonfly_Character_BP_C'") {
		t.Fatalf("IsNonTameableDino(dragonfly) = false, want true")
	}
}

func TestDinoAPIContainerOfInventoryFindsInventoryBearingDino(t *testing.T) {
	save := openSyntheticDinoDetailSave(t)
	defer save.Close()

	api := NewDino(save)
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	id, dino, ok, err := api.ContainerOfInventory(inventoryID, true)
	if err != nil {
		t.Fatalf("ContainerOfInventory() error = %v", err)
	}
	if !ok {
		t.Fatalf("ContainerOfInventory() ok = false, want true")
	}
	if id != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") {
		t.Fatalf("ContainerOfInventory() id = %s", id)
	}
	if dino.InventoryUUID == nil || *dino.InventoryUUID != inventoryID || !dino.IsTamed {
		t.Fatalf("ContainerOfInventory() dino = %#v", dino)
	}

	_, _, ok, err = api.ContainerOfInventory(uuid.MustParse("11111111-2222-3333-4444-555555555555"), true)
	if err != nil {
		t.Fatalf("ContainerOfInventory(missing) error = %v", err)
	}
	if ok {
		t.Fatalf("ContainerOfInventory(missing) ok = true, want false")
	}
}

func TestDinoAPIReadsBabyMaturationStage(t *testing.T) {
	save := openSyntheticDinoBabyStageSave(t)
	defer save.Close()

	api := NewDino(save)
	babies, err := api.Babies()
	if err != nil {
		t.Fatalf("Babies() error = %v", err)
	}
	if len(babies) != 1 {
		t.Fatalf("Babies() length = %d, want 1", len(babies))
	}
	for _, dino := range babies {
		if dino.MaturationPercent != 75 || dino.BabyStage != arkobject.BabyStageAdolescent {
			t.Fatalf("baby maturation = %#v", dino)
		}
	}
}

func TestDinoAPIReadsLinkedStatusComponentStats(t *testing.T) {
	save := openSyntheticDinoStatsSave(t)
	defer save.Close()

	api := NewDino(save)
	dinos, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("All() length = %d, want 1", len(dinos))
	}
	for _, dino := range dinos {
		if dino.Stats == nil {
			t.Fatalf("Stats = nil, want parsed status component")
		}
		if dino.Stats.BaseLevel != 12 || dino.Stats.CurrentLevel != 12 {
			t.Fatalf("Dino stats levels = %#v", dino.Stats)
		}
		if dino.Stats.BaseStatPoints.Health != 5 || dino.Stats.AddedStatPoints.MeleeDamage != 2 {
			t.Fatalf("Dino stat points = %#v", dino.Stats)
		}
		if dino.Stats.ImprintingPercent != 87.5 {
			t.Fatalf("Dino imprinting = %f", dino.Stats.ImprintingPercent)
		}
	}
}

func TestDinoAPIFiltersByLevelAndStats(t *testing.T) {
	save := openSyntheticDinoStatsSave(t)
	defer save.Close()

	api := NewDino(save)
	highLevel, err := api.LevelAtLeast(12)
	if err != nil {
		t.Fatalf("LevelAtLeast() error = %v", err)
	}
	if len(highLevel) != 1 {
		t.Fatalf("LevelAtLeast(12) length = %d, want 1", len(highLevel))
	}
	tooHigh, err := api.LevelAtLeast(13)
	if err != nil {
		t.Fatalf("LevelAtLeast(13) error = %v", err)
	}
	if len(tooHigh) != 0 {
		t.Fatalf("LevelAtLeast(13) length = %d, want 0", len(tooHigh))
	}
	health, err := api.WithStatAtLeast(6, arkobject.DinoStatHealth)
	if err != nil {
		t.Fatalf("WithStatAtLeast(health) error = %v", err)
	}
	if len(health) != 1 {
		t.Fatalf("WithStatAtLeast(health) length = %d, want 1", len(health))
	}
	weight, err := api.WithStatAtLeast(8, arkobject.DinoStatWeight)
	if err != nil {
		t.Fatalf("WithStatAtLeast(weight) error = %v", err)
	}
	if len(weight) != 0 {
		t.Fatalf("WithStatAtLeast(weight) length = %d, want 0", len(weight))
	}
	baseHealth, err := api.WithBaseStatAtLeast(5, arkobject.DinoStatHealth)
	if err != nil {
		t.Fatalf("WithBaseStatAtLeast() error = %v", err)
	}
	if len(baseHealth) != 1 {
		t.Fatalf("WithBaseStatAtLeast() length = %d, want 1", len(baseHealth))
	}
	mutatedHealth, err := api.WithMutatedStatAtLeast(1, arkobject.DinoStatHealth)
	if err != nil {
		t.Fatalf("WithMutatedStatAtLeast() error = %v", err)
	}
	if len(mutatedHealth) != 1 {
		t.Fatalf("WithMutatedStatAtLeast() length = %d, want 1", len(mutatedHealth))
	}
	minLevel := int32(12)
	maxLevel := int32(12)
	tamed := false
	filtered, err := api.Filtered(DinoFilterOptions{
		MinLevel:    &minLevel,
		MaxLevel:    &maxLevel,
		Tamed:       &tamed,
		StatMinimum: 6,
		Stats:       []arkobject.DinoStat{arkobject.DinoStatHealth},
		Blueprints:  []string{"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"},
	})
	if err != nil {
		t.Fatalf("Filtered() error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("Filtered() length = %d, want 1", len(filtered))
	}
	maxLevel = 11
	filtered, err = api.Filtered(DinoFilterOptions{MaxLevel: &maxLevel})
	if err != nil {
		t.Fatalf("Filtered(max) error = %v", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("Filtered(max) length = %d, want 0", len(filtered))
	}
}

func TestDinoAPILevelHelpersFilterByTamedState(t *testing.T) {
	wildSave := openSyntheticDinoStatsSave(t)
	defer wildSave.Close()

	wildAPI := NewDino(wildSave)
	wild, err := wildAPI.WildLevelAtLeast(12)
	if err != nil {
		t.Fatalf("WildLevelAtLeast() error = %v", err)
	}
	if len(wild) != 1 {
		t.Fatalf("WildLevelAtLeast() length = %d, want 1", len(wild))
	}
	tamedFromWildSave, err := wildAPI.TamedLevelAtLeast(12)
	if err != nil {
		t.Fatalf("TamedLevelAtLeast(wild save) error = %v", err)
	}
	if len(tamedFromWildSave) != 0 {
		t.Fatalf("TamedLevelAtLeast(wild save) length = %d, want 0", len(tamedFromWildSave))
	}

	tamedSave := openSyntheticTamedDinoStatsSave(t)
	defer tamedSave.Close()

	tamedAPI := NewDino(tamedSave)
	tamed, err := tamedAPI.TamedLevelAtLeast(12)
	if err != nil {
		t.Fatalf("TamedLevelAtLeast() error = %v", err)
	}
	if len(tamed) != 1 {
		t.Fatalf("TamedLevelAtLeast() length = %d, want 1", len(tamed))
	}
	wildFromTamedSave, err := tamedAPI.WildLevelAtLeast(12)
	if err != nil {
		t.Fatalf("WildLevelAtLeast(tamed save) error = %v", err)
	}
	if len(wildFromTamedSave) != 0 {
		t.Fatalf("WildLevelAtLeast(tamed save) length = %d, want 0", len(wildFromTamedSave))
	}
}

func TestDinoAPIFiltersByGeneTrait(t *testing.T) {
	save := openSyntheticDinoDetailSave(t)
	defer save.Close()

	api := NewDino(save)
	byName, err := api.WithGeneTrait("MutableMelee")
	if err != nil {
		t.Fatalf("WithGeneTrait(name) error = %v", err)
	}
	if len(byName) != 1 {
		t.Fatalf("WithGeneTrait(name) length = %d, want 1", len(byName))
	}
	byLevel, err := api.WithGeneTrait("MutableMelee", 2)
	if err != nil {
		t.Fatalf("WithGeneTrait(name, level) error = %v", err)
	}
	if len(byLevel) != 1 {
		t.Fatalf("WithGeneTrait(name, level) length = %d, want 1", len(byLevel))
	}
	wrongLevel, err := api.WithGeneTrait("MutableMelee", 3)
	if err != nil {
		t.Fatalf("WithGeneTrait(wrong level) error = %v", err)
	}
	if len(wrongLevel) != 0 {
		t.Fatalf("WithGeneTrait(wrong level) length = %d, want 0", len(wrongLevel))
	}
	fallback, err := api.WithGeneTrait("Robust")
	if err != nil {
		t.Fatalf("WithGeneTrait(fallback) error = %v", err)
	}
	if len(fallback) != 1 {
		t.Fatalf("WithGeneTrait(fallback) length = %d, want 1", len(fallback))
	}
	filtered, err := api.Filtered(DinoFilterOptions{GeneTraits: []string{"MutableMelee"}})
	if err != nil {
		t.Fatalf("Filtered(gene trait) error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("Filtered(gene trait) length = %d, want 1", len(filtered))
	}
	missing, err := api.Filtered(DinoFilterOptions{GeneTraits: []string{"MissingTrait"}})
	if err != nil {
		t.Fatalf("Filtered(missing gene trait) error = %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("Filtered(missing gene trait) length = %d, want 0", len(missing))
	}
}

func TestDinoAPIFiltersCryopoddedDinos(t *testing.T) {
	save := openSyntheticCryopoddedDinoSave(t)
	defer save.Close()

	api := NewDino(save)
	cryopodded, err := api.InCryopods()
	if err != nil {
		t.Fatalf("InCryopods() error = %v", err)
	}
	if len(cryopodded) != 1 {
		t.Fatalf("InCryopods() length = %d, want 1", len(cryopodded))
	}
	notCryopodded, err := api.NotInCryopods()
	if err != nil {
		t.Fatalf("NotInCryopods() error = %v", err)
	}
	if len(notCryopodded) != 0 {
		t.Fatalf("NotInCryopods() length = %d, want 0", len(notCryopodded))
	}
	wantCryopodded := true
	filtered, err := api.Filtered(DinoFilterOptions{Cryopodded: &wantCryopodded})
	if err != nil {
		t.Fatalf("Filtered(cryopodded) error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("Filtered(cryopodded) length = %d, want 1", len(filtered))
	}
}

func TestDinoAPIBestDinoForStatUsesParsedStatusStats(t *testing.T) {
	save := openSyntheticDinoStatsSave(t)
	defer save.Close()

	api := NewDino(save)
	id, dino, stat, points, ok, err := api.BestDinoForStat()
	if err != nil {
		t.Fatalf("BestDinoForStat() error = %v", err)
	}
	if !ok {
		t.Fatalf("BestDinoForStat() ok = false, want true")
	}
	if id != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") {
		t.Fatalf("BestDinoForStat() id = %s", id)
	}
	if dino.ID1 != 1001 || stat != arkobject.DinoStatHealth || points != 6 {
		t.Fatalf("BestDinoForStat() = %#v, %v, %d; want dino 1001 health 6", dino, stat, points)
	}

	_, _, baseStat, basePoints, ok, err := api.BestDinoForStat(arkobject.StatScopeBase)
	if err != nil {
		t.Fatalf("BestDinoForStat(base) error = %v", err)
	}
	if !ok || baseStat != arkobject.DinoStatHealth || basePoints != 5 {
		t.Fatalf("BestDinoForStat(base) = %v, %d, %v; want health, 5, true", baseStat, basePoints, ok)
	}
}

func TestDinoAPIBestDinoForStatFilteredAppliesUpstreamStyleOptions(t *testing.T) {
	save := openSyntheticTamedDinoStatsSave(t)
	defer save.Close()

	api := NewDino(save)
	levelUpperBound := int32(12)
	id, dino, stat, points, ok, err := api.BestDinoForStatFiltered(DinoBestStatOptions{
		Blueprints:      []string{"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"},
		Stats:           []arkobject.DinoStat{arkobject.DinoStatHealth},
		OnlyTamed:       true,
		BaseStat:        true,
		LevelUpperBound: &levelUpperBound,
	})
	if err != nil {
		t.Fatalf("BestDinoForStatFiltered() error = %v", err)
	}
	if !ok {
		t.Fatalf("BestDinoForStatFiltered() ok = false, want true")
	}
	if id != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") || !dino.IsTamed {
		t.Fatalf("BestDinoForStatFiltered() dino = %s %#v", id, dino)
	}
	if stat != arkobject.DinoStatHealth || points != 5 {
		t.Fatalf("BestDinoForStatFiltered() stat = %v points = %d, want health 5", stat, points)
	}

	tooLow := int32(11)
	_, _, _, _, ok, err = api.BestDinoForStatFiltered(DinoBestStatOptions{
		OnlyTamed:       true,
		LevelUpperBound: &tooLow,
	})
	if err != nil {
		t.Fatalf("BestDinoForStatFiltered(level cap) error = %v", err)
	}
	if ok {
		t.Fatalf("BestDinoForStatFiltered(level cap) ok = true, want false")
	}

	_, _, _, _, ok, err = api.BestDinoForStatFiltered(DinoBestStatOptions{OnlyUntamed: true})
	if err != nil {
		t.Fatalf("BestDinoForStatFiltered(untamed) error = %v", err)
	}
	if ok {
		t.Fatalf("BestDinoForStatFiltered(untamed) ok = true, want false")
	}

	_, _, _, _, _, err = api.BestDinoForStatFiltered(DinoBestStatOptions{OnlyTamed: true, OnlyUntamed: true})
	if err == nil {
		t.Fatalf("BestDinoForStatFiltered(conflicting tame filters) error = nil, want error")
	}
	_, _, _, _, _, err = api.BestDinoForStatFiltered(DinoBestStatOptions{BaseStat: true, MutatedStat: true})
	if err == nil {
		t.Fatalf("BestDinoForStatFiltered(conflicting stat scopes) error = nil, want error")
	}
}

func TestDinoAPIMostMutatedTamedUsesTotalMutations(t *testing.T) {
	save := openSyntheticTamedDinoStatsSave(t)
	defer save.Close()

	api := NewDino(save)
	id, dino, total, ok, err := api.MostMutatedTamed()
	if err != nil {
		t.Fatalf("MostMutatedTamed() error = %v", err)
	}
	if !ok {
		t.Fatalf("MostMutatedTamed() ok = false, want true")
	}
	if id != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") || dino.ID1 != 1001 || total != 1 {
		t.Fatalf("MostMutatedTamed() = %s, %#v, %d; want synthetic dino total 1", id, dino, total)
	}
}

func TestDinoAPIStatSelectionReportsNoResultWhenStatsAreMissing(t *testing.T) {
	save := openSyntheticDinoSave(t)
	defer save.Close()

	api := NewDino(save)
	_, _, _, _, ok, err := api.BestDinoForStat()
	if err != nil {
		t.Fatalf("BestDinoForStat() error = %v", err)
	}
	if ok {
		t.Fatalf("BestDinoForStat() ok = true, want false")
	}
	_, _, _, ok, err = api.MostMutatedTamed()
	if err != nil {
		t.Fatalf("MostMutatedTamed() error = %v", err)
	}
	if ok {
		t.Fatalf("MostMutatedTamed() ok = true, want false")
	}
}

func TestDinoAPICountsByLevelClassAndTamedState(t *testing.T) {
	api := NewDino(nil)
	dinos := map[uuid.UUID]arkobject.Dino{
		uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			IsTamed:   true,
			Stats:     &arkobject.DinoStats{CurrentLevel: 12},
		},
		uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			IsTamed:   false,
			Stats:     &arkobject.DinoStats{CurrentLevel: 12},
		},
		uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000"): {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Dodo/Dodo_Character_BP.Dodo_Character_BP_C'",
			IsTamed:   true,
			Stats:     &arkobject.DinoStats{CurrentLevel: 8},
		},
	}

	byLevel := api.CountByLevel(dinos)
	if byLevel[12] != 2 || byLevel[8] != 1 {
		t.Fatalf("CountByLevel() = %#v", byLevel)
	}
	byClass := api.CountByClass(dinos)
	if byClass["Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"] != 2 ||
		byClass["Blueprint'/Game/PrimalEarth/Dinos/Dodo/Dodo_Character_BP.Dodo_Character_BP_C'"] != 1 {
		t.Fatalf("CountByClass() = %#v", byClass)
	}
	byShortName := api.CountByShortName(dinos)
	if byShortName["Raptor"] != 2 || byShortName["Dodo"] != 1 {
		t.Fatalf("CountByShortName() = %#v", byShortName)
	}
	byTamed := api.CountByTamed(dinos)
	if byTamed[true] != 2 || byTamed[false] != 1 {
		t.Fatalf("CountByTamed() = %#v", byTamed)
	}
	dinos[uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")] = arkobject.Dino{
		Blueprint:    "Blueprint'/Game/PrimalEarth/Dinos/Dodo/Dodo_Character_BP.Dodo_Character_BP_C'",
		IsCryopodded: true,
		Stats:        &arkobject.DinoStats{CurrentLevel: 8},
	}
	byCryopodded := api.CountByCryopodded(dinos)
	if byCryopodded[true] != 1 || byCryopodded[false] != 2 {
		t.Fatalf("CountByCryopodded() = %#v", byCryopodded)
	}
	byCryopoddedClass := api.CountCryopoddedByClass(dinos)
	if byCryopoddedClass["all"] != 1 ||
		byCryopoddedClass["Blueprint'/Game/PrimalEarth/Dinos/Dodo/Dodo_Character_BP.Dodo_Character_BP_C'"] != 1 {
		t.Fatalf("CountCryopoddedByClass() = %#v", byCryopoddedClass)
	}
	byCryopoddedShortName := api.CountCryopoddedByShortName(dinos)
	if byCryopoddedShortName["all"] != 1 || byCryopoddedShortName["Dodo"] != 1 {
		t.Fatalf("CountCryopoddedByShortName() = %#v", byCryopoddedShortName)
	}
	if _, ok := byCryopoddedClass["Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"]; ok {
		t.Fatalf("CountCryopoddedByClass() counted non-cryopodded raptor: %#v", byCryopoddedClass)
	}
}

func TestDinoAPIFilterChildlessTamedDinosUsesAncestorIDsAndSkipsBabiesAsParents(t *testing.T) {
	api := NewDino(nil)
	parentID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	childID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	babyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	unrelatedID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	wildID := uuid.MustParse("eeeeeeee-ffff-0000-1111-222222222222")
	dinos := map[uuid.UUID]arkobject.Dino{
		parentID: {
			ID1:     11,
			ID2:     12,
			IsTamed: true,
		},
		childID: {
			ID1:         21,
			ID2:         22,
			IsTamed:     true,
			AncestorIDs: []arkobject.DinoID{{ID1: 11, ID2: 12}},
		},
		babyID: {
			ID1:         31,
			ID2:         32,
			IsTamed:     true,
			IsBaby:      true,
			AncestorIDs: []arkobject.DinoID{{ID1: 21, ID2: 22}},
		},
		unrelatedID: {
			ID1:     41,
			ID2:     42,
			IsTamed: true,
		},
		wildID: {
			ID1: 51,
			ID2: 52,
		},
	}

	childless := api.FilterChildlessTamed(dinos)

	if len(childless) != 3 {
		t.Fatalf("FilterChildlessTamed() length = %d, want 3: %#v", len(childless), childless)
	}
	if _, ok := childless[parentID]; ok {
		t.Fatalf("FilterChildlessTamed() included parent ancestor")
	}
	if _, ok := childless[childID]; !ok {
		t.Fatalf("FilterChildlessTamed() excluded adult child referenced only by baby")
	}
	if _, ok := childless[babyID]; !ok {
		t.Fatalf("FilterChildlessTamed() excluded baby")
	}
	if _, ok := childless[unrelatedID]; !ok {
		t.Fatalf("FilterChildlessTamed() excluded unrelated tamed dino")
	}
	if _, ok := childless[wildID]; ok {
		t.Fatalf("FilterChildlessTamed() included wild dino")
	}
}

func TestDinoAPIPedigreeHelpersIndexChildrenAndWalkDescendants(t *testing.T) {
	api := NewDino(nil)
	parentAID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	parentBID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	childID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	grandchildID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	unrelatedID := uuid.MustParse("eeeeeeee-ffff-0000-1111-222222222222")
	wildChildID := uuid.MustParse("ffffffff-0000-1111-2222-333333333333")
	parentAArkID := arkobject.DinoID{ID1: 11, ID2: 12}
	parentBArkID := arkobject.DinoID{ID1: 21, ID2: 22}
	childArkID := arkobject.DinoID{ID1: 31, ID2: 32}
	grandchildArkID := arkobject.DinoID{ID1: 41, ID2: 42}
	dinos := map[uuid.UUID]arkobject.Dino{
		parentBID: {
			ID1:     parentBArkID.ID1,
			ID2:     parentBArkID.ID2,
			IsTamed: true,
		},
		grandchildID: {
			ID1:         grandchildArkID.ID1,
			ID2:         grandchildArkID.ID2,
			IsTamed:     true,
			AncestorIDs: []arkobject.DinoID{childArkID},
		},
		parentAID: {
			ID1:         parentAArkID.ID1,
			ID2:         parentAArkID.ID2,
			IsTamed:     true,
			AncestorIDs: []arkobject.DinoID{childArkID},
		},
		unrelatedID: {
			ID1:     51,
			ID2:     52,
			IsTamed: true,
		},
		childID: {
			ID1:         childArkID.ID1,
			ID2:         childArkID.ID2,
			IsTamed:     true,
			AncestorIDs: []arkobject.DinoID{parentAArkID, parentBArkID},
		},
		wildChildID: {
			ID1:         61,
			ID2:         62,
			AncestorIDs: []arkobject.DinoID{parentAArkID},
		},
	}

	children := api.ChildrenByAncestor(dinos)

	if got := children[parentAArkID]; len(got) != 1 || got[0] != childID {
		t.Fatalf("ChildrenByAncestor(parent A) = %#v, want [%s]", got, childID)
	}
	if got := children[parentBArkID]; len(got) != 1 || got[0] != childID {
		t.Fatalf("ChildrenByAncestor(parent B) = %#v, want [%s]", got, childID)
	}
	if got := children[childArkID]; len(got) != 2 || got[0] != parentAID || got[1] != grandchildID {
		t.Fatalf("ChildrenByAncestor(child) = %#v, want [%s %s]", got, parentAID, grandchildID)
	}

	descendants := api.DescendantsOf(dinos, parentAArkID)
	if len(descendants) != 2 {
		t.Fatalf("DescendantsOf(parent A) length = %d, want 2: %#v", len(descendants), descendants)
	}
	if _, ok := descendants[childID]; !ok {
		t.Fatalf("DescendantsOf(parent A) missing child")
	}
	if _, ok := descendants[grandchildID]; !ok {
		t.Fatalf("DescendantsOf(parent A) missing grandchild")
	}
	if _, ok := descendants[parentAID]; ok {
		t.Fatalf("DescendantsOf(parent A) included root")
	}
	if _, ok := descendants[unrelatedID]; ok {
		t.Fatalf("DescendantsOf(parent A) included unrelated dino")
	}
	if _, ok := descendants[wildChildID]; ok {
		t.Fatalf("DescendantsOf(parent A) included wild child")
	}
}

func openSyntheticDinoStatsSave(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	statusID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		dinoID:   syntheticDinoStatsObjectBytesWithTamed(statusID, false),
		statusID: syntheticDinoStatusObjectBytes(),
	})
}

func openSyntheticTamedDinoStatsSave(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	statusID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		dinoID:   syntheticDinoStatsObjectBytesWithTamed(statusID, true),
		statusID: syntheticDinoStatusObjectBytes(),
	})
}

func syntheticDinoStatsObjectBytesWithTamed(statusID uuid.UUID, tamed bool) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, 1001)
	writeIntProperty(&buf, 0x10000016, 2002)
	if tamed {
		writeDoubleProperty(&buf, 0x10000018, 42)
	}
	writeObjectReferenceProperty(&buf, 0x10000035, statusID)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticDinoStatusObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000036))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000037, 12)
	writePositionedIntProperty(&buf, 0x10000038, 0, 5)
	writePositionedIntProperty(&buf, 0x10000038, 7, 3)
	writePositionedIntProperty(&buf, 0x10000039, 8, 2)
	writePositionedIntProperty(&buf, 0x1000003a, 0, 1)
	writePositionedFloatProperty(&buf, 0x1000003b, 0, 1234.5)
	writePositionedFloatProperty(&buf, 0x1000003b, 7, 321.25)
	writeFloatProperty(&buf, 0x1000003c, 0.875)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func openSyntheticDinoSave(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	otherID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	return openSyntheticSaveWith(t, "dinos.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransforms(dinoID),
	}, map[uuid.UUID][]byte{
		dinoID:  syntheticDinoObjectBytes(),
		otherID: syntheticObjectBytes(0x10000001),
	})
}

func openSyntheticDinoSaveWithEmptyCryopod(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	podID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	return openSyntheticSaveWith(t, "dinos.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransforms(dinoID),
	}, map[uuid.UUID][]byte{
		dinoID: syntheticDinoObjectBytes(),
		podID:  syntheticObjectBytes(0x10000047),
	})
}

func openSyntheticCryopoddedDinoSave(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	podID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	payload := syntheticCryopodDinoPayload(t, dinoID, statusID)
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		podID: syntheticCryopodItemObjectBytes(payload),
	})
}

func openSyntheticCryopoddedDinoSaveWithSaddle(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	podID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	dinoPayload := syntheticCryopodDinoPayload(t, dinoID, statusID)
	saddlePayload := syntheticCryopodSaddlePayload()
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		podID: syntheticCryopodItemObjectBytesWithPayloads(dinoPayload, saddlePayload),
	})
}

func openSyntheticMalformedCryopodSave(t *testing.T) *arksave.Save {
	t.Helper()

	podID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		podID: syntheticCryopodItemObjectBytes(syntheticLegacyCryopodPayload()),
	})
}

func openSyntheticDinoSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "dinos.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransforms(dinoID),
	}, map[uuid.UUID][]byte{
		dinoID:   syntheticDinoObjectBytes(),
		faultyID: truncatedDinoObjectBytes(),
	})
}

func openSyntheticDinoFilterSave(t *testing.T) *arksave.Save {
	t.Helper()

	femaleAliveID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	maleDeadID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		femaleAliveID: syntheticDinoObjectBytesWithFlags(1001, 2002, true, false, false, true),
		maleDeadID:    syntheticDinoObjectBytesWithFlags(3003, 4004, false, true, true, false),
	})
}

func openSyntheticDinoDetailSave(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		dinoID: syntheticDinoDetailObjectBytes(),
	})
}

func openSyntheticDinoBabyStageSave(t *testing.T) *arksave.Save {
	t.Helper()

	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		dinoID: syntheticDinoBabyObjectBytes(0.75),
	})
}

func openSyntheticDinoBabyFilterSave(t *testing.T) *arksave.Save {
	t.Helper()

	tamedBabyID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	wildBabyID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	adultID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "dinos.ark", nil, map[uuid.UUID][]byte{
		tamedBabyID: syntheticDinoObjectBytesWithFlags(1001, 2002, true, false, true, true),
		wildBabyID:  syntheticDinoObjectBytesWithFlags(3003, 4004, false, false, true, false),
		adultID:     syntheticDinoObjectBytesWithFlags(5005, 6006, true, false, false, true),
	})
}

func syntheticDinoObjectBytes() []byte {
	return syntheticDinoObjectBytesWithFlags(1001, 2002, true, false, false, true)
}

func syntheticCryopodItemObjectBytes(payload []byte) []byte {
	return syntheticCryopodItemObjectBytesWithPayloads(payload)
}

func syntheticCryopodItemObjectBytesWithPayloads(payloads ...[]byte) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000047))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeCustomItemDatasProperty(&buf, payloads...)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func truncatedDinoObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticCryopodDinoPayload(t *testing.T, dinoID uuid.UUID, statusID uuid.UUID) []byte {
	t.Helper()

	var decoded bytes.Buffer
	_ = binary.Write(&decoded, binary.LittleEndian, int32(0))
	_ = binary.Write(&decoded, binary.LittleEndian, int32(0))
	_ = binary.Write(&decoded, binary.LittleEndian, uint32(2))
	dinoOffsetPos := writeCryopodEmbeddedObjectHeader(&decoded, dinoID, "Dino", []string{"D0"})
	statusOffsetPos := writeCryopodEmbeddedObjectHeader(&decoded, statusID, "Status", []string{"S0"})

	dinoPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000001, 1001)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000002, 2002)
	writeCryopodEmbeddedNameDoubleProperty(&decoded, 0x10000003, 42)
	writeCryopodEmbeddedNone(&decoded)
	statusPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000005, 12)
	writeCryopodEmbeddedNone(&decoded)

	binary.LittleEndian.PutUint32(decoded.Bytes()[dinoOffsetPos:dinoOffsetPos+4], uint32(dinoPropsOffset))
	binary.LittleEndian.PutUint32(decoded.Bytes()[statusOffsetPos:statusOffsetPos+4], uint32(statusPropsOffset))

	namesOffset := decoded.Len()
	_ = binary.Write(&decoded, binary.LittleEndian, uint32(7))
	writeArkString(&decoded, "None")
	writeArkString(&decoded, "DinoID1")
	writeArkString(&decoded, "DinoID2")
	writeArkString(&decoded, "TamedTimeStamp")
	writeArkString(&decoded, "IntProperty")
	writeArkString(&decoded, "BaseCharacterLevel")
	writeArkString(&decoded, "DoubleProperty")

	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(decoded.Bytes()); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}

	var payload bytes.Buffer
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0x0407))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(decoded.Len()))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(namesOffset))
	payload.Write(compressed.Bytes())
	return payload.Bytes()
}

func syntheticLegacyCryopodPayload() []byte {
	var payload bytes.Buffer
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0x0406))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0))
	return payload.Bytes()
}

func syntheticCryopodSaddlePayload() []byte {
	var payload bytes.Buffer
	_ = binary.Write(&payload, binary.LittleEndian, uint32(8))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(7))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0))
	writePathObjectProperty(&payload, "ItemArchetype", "BlueprintGeneratedClass /Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C")
	writeArkString(&payload, "None")
	return payload.Bytes()
}

func syntheticDinoDetailObjectBytes() []byte {
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, 1001)
	writeIntProperty(&buf, 0x10000016, 2002)
	writeBoolProperty(&buf, 0x10000017, true)
	writeDoubleProperty(&buf, 0x10000018, 42)
	writeObjectReferenceProperty(&buf, 0x10000023, inventoryID)
	writeStringProperty(&buf, 0x10000024, "Blue")
	writeBoolProperty(&buf, 0x10000025, true)
	writePositionedInt8Property(&buf, 0x1000002e, 0, 11)
	writePositionedInt8Property(&buf, 0x1000002e, 3, 44)
	writePositionedNameProperty(&buf, 0x1000002f, 1, 0x10000034)
	writePositionedNameProperty(&buf, 0x1000002f, 4, 0x10000031)
	writeStringProperty(&buf, 0x10000030, "\nTheIsland")
	writeNameArrayProperty(&buf, 0x1000003d, []uint32{0x1000003e, 0x1000003f})
	writeStringProperty(&buf, 0x10000026, "Porters")
	writeIntProperty(&buf, 0x10000027, 555)
	writeStringProperty(&buf, 0x10000028, "Porters")
	writeStringProperty(&buf, 0x10000029, "Survivor")
	writeStringProperty(&buf, 0x1000002a, "Survivor")
	writeStringProperty(&buf, 0x1000002b, "eos-survivor")
	writeIntProperty(&buf, 0x1000002c, 42)
	writeIntProperty(&buf, 0x10000009, 555)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticDinoBabyObjectBytes(age float32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, 1001)
	writeIntProperty(&buf, 0x10000016, 2002)
	writeBoolProperty(&buf, 0x10000021, true)
	writeFloatProperty(&buf, 0x1000002d, age)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticDinoObjectBytesWithFlags(id1 int32, id2 int32, isFemale bool, isDead bool, isBaby bool, isTamed bool) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, id1)
	writeIntProperty(&buf, 0x10000016, id2)
	writeBoolProperty(&buf, 0x10000017, isFemale)
	writeBoolProperty(&buf, 0x10000020, isDead)
	writeBoolProperty(&buf, 0x10000021, isBaby)
	if isTamed {
		writeDoubleProperty(&buf, 0x10000018, 42)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func writeObjectReferenceProperty(buf *bytes.Buffer, name uint32, id uuid.UUID) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001f))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(18))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, int16(0))
	buf.Write(id[:])
}

func writeCustomItemDatasProperty(buf *bytes.Buffer, payloads ...[]byte) {
	elements := make([][]byte, 0, len(payloads))
	for _, payload := range payloads {
		byteValues := make([]byte, len(payload))
		copy(byteValues, payload)

		var bytesElement bytes.Buffer
		writeByteArrayProperty(&bytesElement, 0x1000004f, 0x10000050, byteValues)
		_ = binary.Write(&bytesElement, binary.LittleEndian, uint32(0x10000004))
		_ = binary.Write(&bytesElement, binary.LittleEndian, int32(0))
		elements = append(elements, bytesElement.Bytes())
	}

	var customDataBytes bytes.Buffer
	writeStructArrayProperty(&customDataBytes, 0x1000004d, 0x10000049, 0x1000004e, elements)
	_ = binary.Write(&customDataBytes, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&customDataBytes, binary.LittleEndian, int32(0))

	var customItemData bytes.Buffer
	writeStructProperty(&customItemData, 0x1000004b, 0x10000049, 0x1000004c, customDataBytes.Bytes())
	_ = binary.Write(&customItemData, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&customItemData, binary.LittleEndian, int32(0))

	writeStructArrayProperty(buf, 0x10000048, 0x10000049, 0x1000004a, [][]byte{customItemData.Bytes()})
}

func writePathObjectProperty(buf *bytes.Buffer, name string, path string) {
	writeArkString(buf, name)
	writeArkString(buf, "ObjectProperty")
	var body bytes.Buffer
	_ = binary.Write(&body, binary.LittleEndian, int32(1))
	writeArkString(&body, path)
	_ = binary.Write(buf, binary.LittleEndian, int32(body.Len()))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	buf.Write(body.Bytes())
}

func writeStructProperty(buf *bytes.Buffer, name uint32, structProperty uint32, structType uint32, body []byte) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, structProperty)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
}

func writeStructArrayProperty(buf *bytes.Buffer, name uint32, structProperty uint32, structType uint32, elements [][]byte) {
	bodySize := 4
	for _, element := range elements {
		bodySize += len(element)
	}
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001e))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	_ = binary.Write(buf, binary.LittleEndian, structProperty)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(bodySize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(elements)))
	for _, element := range elements {
		buf.Write(element)
	}
}

func writeByteArrayProperty(buf *bytes.Buffer, name uint32, byteProperty uint32, values []byte) {
	bodySize := 4 + len(values)
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001e))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	_ = binary.Write(buf, binary.LittleEndian, byteProperty)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(bodySize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	buf.Write(values)
}

func writeDoubleProperty(buf *bytes.Buffer, name uint32, value float64) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000019))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeCryopodEmbeddedObjectHeader(buf *bytes.Buffer, id uuid.UUID, className string, names []string) int {
	buf.Write(id[:])
	writeArkString(buf, className)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(names)))
	for _, name := range names {
		writeArkString(buf, name)
	}
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	offsetPos := buf.Len()
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	return offsetPos
}

func writeCryopodEmbeddedNameIntProperty(buf *bytes.Buffer, nameID uint32, value int32) {
	writeCryopodEmbeddedName(buf, nameID)
	writeCryopodEmbeddedName(buf, 0x10000004)
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeCryopodEmbeddedNameDoubleProperty(buf *bytes.Buffer, nameID uint32, value float64) {
	writeCryopodEmbeddedName(buf, nameID)
	writeCryopodEmbeddedName(buf, 0x10000006)
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeCryopodEmbeddedNone(buf *bytes.Buffer) {
	writeCryopodEmbeddedName(buf, 0x10000000)
}

func writeCryopodEmbeddedName(buf *bytes.Buffer, nameID uint32) {
	_ = binary.Write(buf, binary.LittleEndian, nameID)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
}

func writePositionedInt8Property(buf *bytes.Buffer, name uint32, position int32, value int8) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000032))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	_ = binary.Write(buf, binary.LittleEndian, int32(position))
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(byte(value))
}

func writePositionedNameProperty(buf *bytes.Buffer, name uint32, position int32, valueName uint32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000033))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(position))
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, valueName)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
}

func writePositionedIntProperty(buf *bytes.Buffer, name uint32, position int32, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000003))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writePositionedFloatProperty(buf *bytes.Buffer, name uint32, position int32, value float32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000000a))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeNameArrayProperty(buf *bytes.Buffer, name uint32, values []uint32) {
	dataSize := uint32(4 + len(values)*8)
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001e))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(dataSize))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000033))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, dataSize)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		_ = binary.Write(buf, binary.LittleEndian, value)
		_ = binary.Write(buf, binary.LittleEndian, int32(0))
	}
}
