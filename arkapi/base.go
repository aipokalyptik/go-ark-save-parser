package arkapi

import (
	"math"
	"sort"

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

func (b *BaseAPI) All() ([]arkobject.Base, error) {
	structures, err := b.structures.All()
	if err != nil {
		return nil, err
	}
	adjacent := make(map[uuid.UUID][]uuid.UUID, len(structures))
	for id, structure := range structures {
		for _, linkedID := range structure.LinkedStructureUUIDs {
			if _, ok := structures[linkedID]; !ok {
				continue
			}
			adjacent[id] = append(adjacent[id], linkedID)
			adjacent[linkedID] = append(adjacent[linkedID], id)
		}
	}
	visited := make(map[uuid.UUID]bool, len(structures))
	bases := make([]arkobject.Base, 0)
	for _, start := range sortedUUIDKeys(structures) {
		if visited[start] {
			continue
		}
		component := make(map[uuid.UUID]arkobject.Structure)
		stack := []uuid.UUID{start}
		visited[start] = true
		for len(stack) > 0 {
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			component[id] = structures[id]
			for _, next := range adjacent[id] {
				if visited[next] {
					continue
				}
				visited[next] = true
				stack = append(stack, next)
			}
		}
		keystone := baseKeystone(component)
		bases = append(bases, arkobject.BaseFromStructures(keystone, component))
	}
	sort.Slice(bases, func(i int, j int) bool {
		return bases[i].KeystoneUUID.String() < bases[j].KeystoneUUID.String()
	})
	return bases, nil
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

func baseKeystone(structures map[uuid.UUID]arkobject.Structure) uuid.UUID {
	var keystone uuid.UUID
	for _, id := range sortedUUIDKeys(structures) {
		if keystone == uuid.Nil || id.String() < keystone.String() {
			keystone = id
		}
	}
	return keystone
}
