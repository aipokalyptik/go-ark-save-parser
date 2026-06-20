package arkapi

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type StructureAPI struct {
	save *arksave.Save
}

func NewStructure(save *arksave.Save) *StructureAPI {
	return &StructureAPI{save: save}
}

func (s *StructureAPI) All() (map[uuid.UUID]arkobject.Structure, error) {
	objects, err := s.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for _, info := range objects {
		if _, ok := info.Object.Value("StructureID"); !ok {
			continue
		}
		var location *arkobject.ActorTransform
		if transform, ok := s.save.ActorTransform(info.UUID); ok {
			location = &transform
		}
		out[info.UUID] = arkobject.StructureFromObject(info.Object, location)
	}
	return out, nil
}

func (s *StructureAPI) AllWithFaults() (map[uuid.UUID]arkobject.Structure, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := s.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for _, info := range objects {
		if _, ok := info.Object.Value("StructureID"); !ok {
			continue
		}
		var location *arkobject.ActorTransform
		if transform, ok := s.save.ActorTransform(info.UUID); ok {
			location = &transform
		}
		out[info.UUID] = arkobject.StructureFromObject(info.Object, location)
	}
	return out, faults, nil
}

func (s *StructureAPI) OwnedBy(owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.Structure, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range all {
		if structure.IsOwnedBy(owner) {
			out[id] = structure
		}
	}
	return out, nil
}

func (s *StructureAPI) FilterByOwner(structures map[uuid.UUID]arkobject.Structure, owner *arkobject.ObjectOwner, tribeID int32, invert bool) (map[uuid.UUID]arkobject.Structure, error) {
	if owner == nil && tribeID == 0 {
		return nil, errors.New("either owner or tribeID must be provided")
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range structures {
		if owner != nil && structure.IsOwnedBy(*owner) && !invert {
			out[id] = structure
		} else if tribeID != 0 && structure.Owner.TribeID == tribeID && !invert {
			out[id] = structure
		} else if invert {
			out[id] = structure
		}
	}
	return out, nil
}

func (s *StructureAPI) ByClass(blueprints []string) (map[uuid.UUID]arkobject.Structure, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range all {
		if _, ok := allowed[structure.Blueprint]; ok {
			out[id] = structure
		}
	}
	return out, nil
}

func (s *StructureAPI) ByID(id uuid.UUID) (arkobject.Structure, bool, error) {
	obj, err := s.save.Object(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return arkobject.Structure{}, false, nil
		}
		return arkobject.Structure{}, false, err
	}
	if obj == nil || !isStructureBlueprint(obj.Blueprint) {
		return arkobject.Structure{}, false, nil
	}
	if _, ok := obj.Value("StructureID"); !ok {
		return arkobject.Structure{}, false, nil
	}
	var location *arkobject.ActorTransform
	if transform, ok := s.save.ActorTransform(id); ok {
		location = &transform
	}
	return arkobject.StructureFromObject(obj, location), true, nil
}

func (s *StructureAPI) ConnectedStructures(structures map[uuid.UUID]arkobject.Structure) (map[uuid.UUID]arkobject.Structure, error) {
	out := make(map[uuid.UUID]arkobject.Structure, len(structures))
	queue := make([]uuid.UUID, 0, len(structures))
	queued := map[uuid.UUID]bool{}
	for _, id := range sortedUUIDKeys(structures) {
		out[id] = structures[id]
		queue = append(queue, id)
		queued[id] = true
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		structure, ok := out[id]
		if !ok {
			continue
		}
		for _, linkedID := range structure.LinkedStructureUUIDs {
			if _, exists := out[linkedID]; exists {
				continue
			}
			if queued[linkedID] {
				continue
			}
			linked, ok, err := s.ByID(linkedID)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			out[linkedID] = linked
			queue = append(queue, linkedID)
			queued[linkedID] = true
		}
	}
	return out, nil
}

func (s *StructureAPI) AtLocation(mapName string, coords arkobject.MapCoords, radius float64, blueprints []string) (map[uuid.UUID]arkobject.Structure, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	if mapName == "" && s.save.Context != nil {
		mapName = s.save.Context.MapName
	}
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range all {
		if structure.Location == nil {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[structure.Blueprint]; !ok {
				continue
			}
		}
		if structure.Location.IsAtMapCoordinate(mapName, coords, radius) {
			out[id] = structure
		}
	}
	return out, nil
}

func (s *StructureAPI) FilterByLocation(mapName string, coords arkobject.MapCoords, radius float64, structures map[uuid.UUID]arkobject.Structure) map[uuid.UUID]arkobject.Structure {
	if mapName == "" && s.save.Context != nil {
		mapName = s.save.Context.MapName
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range structures {
		if structure.Location == nil {
			continue
		}
		if structure.Location.IsAtMapCoordinate(mapName, coords, radius) {
			out[id] = structure
		}
	}
	return out
}

func (s *StructureAPI) AllWithInventory() (map[uuid.UUID]arkobject.Structure, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range all {
		if structure.InventoryUUID != nil {
			out[id] = structure
		}
	}
	return out, nil
}

func (s *StructureAPI) ContainerOfInventory(inventoryID uuid.UUID) (uuid.UUID, arkobject.Structure, bool, error) {
	structures, err := s.AllWithInventory()
	if err != nil {
		return uuid.Nil, arkobject.Structure{}, false, err
	}
	for id, structure := range structures {
		if structure.InventoryUUID != nil && *structure.InventoryUUID == inventoryID {
			return id, structure, true, nil
		}
	}
	return uuid.Nil, arkobject.Structure{}, false, nil
}

func isStructureBlueprint(name string) bool {
	if name == "" {
		return false
	}
	if !strings.Contains(name, "Structures") {
		return false
	}
	if strings.Contains(name, "/Skins/") ||
		strings.Contains(name, "PrimalInventory") ||
		strings.Contains(name, "/TreasureMap/") ||
		strings.Contains(name, "PrimalItemStructureSkin") ||
		strings.Contains(name, "PrimalItemResource") ||
		strings.Contains(name, "/TrainCarts/") {
		return false
	}
	return !strings.Contains(name, "PrimalItemStructure_") || strings.Contains(name, "PrimalItemStructure_ASR")
}
