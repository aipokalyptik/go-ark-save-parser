package arkobject

import (
	"math"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestDinoFromObjectReadsCoreFields(t *testing.T) {
	statusID := "00112233-4455-6677-8899-aabbccddeeff"
	inventoryID := "aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"
	object := &GameObject{
		UUID:      uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		Properties: []arkproperty.Property{
			{Name: "DinoID1", Type: arkproperty.TypeUInt32, Value: uint32(1001)},
			{Name: "DinoID2", Type: arkproperty.TypeUInt32, Value: uint32(2002)},
			{Name: "bIsFemale", Type: arkproperty.TypeBool, Value: true},
			{Name: "bIsDead", Type: arkproperty.TypeBool, Value: false},
			{Name: "bIsBaby", Type: arkproperty.TypeBool, Value: true},
			{Name: "BabyAge", Type: arkproperty.TypeFloat, Value: float32(0.42)},
			{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)},
			{Name: "MyCharacterStatusComponent", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{
				Type:  arkproperty.ObjectReferenceUUID,
				Value: statusID,
			}},
			{Name: "MyInventoryComponent", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{
				Type:  arkproperty.ObjectReferenceUUID,
				Value: inventoryID,
			}},
			{Name: "TamedName", Type: arkproperty.TypeString, Value: "Blue"},
			{Name: "bNeutered", Type: arkproperty.TypeBool, Value: true},
			{Name: "ColorSetIndices", Type: arkproperty.TypeInt8, Position: 0, Value: int8(11)},
			{Name: "ColorSetIndices", Type: arkproperty.TypeInt8, Position: 3, Value: int8(44)},
			{Name: "ColorSetNames", Type: arkproperty.TypeName, Position: 1, Value: "Blue"},
			{Name: "ColorSetNames", Type: arkproperty.TypeName, Position: 4, Value: "Black"},
			{Name: "UploadedFromServerName", Type: arkproperty.TypeString, Value: "\nTheIsland"},
			{Name: "TribeName", Type: arkproperty.TypeString, Value: "Porters"},
			{Name: "TamingTeamID", Type: arkproperty.TypeInt, Value: int32(555)},
			{Name: "TamerString", Type: arkproperty.TypeString, Value: "Porters"},
			{Name: "OwningPlayerName", Type: arkproperty.TypeString, Value: "Survivor"},
			{Name: "ImprinterName", Type: arkproperty.TypeString, Value: "Survivor"},
			{Name: "ImprinterPlayerUniqueNetId", Type: arkproperty.TypeString, Value: "eos-survivor"},
			{Name: "OwningPlayerID", Type: arkproperty.TypeInt, Value: int32(42)},
			{Name: "TargetingTeam", Type: arkproperty.TypeInt, Value: int32(555)},
			{Name: "GeneTraits", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeString,
				Values:      []any{"MutableMelee[2]", "Robust"},
			}},
			{Name: "SavedBaseWorldLocation", Type: arkproperty.TypeStruct, Value: ActorTransform{X: 1, Y: 2, Z: 3}},
		},
	}

	dino := DinoFromObject(object, nil)

	if dino.UUID != object.UUID || dino.Blueprint != object.Blueprint {
		t.Fatalf("Dino identity = %#v", dino)
	}
	if dino.ID1 != 1001 || dino.ID2 != 2002 || !dino.IsFemale || dino.IsDead || !dino.IsBaby || !dino.IsTamed {
		t.Fatalf("Dino flags/id = %#v", dino)
	}
	if math.Abs(dino.MaturationPercent-42) > 0.0001 || dino.BabyStage != BabyStageJuvenile {
		t.Fatalf("Baby maturation fields = %#v", dino)
	}
	if dino.StatusComponentUUID == nil || dino.StatusComponentUUID.String() != statusID {
		t.Fatalf("StatusComponentUUID = %v, want %s", dino.StatusComponentUUID, statusID)
	}
	if dino.InventoryUUID == nil || dino.InventoryUUID.String() != inventoryID {
		t.Fatalf("InventoryUUID = %v, want %s", dino.InventoryUUID, inventoryID)
	}
	if dino.TamedName != "Blue" || !dino.IsNeutered {
		t.Fatalf("Tamed detail fields = %#v", dino)
	}
	wantColorIndices := [6]int{11, 0, 0, 44, 0, 0}
	if dino.ColorSetIndices != wantColorIndices {
		t.Fatalf("ColorSetIndices = %#v, want %#v", dino.ColorSetIndices, wantColorIndices)
	}
	wantColorNames := [6]string{"None", "Blue", "None", "None", "Black", "None"}
	if dino.ColorSetNames != wantColorNames {
		t.Fatalf("ColorSetNames = %#v, want %#v", dino.ColorSetNames, wantColorNames)
	}
	if dino.UploadedFromServerName != "TheIsland" {
		t.Fatalf("UploadedFromServerName = %q, want TheIsland", dino.UploadedFromServerName)
	}
	if dino.Owner.TribeName != "Porters" || dino.Owner.TamerTribeID != 555 || dino.Owner.TamerString != "Porters" {
		t.Fatalf("Dino owner tribe fields = %#v", dino.Owner)
	}
	if dino.Owner.PlayerName != "Survivor" || dino.Owner.PlayerID != 42 || dino.Owner.ImprinterUniqueID != "eos-survivor" {
		t.Fatalf("Dino owner player fields = %#v", dino.Owner)
	}
	if len(dino.GeneTraits) != 2 || dino.GeneTraits[0] != "MutableMelee[2]" {
		t.Fatalf("GeneTraits = %#v", dino.GeneTraits)
	}
	if len(dino.ParsedGeneTraits) != 2 || dino.ParsedGeneTraits[0].Name != "MutableMelee" || dino.ParsedGeneTraits[0].Level != 2 {
		t.Fatalf("ParsedGeneTraits = %#v", dino.ParsedGeneTraits)
	}
	if dino.ParsedGeneTraits[1].Name != "Robust" || dino.ParsedGeneTraits[1].Level != 0 || dino.ParsedGeneTraits[1].Raw != "Robust" {
		t.Fatalf("ParsedGeneTraits fallback = %#v", dino.ParsedGeneTraits[1])
	}
	if dino.Location == nil || dino.Location.X != 1 || dino.Location.Z != 3 {
		t.Fatalf("Location = %#v", dino.Location)
	}
	if dino.ShortName() != "Raptor" {
		t.Fatalf("ShortName() = %q, want Raptor", dino.ShortName())
	}
}

func TestDinoFromObjectPrefersSaveContextLocation(t *testing.T) {
	object := &GameObject{UUID: uuid.MustParse("11112222-3333-4444-5555-666677778888")}
	location := ActorTransform{X: 10, Y: 20, Z: 30}

	dino := DinoFromObject(object, &location)

	if dino.Location == nil || dino.Location.X != 10 || dino.Location.Y != 20 {
		t.Fatalf("Location = %#v, want save context location", dino.Location)
	}
}

func TestDinoFromObjectCalculatesTamedGenerationFromAncestorArrays(t *testing.T) {
	object := &GameObject{
		UUID: uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Properties: []arkproperty.Property{
			{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)},
			{Name: "DinoAncestors", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeStruct,
				Values:      []any{arkproperty.Container{}, arkproperty.Container{}},
			}},
			{Name: "DinoAncestorsMale", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeStruct,
				Values:      []any{arkproperty.Container{}},
			}},
		},
	}

	dino := DinoFromObject(object, nil)

	if dino.Generation != 3 {
		t.Fatalf("Generation = %d, want 3", dino.Generation)
	}
}

func TestDinoFromObjectReadsAncestorIDsInUpstreamOrder(t *testing.T) {
	object := &GameObject{
		UUID: uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Properties: []arkproperty.Property{
			{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)},
			{Name: "DinoAncestors", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeStruct,
				Values: []any{
					arkproperty.Container{Properties: []arkproperty.Property{
						{Name: "FemaleDinoID1", Type: arkproperty.TypeUInt32, Value: uint32(11)},
						{Name: "FemaleDinoID2", Type: arkproperty.TypeUInt32, Value: uint32(12)},
						{Name: "MaleDinoID1", Type: arkproperty.TypeUInt32, Value: uint32(21)},
						{Name: "MaleDinoID2", Type: arkproperty.TypeUInt32, Value: uint32(22)},
					}},
				},
			}},
			{Name: "DinoAncestorsMale", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeStruct,
				Values: []any{
					arkproperty.Container{Properties: []arkproperty.Property{
						{Name: "FemaleDinoID1", Type: arkproperty.TypeUInt32, Value: uint32(31)},
						{Name: "FemaleDinoID2", Type: arkproperty.TypeUInt32, Value: uint32(32)},
						{Name: "MaleDinoID1", Type: arkproperty.TypeUInt32, Value: uint32(41)},
						{Name: "MaleDinoID2", Type: arkproperty.TypeUInt32, Value: uint32(42)},
					}},
				},
			}},
		},
	}

	dino := DinoFromObject(object, nil)

	want := []DinoID{{ID1: 11, ID2: 12}, {ID1: 21, ID2: 22}, {ID1: 31, ID2: 32}, {ID1: 41, ID2: 42}}
	if len(dino.AncestorIDs) != len(want) {
		t.Fatalf("AncestorIDs length = %d, want %d: %#v", len(dino.AncestorIDs), len(want), dino.AncestorIDs)
	}
	for i := range want {
		if dino.AncestorIDs[i] != want[i] {
			t.Fatalf("AncestorIDs[%d] = %#v, want %#v", i, dino.AncestorIDs[i], want[i])
		}
	}
}

func TestDinoFromObjectSetsFirstGenerationForTamedDinosWithoutAncestors(t *testing.T) {
	object := &GameObject{
		UUID: uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Properties: []arkproperty.Property{
			{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)},
		},
	}

	dino := DinoFromObject(object, nil)

	if dino.Generation != 1 {
		t.Fatalf("Generation = %d, want 1", dino.Generation)
	}
}

func TestDinoFromObjectLeavesWildGenerationUnset(t *testing.T) {
	object := &GameObject{UUID: uuid.MustParse("11112222-3333-4444-5555-666677778888")}

	dino := DinoFromObject(object, nil)

	if dino.Generation != 0 {
		t.Fatalf("Generation = %d, want 0", dino.Generation)
	}
}

func TestDinoStatsFromObjectReadsPositionedStatusFields(t *testing.T) {
	object := &GameObject{
		UUID: uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff"),
		Properties: []arkproperty.Property{
			{Name: "BaseCharacterLevel", Type: arkproperty.TypeInt, Value: int32(12)},
			{Name: "NumberOfLevelUpPointsApplied", Type: arkproperty.TypeInt, Position: 0, Value: int32(5)},
			{Name: "NumberOfLevelUpPointsApplied", Type: arkproperty.TypeInt, Position: 7, Value: int32(3)},
			{Name: "NumberOfLevelUpPointsAppliedTamed", Type: arkproperty.TypeInt, Position: 8, Value: int32(2)},
			{Name: "NumberOfMutationsAppliedTamed", Type: arkproperty.TypeInt, Position: 0, Value: int32(1)},
			{Name: "CurrentStatusValues", Type: arkproperty.TypeFloat, Position: 0, Value: float32(1234.5)},
			{Name: "CurrentStatusValues", Type: arkproperty.TypeFloat, Position: 7, Value: float32(321.25)},
			{Name: "DinoImprintingQuality", Type: arkproperty.TypeFloat, Value: float32(0.875)},
		},
	}

	stats := DinoStatsFromObject(object)

	if stats.BaseLevel != 12 || stats.CurrentLevel != 12 {
		t.Fatalf("levels = base %d current %d, want 12 and 12", stats.BaseLevel, stats.CurrentLevel)
	}
	if stats.BaseStatPoints.Health != 5 || stats.BaseStatPoints.Weight != 3 {
		t.Fatalf("base points = %#v", stats.BaseStatPoints)
	}
	if stats.AddedStatPoints.MeleeDamage != 2 || stats.MutatedStatPoints.Health != 1 {
		t.Fatalf("tamed/mutated points = %#v / %#v", stats.AddedStatPoints, stats.MutatedStatPoints)
	}
	if math.Abs(stats.StatValues.Health-1234.5) > 0.0001 || math.Abs(stats.StatValues.Weight-321.25) > 0.0001 {
		t.Fatalf("stat values = %#v", stats.StatValues)
	}
	if math.Abs(stats.ImprintingPercent-87.5) > 0.0001 {
		t.Fatalf("imprinting = %f, want 87.5", stats.ImprintingPercent)
	}
	if got := stats.Points(DinoStatHealth); got != 6 {
		t.Fatalf("Points(Health) = %d, want 6", got)
	}
	if got := stats.Points(DinoStatWeight); got != 3 {
		t.Fatalf("Points(Weight) = %d, want 3", got)
	}
	if got := stats.Points(DinoStatHealth, StatScopeBase); got != 5 {
		t.Fatalf("Points(Health, base) = %d, want 5", got)
	}
	if got := stats.Points(DinoStatHealth, StatScopeMutated); got != 1 {
		t.Fatalf("Points(Health, mutated) = %d, want 1", got)
	}
	statsAtLeast := stats.StatsAtLeast(6)
	if len(statsAtLeast) != 1 || statsAtLeast[0] != DinoStatHealth {
		t.Fatalf("StatsAtLeast(6) = %#v, want health only", statsAtLeast)
	}
	if stat, points, ok := stats.BestStat(); !ok || stat != DinoStatHealth || points != 6 {
		t.Fatalf("BestStat() = %v, %d, %v; want health, 6, true", stat, points, ok)
	}
	if stat, points, ok := stats.BestStat(StatScopeBase); !ok || stat != DinoStatHealth || points != 5 {
		t.Fatalf("BestStat(base) = %v, %d, %v; want health, 5, true", stat, points, ok)
	}
	if got := stats.TotalMutations(); got != 1 {
		t.Fatalf("TotalMutations() = %d, want 1", got)
	}
}

func TestDinoStatsBestStatBreaksTiesByStableStatOrder(t *testing.T) {
	stats := DinoStats{
		BaseStatPoints: DinoStatPoints{
			Stamina: 5,
			Weight:  5,
		},
	}

	stat, points, ok := stats.BestStat(StatScopeBase)
	if !ok || stat != DinoStatStamina || points != 5 {
		t.Fatalf("BestStat(base) = %v, %d, %v; want stamina, 5, true", stat, points, ok)
	}
}

func TestDinoStatsTotalMutationsSumsAllStatMutationPoints(t *testing.T) {
	stats := DinoStats{
		MutatedStatPoints: DinoStatPoints{
			Health:      1,
			Weight:      2,
			MeleeDamage: 3,
		},
	}

	if got := stats.TotalMutations(); got != 6 {
		t.Fatalf("TotalMutations() = %d, want 6", got)
	}
}
