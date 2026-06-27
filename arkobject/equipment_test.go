package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestEquipmentItemFromObjectReadsBaseEquipmentFields(t *testing.T) {
	object := &GameObject{
		UUID:      uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"),
		Blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
		Properties: []arkproperty.Property{
			{Name: "ItemQuantity", Type: arkproperty.TypeInt, Value: int32(1)},
			{Name: "bIsBlueprint", Type: arkproperty.TypeBool, Value: true},
			{Name: "bEquippedItem", Type: arkproperty.TypeBool, Value: false},
			{Name: "ItemRating", Type: arkproperty.TypeFloat, Value: float32(7.5)},
			{Name: "ItemQualityIndex", Type: arkproperty.TypeByte, Value: byte(3)},
			{Name: "SavedDurability", Type: arkproperty.TypeFloat, Value: float32(0.75)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatDurability), Value: uint16(1000)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatDamage), Value: uint16(1234)},
			{Name: "CrafterCharacterName", Type: arkproperty.TypeString, Value: "Survivor"},
			{Name: "CrafterTribeName", Type: arkproperty.TypeString, Value: "Porters"},
		},
	}

	item := EquipmentItemFromObject(object, EquipmentWeapon)

	if item.UUID != object.UUID || item.Blueprint != object.Blueprint || item.Kind != EquipmentWeapon {
		t.Fatalf("EquipmentItem identity = %#v", item)
	}
	if item.Quantity != 1 || !item.IsBlueprint || item.IsEquipped {
		t.Fatalf("EquipmentItem inventory flags = %#v", item)
	}
	if item.Rating != 7.5 || item.Quality != 3 || item.CurrentDurability != 0.75 {
		t.Fatalf("EquipmentItem equipment fields = %#v", item)
	}
	if item.Stats.Internal[EquipmentStatDamage] != 1234 || item.Stats.Internal[EquipmentStatDurability] != 1000 {
		t.Fatalf("EquipmentItem internal stats = %#v", item.Stats.Internal)
	}
	if item.Stats.Damage != 112.3 || item.Stats.Durability != 62.5 {
		t.Fatalf("EquipmentItem actual stats = %#v", item.Stats)
	}
	if item.Crafter == nil || item.Crafter.CharacterName != "Survivor" || item.Crafter.TribeName != "Porters" {
		t.Fatalf("EquipmentItem.Crafter = %#v", item.Crafter)
	}
	if !item.IsCrafted() {
		t.Fatalf("EquipmentItem.IsCrafted() = false, want true")
	}
	if got := item.AverageStat(); got != 1117 {
		t.Fatalf("EquipmentItem.AverageStat() = %f, want 1117", got)
	}
	if stats := item.ImplementedStats(); len(stats) != 2 || stats[0] != EquipmentStatDurability || stats[1] != EquipmentStatDamage {
		t.Fatalf("EquipmentItem.ImplementedStats() = %#v, want durability and damage", stats)
	}
}

func TestEquipmentItemFromObjectReadsArmorStats(t *testing.T) {
	object := &GameObject{
		UUID:      uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"),
		Blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'",
		Properties: []arkproperty.Property{
			{Name: "ItemQuantity", Type: arkproperty.TypeInt, Value: int32(1)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatArmor), Value: uint16(1000)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatHypothermalResistance), Value: uint16(500)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatHyperthermalResistance), Value: uint16(200)},
		},
	}

	item := EquipmentItemFromObject(object, EquipmentArmor)

	if item.Stats.Internal[EquipmentStatArmor] != 1000 || item.Stats.Internal[EquipmentStatHypothermalResistance] != 500 {
		t.Fatalf("EquipmentItem internal armor stats = %#v", item.Stats.Internal)
	}
	if item.Stats.Armor != 12 || item.Stats.HypothermalResistance != 8.8 || item.Stats.HyperthermalResistance != 15.6 {
		t.Fatalf("EquipmentItem armor stats = %#v", item.Stats)
	}
	if got := item.AverageStat(); got != 425 {
		t.Fatalf("EquipmentItem.AverageStat() = %f, want 425", got)
	}
	if stats := item.ImplementedStats(); len(stats) != 4 ||
		stats[0] != EquipmentStatDurability ||
		stats[1] != EquipmentStatArmor ||
		stats[2] != EquipmentStatHypothermalResistance ||
		stats[3] != EquipmentStatHyperthermalResistance {
		t.Fatalf("EquipmentItem.ImplementedStats() = %#v, want durability, armor, hypo, hyper", stats)
	}
}

func TestEquipmentItemFromObjectUsesUpstreamDefaultStatTables(t *testing.T) {
	tests := []struct {
		name                   string
		blueprint              string
		kind                   EquipmentKind
		wantDurability         float64
		wantArmor              float64
		wantHypothermal        float64
		wantHyperthermal       float64
		wantAverageStat        float64
		wantImplementedStats   int
		wantInternalDurability uint16
	}{
		{
			name:                   "compound bow weapon",
			blueprint:              "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponCompoundBow.PrimalItem_WeaponCompoundBow_C'",
			kind:                   EquipmentWeapon,
			wantDurability:         68.75,
			wantAverageStat:        1000,
			wantImplementedStats:   2,
			wantInternalDurability: 1000,
		},
		{
			name:                 "chitin armor",
			blueprint:            "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Chitin/PrimalItemArmor_ChitinShirt.PrimalItemArmor_ChitinShirt_C'",
			kind:                 EquipmentArmor,
			wantDurability:       62.5,
			wantArmor:            60,
			wantHypothermal:      11,
			wantHyperthermal:     -5.2,
			wantAverageStat:      675,
			wantImplementedStats: 4,
		},
		{
			name:                 "desert armor",
			blueprint:            "Blueprint'/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothGogglesHelmet.PrimalItemArmor_DesertClothGogglesHelmet_C'",
			kind:                 EquipmentArmor,
			wantDurability:       56.25,
			wantArmor:            48,
			wantHypothermal:      5.5,
			wantHyperthermal:     31.2,
			wantAverageStat:      675,
			wantImplementedStats: 4,
		},
		{
			name:                 "metal shield",
			blueprint:            "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_MetalShield.PrimalItemArmor_MetalShield_C'",
			kind:                 EquipmentShield,
			wantDurability:       1562.5,
			wantArmor:            1.2,
			wantAverageStat:      1000,
			wantImplementedStats: 1,
		},
		{
			name:                 "tek saddle",
			blueprint:            "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_RexSaddle_Tek.PrimalItemArmor_RexSaddle_Tek_C'",
			kind:                 EquipmentSaddle,
			wantDurability:       150,
			wantArmor:            54,
			wantAverageStat:      1000,
			wantImplementedStats: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object := &GameObject{
				UUID:      uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"),
				Blueprint: tt.blueprint,
				Properties: []arkproperty.Property{
					{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatDurability), Value: uint16(1000)},
					{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatArmor), Value: uint16(1000)},
					{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatDamage), Value: uint16(1000)},
					{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatHypothermalResistance), Value: uint16(500)},
					{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatHyperthermalResistance), Value: uint16(200)},
				},
			}

			item := EquipmentItemFromObject(object, tt.kind)
			if item.Stats.Durability != tt.wantDurability {
				t.Fatalf("Durability = %f, want %f", item.Stats.Durability, tt.wantDurability)
			}
			if item.Stats.Armor != tt.wantArmor {
				t.Fatalf("Armor = %f, want %f", item.Stats.Armor, tt.wantArmor)
			}
			if item.Stats.HypothermalResistance != tt.wantHypothermal {
				t.Fatalf("HypothermalResistance = %f, want %f", item.Stats.HypothermalResistance, tt.wantHypothermal)
			}
			if item.Stats.HyperthermalResistance != tt.wantHyperthermal {
				t.Fatalf("HyperthermalResistance = %f, want %f", item.Stats.HyperthermalResistance, tt.wantHyperthermal)
			}
			if got := item.AverageStat(); got != tt.wantAverageStat {
				t.Fatalf("AverageStat() = %f, want %f", got, tt.wantAverageStat)
			}
			if stats := item.ImplementedStats(); len(stats) != tt.wantImplementedStats {
				t.Fatalf("ImplementedStats() = %#v, want %d stats", stats, tt.wantImplementedStats)
			}
			if tt.wantInternalDurability != 0 && item.Stats.Internal[EquipmentStatDurability] != tt.wantInternalDurability {
				t.Fatalf("Internal durability = %d, want %d", item.Stats.Internal[EquipmentStatDurability], tt.wantInternalDurability)
			}
		})
	}
}

func TestEquipmentItemAverageStatUsesKindSpecificInternalStats(t *testing.T) {
	shield := EquipmentItem{
		Kind: EquipmentShield,
		Stats: EquipmentStats{Internal: map[EquipmentStat]uint16{
			EquipmentStatDurability: 900,
			EquipmentStatArmor:      100,
		}},
	}
	if got := shield.AverageStat(); got != 900 {
		t.Fatalf("shield AverageStat() = %f, want durability only", got)
	}

	saddle := EquipmentItem{
		Kind: EquipmentSaddle,
		Stats: EquipmentStats{Internal: map[EquipmentStat]uint16{
			EquipmentStatDurability: 900,
			EquipmentStatArmor:      100,
		}},
	}
	if got := saddle.AverageStat(); got != 500 {
		t.Fatalf("saddle AverageStat() = %f, want average durability and armor", got)
	}
}

func TestEquipmentItemIsCraftedRequiresValidCrafterMetadata(t *testing.T) {
	zero := EquipmentItem{}
	if zero.IsCrafted() {
		t.Fatalf("zero EquipmentItem IsCrafted() = true, want false")
	}
	emptyCrafter := EquipmentItem{InventoryItem: InventoryItem{Crafter: &ObjectCrafter{}}}
	if emptyCrafter.IsCrafted() {
		t.Fatalf("empty crafter IsCrafted() = true, want false")
	}
	namedCrafter := EquipmentItem{InventoryItem: InventoryItem{Crafter: &ObjectCrafter{CharacterName: "Survivor"}}}
	if !namedCrafter.IsCrafted() {
		t.Fatalf("named crafter IsCrafted() = false, want true")
	}
}
