package arkapi

import (
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type StackableAPI struct {
	save *arksave.Save
}

func NewStackable(save *arksave.Save) *StackableAPI {
	return &StackableAPI{save: save}
}

func (s *StackableAPI) IsApplicableBlueprint(blueprint string) bool {
	return strings.Contains(blueprint, "Resources/PrimalItemResource") ||
		strings.Contains(blueprint, "/PrimalItemConsumable") ||
		strings.Contains(blueprint, "PrimalItemAmmo")
}

func (s *StackableAPI) Count(items map[uuid.UUID]arkobject.InventoryItem) int32 {
	var count int32
	for _, item := range items {
		count += item.Quantity
	}
	return count
}

func (s *StackableAPI) All() (map[uuid.UUID]arkobject.InventoryItem, error) {
	objects, err := s.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return s.IsApplicableBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for _, info := range objects {
		if boolProperty(info.Object, "bIsBlueprint") || boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.InventoryItemFromObject(info.Object)
	}
	return out, nil
}

func (s *StackableAPI) AllWithFaults() (map[uuid.UUID]arkobject.InventoryItem, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := s.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return s.IsApplicableBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for _, info := range objects {
		if boolProperty(info.Object, "bIsBlueprint") || boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.InventoryItemFromObject(info.Object)
	}
	return out, faults, nil
}

func (s *StackableAPI) ByClass(blueprints []string) (map[uuid.UUID]arkobject.InventoryItem, error) {
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	objects, err := s.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		_, ok := allowed[info.ClassName]
		return ok
	})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for _, info := range objects {
		if boolProperty(info.Object, "bIsBlueprint") || boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.InventoryItemFromObject(info.Object)
	}
	return out, nil
}

func (s *StackableAPI) Resources() (map[uuid.UUID]arkobject.InventoryItem, error) {
	return s.byBlueprintSubstring("Resources/PrimalItemResource")
}

func (s *StackableAPI) Ammo() (map[uuid.UUID]arkobject.InventoryItem, error) {
	return s.byBlueprintSubstring("PrimalItemAmmo")
}

func (s *StackableAPI) Consumables() (map[uuid.UUID]arkobject.InventoryItem, error) {
	return s.byBlueprintSubstring("/PrimalItemConsumable")
}

func (s *StackableAPI) OwnedBy(owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.InventoryItem, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	return s.FilterOwnedBy(all, owner)
}

func (s *StackableAPI) FilterOwnedBy(items map[uuid.UUID]arkobject.InventoryItem, owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.InventoryItem, error) {
	containers, err := selectedInventoryContainerOwners(s.save)
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for id, item := range items {
		if item.OwnerInventory == nil {
			continue
		}
		containerOwner, ok := containers[*item.OwnerInventory]
		if ok && ownerMatches(containerOwner, owner) {
			out[id] = item
		}
	}
	return out, nil
}

func selectedInventoryContainerOwners(save *arksave.Save) (map[uuid.UUID]arkobject.ObjectOwner, error) {
	infos, _, err := save.SelectedObjectPropertiesWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName) || isPotentialStructureContainer(info.ClassName)
	}, []string{"StructureID", "MyInventoryComponent", "OriginalPlacerPlayerID", "OwnerName", "OwningPlayerName", "OwningPlayerID", "TargetingTeam", "bIsEngram"})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.ObjectOwner{}
	for _, info := range infos {
		container := arkproperty.Container{Properties: info.Properties}
		if _, ok := container.Value("StructureID"); !ok {
			continue
		}
		if _, ok := container.Value("bIsEngram"); ok {
			continue
		}
		if !isStructureBlueprint(info.ClassName) {
			if !isPotentialStructureContainer(info.ClassName) {
				continue
			}
			if _, ok := container.Value("MyInventoryComponent"); !ok {
				continue
			}
		}
		inventoryID, ok := selectedObjectReferenceUUID(container, "MyInventoryComponent")
		if !ok {
			continue
		}
		out[inventoryID] = arkobject.ObjectOwnerFromContainer(container)
	}
	return out, nil
}

func selectedObjectReferenceUUID(container arkproperty.Container, name string) (uuid.UUID, bool) {
	raw, ok := container.Value(name)
	if !ok {
		return uuid.Nil, false
	}
	ref, ok := raw.(arkproperty.ObjectReference)
	if !ok || ref.Type != arkproperty.ObjectReferenceUUID {
		return uuid.Nil, false
	}
	rawID, ok := ref.Value.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(rawID)
	return id, err == nil
}

func ownerMatches(candidate arkobject.ObjectOwner, target arkobject.ObjectOwner) bool {
	if candidate.PlayerID != 0 && candidate.PlayerID == target.PlayerID {
		return true
	}
	if candidate.PlayerName != "" && candidate.PlayerName == target.PlayerName {
		return true
	}
	if candidate.TribeName != "" && candidate.TribeName == target.TribeName {
		return true
	}
	if candidate.TribeID != 0 && candidate.TribeID == target.TribeID {
		return true
	}
	if candidate.OriginalPlacerID != 0 && candidate.OriginalPlacerID == target.OriginalPlacerID {
		return true
	}
	return false
}

func (s *StackableAPI) byBlueprintSubstring(substring string) (map[uuid.UUID]arkobject.InventoryItem, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for id, item := range all {
		if strings.Contains(item.Blueprint, substring) {
			out[id] = item
		}
	}
	return out, nil
}

func boolProperty(object *arkobject.GameObject, name string) bool {
	value, ok := object.Value(name)
	if !ok {
		return false
	}
	out, _ := value.(bool)
	return out
}
