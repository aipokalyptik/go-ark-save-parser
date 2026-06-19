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
	ids, err := s.save.ObjectIDs()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for _, id := range ids {
		object, err := s.save.Object(id)
		if err != nil {
			return nil, err
		}
		if !isStructureBlueprint(object.Blueprint) {
			continue
		}
		if _, ok := object.Value("StructureID"); !ok {
			continue
		}
		var location *arkobject.ActorTransform
		if transform, ok := s.save.ActorTransform(id); ok {
			location = &transform
		}
		out[id] = arkobject.StructureFromObject(object, location)
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
