package arkapi

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/google/uuid"
)

type DomainExport struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
	Items  any    `json:"items"`
}

type DinoInfo struct {
	UUID                   string              `json:"uuid"`
	Blueprint              string              `json:"blueprint"`
	ID1                    uint32              `json:"id1"`
	ID2                    uint32              `json:"id2"`
	IsFemale               bool                `json:"is_female"`
	IsTamed                bool                `json:"is_tamed"`
	IsBaby                 bool                `json:"is_baby"`
	IsDead                 bool                `json:"is_dead"`
	IsCryopodded           bool                `json:"is_cryopodded"`
	Generation             int                 `json:"generation,omitempty"`
	AncestorIDs            []DinoIDInfo        `json:"ancestor_ids,omitempty"`
	MaturationPercent      float64             `json:"maturation_percent,omitempty"`
	BabyStage              arkobject.BabyStage `json:"baby_stage,omitempty"`
	InventoryUUID          string              `json:"inventory_uuid,omitempty"`
	TamedName              string              `json:"tamed_name,omitempty"`
	IsNeutered             bool                `json:"is_neutered,omitempty"`
	ColorSetIndices        []int               `json:"color_set_indices,omitempty"`
	ColorSetNames          []string            `json:"color_set_names,omitempty"`
	UploadedFromServerName string              `json:"uploaded_from_server_name,omitempty"`
	Stats                  *DinoStatsInfo      `json:"stats,omitempty"`
	Owner                  DinoOwnerInfo       `json:"owner,omitempty"`
	GeneTraits             []string            `json:"gene_traits,omitempty"`
	ParsedGeneTraits       []GeneTraitInfo     `json:"parsed_gene_traits,omitempty"`
	Location               *LocationInfo       `json:"location,omitempty"`
}

type StructureInfo struct {
	UUID                          string        `json:"uuid"`
	Blueprint                     string        `json:"blueprint"`
	ID                            int32         `json:"id"`
	Owner                         OwnerInfo     `json:"owner"`
	MaxHealth                     float64       `json:"max_health"`
	CurrentHealth                 float64       `json:"current_health"`
	Location                      *LocationInfo `json:"location,omitempty"`
	InventoryUUID                 string        `json:"inventory_uuid,omitempty"`
	ItemCount                     int32         `json:"item_count,omitempty"`
	MaxItemCount                  int32         `json:"max_item_count,omitempty"`
	OpenSlots                     int32         `json:"open_slots,omitempty"`
	IsEmpty                       bool          `json:"is_empty,omitempty"`
	LinkedStructureUUIDs          []string      `json:"linked_structure_uuids,omitempty"`
	OriginalCreationTime          float64       `json:"original_creation_time,omitempty"`
	LastEnterStasisTime           float64       `json:"last_enter_stasis_time,omitempty"`
	HasResetDecayTime             bool          `json:"has_reset_decay_time,omitempty"`
	SavedWhenStasised             bool          `json:"saved_when_stasised,omitempty"`
	WasPlacementSnapped           bool          `json:"was_placement_snapped,omitempty"`
	LastInAllyRangeTimeSerialized float64       `json:"last_in_ally_range_time_serialized,omitempty"`
}

type EquipmentInfo struct {
	UUID               string              `json:"uuid"`
	Blueprint          string              `json:"blueprint"`
	Kind               string              `json:"kind"`
	Quantity           int32               `json:"quantity"`
	OwnerInventoryUUID string              `json:"owner_inventory_uuid,omitempty"`
	InCryopod          bool                `json:"in_cryopod,omitempty"`
	IsEquipped         bool                `json:"is_equipped"`
	IsBlueprint        bool                `json:"is_blueprint"`
	Rating             float64             `json:"rating"`
	Quality            int32               `json:"quality"`
	CurrentDurability  float64             `json:"current_durability"`
	IsCrafted          bool                `json:"is_crafted"`
	AverageStat        float64             `json:"average_stat,omitempty"`
	ImplementedStats   []string            `json:"implemented_stats,omitempty"`
	Stats              *EquipmentStatsInfo `json:"stats,omitempty"`
	Crafter            *CrafterInfo        `json:"crafter,omitempty"`
}

type StackableInfo struct {
	UUID               string `json:"uuid"`
	Blueprint          string `json:"blueprint"`
	Quantity           int32  `json:"quantity"`
	OwnerInventoryUUID string `json:"owner_inventory_uuid,omitempty"`
}

type BaseInfo struct {
	KeystoneUUID       string         `json:"keystone_uuid"`
	StructureUUIDs     []string       `json:"structure_uuids"`
	StructureCount     int            `json:"structure_count"`
	Owner              OwnerInfo      `json:"owner"`
	Location           *LocationInfo  `json:"location,omitempty"`
	AverageLocation    *LocationInfo  `json:"average_location,omitempty"`
	MapLocation        *MapCoordsInfo `json:"map_location,omitempty"`
	AverageMapLocation *MapCoordsInfo `json:"average_map_location,omitempty"`
	TurretCount        float64        `json:"turret_count"`
}

type LocationInfo struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Z          float64 `json:"z"`
	Pitch      float64 `json:"pitch"`
	Roll       float64 `json:"roll"`
	Yaw        float64 `json:"yaw"`
	Quaternion float64 `json:"quaternion"`
}

type MapCoordsInfo struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type OwnerInfo struct {
	PlayerID         int32  `json:"player_id,omitempty"`
	PlayerName       string `json:"player_name,omitempty"`
	TribeID          int32  `json:"tribe_id,omitempty"`
	TribeName        string `json:"tribe_name,omitempty"`
	OriginalPlacerID int32  `json:"original_placer_id,omitempty"`
}

type CrafterInfo struct {
	CharacterName string `json:"character_name,omitempty"`
	TribeName     string `json:"tribe_name,omitempty"`
}

type DinoOwnerInfo struct {
	PlayerID          int32  `json:"player_id,omitempty"`
	PlayerName        string `json:"player_name,omitempty"`
	TribeName         string `json:"tribe_name,omitempty"`
	TamerTribeID      int32  `json:"tamer_tribe_id,omitempty"`
	TamerString       string `json:"tamer_string,omitempty"`
	ImprinterName     string `json:"imprinter_name,omitempty"`
	ImprinterUniqueID string `json:"imprinter_unique_id,omitempty"`
	TargetTeam        int32  `json:"target_team,omitempty"`
}

type DinoStatsInfo struct {
	BaseLevel         int32              `json:"base_level"`
	CurrentLevel      int32              `json:"current_level"`
	BaseStatPoints    DinoStatPointsInfo `json:"base_stat_points"`
	AddedStatPoints   DinoStatPointsInfo `json:"added_stat_points"`
	MutatedStatPoints DinoStatPointsInfo `json:"mutated_stat_points"`
	StatValues        DinoStatValuesInfo `json:"stat_values"`
	ImprintingPercent float64            `json:"imprinting_percent,omitempty"`
}

type GeneTraitInfo struct {
	Raw   string `json:"raw"`
	Name  string `json:"name"`
	Level int    `json:"level,omitempty"`
}

type DinoIDInfo struct {
	ID1 uint32 `json:"id1"`
	ID2 uint32 `json:"id2"`
}

type EquipmentStatsInfo struct {
	Internal               map[string]uint16 `json:"internal,omitempty"`
	Damage                 float64           `json:"damage,omitempty"`
	Durability             float64           `json:"durability,omitempty"`
	Armor                  float64           `json:"armor,omitempty"`
	HypothermalResistance  float64           `json:"hypothermal_resistance,omitempty"`
	HyperthermalResistance float64           `json:"hyperthermal_resistance,omitempty"`
}

type DinoStatPointsInfo struct {
	Health        int32 `json:"health,omitempty"`
	Stamina       int32 `json:"stamina,omitempty"`
	Torpidity     int32 `json:"torpidity,omitempty"`
	Oxygen        int32 `json:"oxygen,omitempty"`
	Food          int32 `json:"food,omitempty"`
	Water         int32 `json:"water,omitempty"`
	Temperature   int32 `json:"temperature,omitempty"`
	Weight        int32 `json:"weight,omitempty"`
	MeleeDamage   int32 `json:"melee_damage,omitempty"`
	MovementSpeed int32 `json:"movement_speed,omitempty"`
	Fortitude     int32 `json:"fortitude,omitempty"`
	CraftingSpeed int32 `json:"crafting_speed,omitempty"`
}

type DinoStatValuesInfo struct {
	Health        float64 `json:"health,omitempty"`
	Stamina       float64 `json:"stamina,omitempty"`
	Torpidity     float64 `json:"torpidity,omitempty"`
	Oxygen        float64 `json:"oxygen,omitempty"`
	Food          float64 `json:"food,omitempty"`
	Water         float64 `json:"water,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
	Weight        float64 `json:"weight,omitempty"`
	MeleeDamage   float64 `json:"melee_damage,omitempty"`
	MovementSpeed float64 `json:"movement_speed,omitempty"`
	Fortitude     float64 `json:"fortitude,omitempty"`
	CraftingSpeed float64 `json:"crafting_speed,omitempty"`
}

func (j *JSONAPI) ExportDomain(domain string) (DomainExport, error) {
	switch domain {
	case "dinos":
		items, err := j.ExportDinos()
		return DomainExport{Domain: domain, Count: len(items), Items: items}, err
	case "structures":
		items, err := j.ExportStructures()
		return DomainExport{Domain: domain, Count: len(items), Items: items}, err
	case "equipment":
		items, err := j.ExportEquipment()
		return DomainExport{Domain: domain, Count: len(items), Items: items}, err
	case "stackables":
		items, err := j.ExportStackables()
		return DomainExport{Domain: domain, Count: len(items), Items: items}, err
	case "bases":
		items, err := j.ExportBases()
		return DomainExport{Domain: domain, Count: len(items), Items: items}, err
	default:
		return DomainExport{}, fmt.Errorf("unsupported export domain %q", domain)
	}
}

func (j *JSONAPI) ExportDomainJSON(domain string) ([]byte, error) {
	data, err := j.ExportDomain(domain)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}

func (j *JSONAPI) ExportDinos() ([]DinoInfo, error) {
	dinos, _, err := NewDino(j.save).AllWithFaults()
	if err != nil {
		return nil, err
	}
	out := make([]DinoInfo, 0, len(dinos))
	for _, id := range sortedUUIDKeys(dinos) {
		dino := dinos[id]
		out = append(out, DinoInfo{
			UUID:                   id.String(),
			Blueprint:              dino.Blueprint,
			ID1:                    dino.ID1,
			ID2:                    dino.ID2,
			IsFemale:               dino.IsFemale,
			IsTamed:                dino.IsTamed,
			IsBaby:                 dino.IsBaby,
			IsDead:                 dino.IsDead,
			IsCryopodded:           dino.IsCryopodded,
			Generation:             dino.Generation,
			AncestorIDs:            dinoIDInfos(dino.AncestorIDs),
			MaturationPercent:      dino.MaturationPercent,
			BabyStage:              dino.BabyStage,
			InventoryUUID:          optionalUUIDString(dino.InventoryUUID),
			TamedName:              dino.TamedName,
			IsNeutered:             dino.IsNeutered,
			ColorSetIndices:        intArrayFromFixed6(dino.ColorSetIndices),
			ColorSetNames:          stringArrayFromFixed6(dino.ColorSetNames),
			UploadedFromServerName: dino.UploadedFromServerName,
			Stats:                  dinoStatsInfo(dino.Stats),
			Owner:                  dinoOwnerInfo(dino.Owner),
			GeneTraits:             dino.GeneTraits,
			ParsedGeneTraits:       geneTraitInfos(dino.ParsedGeneTraits),
			Location:               locationInfo(dino.Location),
		})
	}
	return out, nil
}

func (j *JSONAPI) ExportStructures() ([]StructureInfo, error) {
	structures, _, err := NewStructure(j.save).AllWithFaults()
	if err != nil {
		return nil, err
	}
	out := make([]StructureInfo, 0, len(structures))
	for _, id := range sortedUUIDKeys(structures) {
		structure := structures[id]
		out = append(out, StructureInfo{
			UUID:                          id.String(),
			Blueprint:                     structure.Blueprint,
			ID:                            structure.ID,
			Owner:                         ownerInfo(structure.Owner),
			MaxHealth:                     structure.MaxHealth,
			CurrentHealth:                 structure.CurrentHealth,
			Location:                      locationInfo(structure.Location),
			InventoryUUID:                 optionalUUIDString(structure.InventoryUUID),
			ItemCount:                     structure.ItemCount,
			MaxItemCount:                  structure.MaxItemCount,
			OpenSlots:                     structure.OpenSlots(),
			IsEmpty:                       structure.IsEmpty(),
			LinkedStructureUUIDs:          sortedUUIDStrings(structure.LinkedStructureUUIDs),
			OriginalCreationTime:          structure.OriginalCreationTime,
			LastEnterStasisTime:           structure.LastEnterStasisTime,
			HasResetDecayTime:             structure.HasResetDecayTime,
			SavedWhenStasised:             structure.SavedWhenStasised,
			WasPlacementSnapped:           structure.WasPlacementSnapped,
			LastInAllyRangeTimeSerialized: structure.LastInAllyRangeTimeSerialized,
		})
	}
	return out, nil
}

func (j *JSONAPI) ExportEquipment() ([]EquipmentInfo, error) {
	equipment, err := NewEquipment(j.save).All()
	if err != nil {
		return nil, err
	}
	out := make([]EquipmentInfo, 0, len(equipment))
	for _, id := range sortedUUIDKeys(equipment) {
		out = append(out, equipmentInfo(id, equipment[id], false))
	}
	cryopodSaddles, _, err := NewDino(j.save).SaddlesFromCryopodsWithFaults()
	if err != nil {
		return nil, err
	}
	for _, id := range sortedUUIDKeys(cryopodSaddles) {
		out = append(out, equipmentInfo(id, cryopodSaddles[id], true))
	}
	return out, nil
}

func equipmentInfo(id uuid.UUID, item arkobject.EquipmentItem, inCryopod bool) EquipmentInfo {
	info := EquipmentInfo{
		UUID:               id.String(),
		Blueprint:          item.Blueprint,
		Kind:               string(item.Kind),
		Quantity:           item.Quantity,
		OwnerInventoryUUID: optionalUUIDString(item.OwnerInventory),
		InCryopod:          inCryopod,
		IsEquipped:         item.IsEquipped,
		IsBlueprint:        item.IsBlueprint,
		Rating:             item.Rating,
		Quality:            item.Quality,
		CurrentDurability:  item.CurrentDurability,
		IsCrafted:          item.IsCrafted(),
		AverageStat:        item.AverageStat(),
		ImplementedStats:   equipmentStatNames(item.ImplementedStats()),
		Stats:              equipmentStatsInfo(item.Stats),
		Crafter:            crafterInfo(item.Crafter),
	}
	info.sanitize()
	return info
}

func (j *JSONAPI) ExportStackables() ([]StackableInfo, error) {
	stackables, err := NewStackable(j.save).All()
	if err != nil {
		return nil, err
	}
	out := make([]StackableInfo, 0, len(stackables))
	for _, id := range sortedUUIDKeys(stackables) {
		item := stackables[id]
		out = append(out, StackableInfo{
			UUID:               id.String(),
			Blueprint:          item.Blueprint,
			Quantity:           item.Quantity,
			OwnerInventoryUUID: optionalUUIDString(item.OwnerInventory),
		})
	}
	return out, nil
}

func (j *JSONAPI) ExportBases() ([]BaseInfo, error) {
	baseAPI := NewBase(j.save, "")
	bases, err := baseAPI.All()
	if err != nil {
		return nil, err
	}
	out := make([]BaseInfo, 0, len(bases))
	for _, base := range bases {
		out = append(out, BaseInfo{
			KeystoneUUID:       base.KeystoneUUID.String(),
			StructureUUIDs:     sortedBaseStructureUUIDs(base),
			StructureCount:     base.StructureCount,
			Owner:              ownerInfo(base.Owner),
			Location:           locationInfo(base.Location),
			AverageLocation:    locationInfo(base.AverageLocation),
			MapLocation:        mapCoordsInfo(base.Location, baseAPI.mapName),
			AverageMapLocation: mapCoordsInfo(base.AverageLocation, baseAPI.mapName),
			TurretCount:        base.TurretCount,
		})
	}
	return out, nil
}

func sortedUUIDKeys[T any](values map[uuid.UUID]T) []uuid.UUID {
	keys := make([]uuid.UUID, 0, len(values))
	for id := range values {
		keys = append(keys, id)
	}
	sort.Slice(keys, func(i int, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	return keys
}

func sortedBaseStructureUUIDs(base arkobject.Base) []string {
	ids := sortedUUIDKeys(base.Structures)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func sortedUUIDStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	sort.Strings(out)
	return out
}

func locationInfo(value *arkobject.ActorTransform) *LocationInfo {
	if value == nil {
		return nil
	}
	return &LocationInfo{
		X:          value.X,
		Y:          value.Y,
		Z:          value.Z,
		Pitch:      value.Pitch,
		Roll:       value.Roll,
		Yaw:        value.Yaw,
		Quaternion: value.Quaternion,
	}
}

func mapCoordsInfo(value *arkobject.ActorTransform, mapName string) *MapCoordsInfo {
	if value == nil || mapName == "" {
		return nil
	}
	coords := value.AsMapCoords(mapName)
	return &MapCoordsInfo{Lat: coords.Lat, Lon: coords.Long}
}

func ownerInfo(value arkobject.ObjectOwner) OwnerInfo {
	return OwnerInfo{
		PlayerID:         value.PlayerID,
		PlayerName:       value.PlayerName,
		TribeID:          value.TribeID,
		TribeName:        value.TribeName,
		OriginalPlacerID: value.OriginalPlacerID,
	}
}

func dinoOwnerInfo(value arkobject.DinoOwner) DinoOwnerInfo {
	return DinoOwnerInfo{
		PlayerID:          value.PlayerID,
		PlayerName:        value.PlayerName,
		TribeName:         value.TribeName,
		TamerTribeID:      value.TamerTribeID,
		TamerString:       value.TamerString,
		ImprinterName:     value.ImprinterName,
		ImprinterUniqueID: value.ImprinterUniqueID,
		TargetTeam:        value.TargetTeam,
	}
}

func dinoStatsInfo(value *arkobject.DinoStats) *DinoStatsInfo {
	if value == nil {
		return nil
	}
	return &DinoStatsInfo{
		BaseLevel:         value.BaseLevel,
		CurrentLevel:      value.CurrentLevel,
		BaseStatPoints:    dinoStatPointsInfo(value.BaseStatPoints),
		AddedStatPoints:   dinoStatPointsInfo(value.AddedStatPoints),
		MutatedStatPoints: dinoStatPointsInfo(value.MutatedStatPoints),
		StatValues:        dinoStatValuesInfo(value.StatValues),
		ImprintingPercent: value.ImprintingPercent,
	}
}

func geneTraitInfos(values []arkobject.GeneTrait) []GeneTraitInfo {
	if len(values) == 0 {
		return nil
	}
	out := make([]GeneTraitInfo, 0, len(values))
	for _, value := range values {
		out = append(out, GeneTraitInfo{
			Raw:   value.Raw,
			Name:  value.Name,
			Level: value.Level,
		})
	}
	return out
}

func dinoIDInfos(values []arkobject.DinoID) []DinoIDInfo {
	if len(values) == 0 {
		return nil
	}
	out := make([]DinoIDInfo, 0, len(values))
	for _, value := range values {
		out = append(out, DinoIDInfo{ID1: value.ID1, ID2: value.ID2})
	}
	return out
}

func dinoStatPointsInfo(value arkobject.DinoStatPoints) DinoStatPointsInfo {
	return DinoStatPointsInfo{
		Health:        value.Health,
		Stamina:       value.Stamina,
		Torpidity:     value.Torpidity,
		Oxygen:        value.Oxygen,
		Food:          value.Food,
		Water:         value.Water,
		Temperature:   value.Temperature,
		Weight:        value.Weight,
		MeleeDamage:   value.MeleeDamage,
		MovementSpeed: value.MovementSpeed,
		Fortitude:     value.Fortitude,
		CraftingSpeed: value.CraftingSpeed,
	}
}

func dinoStatValuesInfo(value arkobject.DinoStatValues) DinoStatValuesInfo {
	return DinoStatValuesInfo{
		Health:        value.Health,
		Stamina:       value.Stamina,
		Torpidity:     value.Torpidity,
		Oxygen:        value.Oxygen,
		Food:          value.Food,
		Water:         value.Water,
		Temperature:   value.Temperature,
		Weight:        value.Weight,
		MeleeDamage:   value.MeleeDamage,
		MovementSpeed: value.MovementSpeed,
		Fortitude:     value.Fortitude,
		CraftingSpeed: value.CraftingSpeed,
	}
}

func optionalUUIDString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func intArrayFromFixed6(values [6]int) []int {
	out := make([]int, len(values))
	copy(out, values[:])
	return out
}

func stringArrayFromFixed6(values [6]string) []string {
	out := make([]string, len(values))
	copy(out, values[:])
	return out
}

func crafterInfo(value *arkobject.ObjectCrafter) *CrafterInfo {
	if value == nil {
		return nil
	}
	return &CrafterInfo{CharacterName: value.CharacterName, TribeName: value.TribeName}
}

func equipmentStatsInfo(value arkobject.EquipmentStats) *EquipmentStatsInfo {
	if len(value.Internal) == 0 && value.Damage == 0 && value.Durability == 0 && value.Armor == 0 &&
		value.HypothermalResistance == 0 && value.HyperthermalResistance == 0 {
		return nil
	}
	internal := map[string]uint16{}
	for stat, raw := range value.Internal {
		internal[equipmentStatName(stat)] = raw
	}
	return &EquipmentStatsInfo{
		Internal:               internal,
		Damage:                 finiteFloat(value.Damage),
		Durability:             finiteFloat(value.Durability),
		Armor:                  finiteFloat(value.Armor),
		HypothermalResistance:  finiteFloat(value.HypothermalResistance),
		HyperthermalResistance: finiteFloat(value.HyperthermalResistance),
	}
}

func equipmentStatNames(stats []arkobject.EquipmentStat) []string {
	if len(stats) == 0 {
		return nil
	}
	out := make([]string, 0, len(stats))
	for _, stat := range stats {
		out = append(out, equipmentStatName(stat))
	}
	return out
}

func finiteFloat(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func (e *EquipmentInfo) sanitize() {
	e.Rating = finiteFloat(e.Rating)
	e.CurrentDurability = finiteFloat(e.CurrentDurability)
}

func equipmentStatName(stat arkobject.EquipmentStat) string {
	switch stat {
	case arkobject.EquipmentStatArmor:
		return "armor"
	case arkobject.EquipmentStatDurability:
		return "durability"
	case arkobject.EquipmentStatDamage:
		return "damage"
	case arkobject.EquipmentStatHypothermalResistance:
		return "hypothermal_resistance"
	case arkobject.EquipmentStatHyperthermalResistance:
		return "hyperthermal_resistance"
	default:
		return "unknown"
	}
}
