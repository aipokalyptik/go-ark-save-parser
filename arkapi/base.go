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
	nearby, err := b.structures.AtLocation(b.mapName, coords, radius, nil)
	if err != nil {
		return nil, err
	}
	if len(nearby) == 0 {
		return nil, nil
	}
	structures, err := b.structures.ConnectedStructures(nearby)
	if err != nil {
		return nil, err
	}
	if owner != nil {
		structures, err = b.structures.FilterByOwner(structures, owner, 0, false)
		if err != nil {
			return nil, err
		}
	}
	if len(structures) == 0 {
		return nil, nil
	}
	keystone := closestStructure(structures, b.mapName, coords)
	keystoneOwner := structures[keystone].Owner
	filtered := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range structures {
		if structure.Owner == keystoneOwner {
			filtered[id] = structure
		}
	}
	base := arkobject.BaseFromStructures(keystone, filtered)
	return &base, nil
}

func (b *BaseAPI) All() ([]arkobject.Base, error) {
	structures, err := b.structures.All()
	if err != nil {
		return nil, err
	}
	return basesFromStructures(structures), nil
}

func (b *BaseAPI) AllWithFaults() ([]arkobject.Base, []arksave.FaultyObjectInfo, error) {
	structures, faults, err := b.structures.AllWithFaults()
	if err != nil {
		return nil, nil, err
	}
	return basesFromStructures(structures), faults, nil
}

func basesFromStructures(structures map[uuid.UUID]arkobject.Structure) []arkobject.Base {
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
	return bases
}

func (b *BaseAPI) AllWithMinStructures(minStructures int) ([]arkobject.Base, error) {
	all, err := b.All()
	if err != nil {
		return nil, err
	}
	if minStructures <= 0 {
		return all, nil
	}
	out := make([]arkobject.Base, 0, len(all))
	for _, base := range all {
		if base.StructureCount >= minStructures {
			out = append(out, base)
		}
	}
	return out, nil
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
