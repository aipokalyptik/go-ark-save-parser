package arkapi

import (
	"math"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type BaseAPI struct {
	structures *StructureAPI
	mapName    string
}

func NewBase(save *arksave.Save, mapName string) *BaseAPI {
	if mapName == "" && save.Context != nil {
		mapName = save.Context.MapName
	}
	return &BaseAPI{structures: NewStructure(save), mapName: mapName}
}

func (b *BaseAPI) At(coords arkobject.MapCoords, radius float64, owner *arkobject.ObjectOwner) (*arkobject.Base, error) {
	all, err := b.structures.All()
	if err != nil {
		return nil, err
	}
	nearby := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range all {
		if structure.Location == nil {
			continue
		}
		if owner != nil && !structure.IsOwnedBy(*owner) {
			continue
		}
		if structure.Location.IsAtMapCoordinate(b.mapName, coords, radius) {
			nearby[id] = structure
		}
	}
	if len(nearby) == 0 {
		return nil, nil
	}
	keystone := closestStructure(nearby, b.mapName, coords)
	base := arkobject.BaseFromStructures(keystone, nearby)
	return &base, nil
}

func closestStructure(structures map[uuid.UUID]arkobject.Structure, mapName string, coords arkobject.MapCoords) uuid.UUID {
	var closest uuid.UUID
	closestDistance := math.Inf(1)
	for id, structure := range structures {
		if structure.Location == nil {
			continue
		}
		distance := structure.Location.AsMapCoords(mapName).DistanceTo(coords)
		if distance < closestDistance {
			closest = id
			closestDistance = distance
		}
	}
	return closest
}
