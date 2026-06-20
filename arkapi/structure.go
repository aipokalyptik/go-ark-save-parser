package arkapi

import (
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
