package arkapi

import (
	"encoding/json"
	"fmt"
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
	UUID         string        `json:"uuid"`
	Blueprint    string        `json:"blueprint"`
	ID1          uint32        `json:"id1"`
	ID2          uint32        `json:"id2"`
	IsFemale     bool          `json:"is_female"`
	IsTamed      bool          `json:"is_tamed"`
	IsBaby       bool          `json:"is_baby"`
	IsDead       bool          `json:"is_dead"`
	IsCryopodded bool          `json:"is_cryopodded"`
	GeneTraits   []string      `json:"gene_traits,omitempty"`
	Location     *LocationInfo `json:"location,omitempty"`
}

type StructureInfo struct {
	UUID          string        `json:"uuid"`
	Blueprint     string        `json:"blueprint"`
	ID            int32         `json:"id"`
	Owner         OwnerInfo     `json:"owner"`
	MaxHealth     float64       `json:"max_health"`
	CurrentHealth float64       `json:"current_health"`
	Location      *LocationInfo `json:"location,omitempty"`
}

type EquipmentInfo struct {
	UUID              string       `json:"uuid"`
	Blueprint         string       `json:"blueprint"`
	Kind              string       `json:"kind"`
	Quantity          int32        `json:"quantity"`
	IsEquipped        bool         `json:"is_equipped"`
	IsBlueprint       bool         `json:"is_blueprint"`
	Rating            float64      `json:"rating"`
	Quality           int32        `json:"quality"`
	CurrentDurability float64      `json:"current_durability"`
	Crafter           *CrafterInfo `json:"crafter,omitempty"`
}

type StackableInfo struct {
	UUID      string `json:"uuid"`
	Blueprint string `json:"blueprint"`
	Quantity  int32  `json:"quantity"`
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
	dinos, err := NewDino(j.save).All()
	if err != nil {
		return nil, err
	}
	out := make([]DinoInfo, 0, len(dinos))
	for _, id := range sortedUUIDKeys(dinos) {
		dino := dinos[id]
		out = append(out, DinoInfo{
			UUID:         id.String(),
			Blueprint:    dino.Blueprint,
			ID1:          dino.ID1,
			ID2:          dino.ID2,
			IsFemale:     dino.IsFemale,
			IsTamed:      dino.IsTamed,
			IsBaby:       dino.IsBaby,
			IsDead:       dino.IsDead,
			IsCryopodded: dino.IsCryopodded,
			GeneTraits:   dino.GeneTraits,
			Location:     locationInfo(dino.Location),
		})
	}
	return out, nil
}

func (j *JSONAPI) ExportStructures() ([]StructureInfo, error) {
	structures, err := NewStructure(j.save).All()
	if err != nil {
		return nil, err
	}
	out := make([]StructureInfo, 0, len(structures))
	for _, id := range sortedUUIDKeys(structures) {
		structure := structures[id]
		out = append(out, StructureInfo{
			UUID:          id.String(),
			Blueprint:     structure.Blueprint,
			ID:            structure.ID,
			Owner:         ownerInfo(structure.Owner),
			MaxHealth:     structure.MaxHealth,
			CurrentHealth: structure.CurrentHealth,
			Location:      locationInfo(structure.Location),
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
		item := equipment[id]
		out = append(out, EquipmentInfo{
			UUID:              id.String(),
			Blueprint:         item.Blueprint,
			Kind:              string(item.Kind),
			Quantity:          item.Quantity,
			IsEquipped:        item.IsEquipped,
			IsBlueprint:       item.IsBlueprint,
			Rating:            item.Rating,
			Quality:           item.Quality,
			CurrentDurability: item.CurrentDurability,
			Crafter:           crafterInfo(item.Crafter),
		})
	}
	return out, nil
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
			UUID:      id.String(),
			Blueprint: item.Blueprint,
			Quantity:  item.Quantity,
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

func ownerInfo(value arkobject.ObjectOwner) OwnerInfo {
	return OwnerInfo{
		PlayerID:         value.PlayerID,
		PlayerName:       value.PlayerName,
		TribeID:          value.TribeID,
		TribeName:        value.TribeName,
		OriginalPlacerID: value.OriginalPlacerID,
	}
}

func crafterInfo(value *arkobject.ObjectCrafter) *CrafterInfo {
	if value == nil {
		return nil
	}
	return &CrafterInfo{CharacterName: value.CharacterName, TribeName: value.TribeName}
}
