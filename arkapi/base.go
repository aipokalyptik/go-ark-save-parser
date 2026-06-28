package arkapi

import (
	"math"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type BaseAPI struct {
	structures *StructureAPI
	mapName    string
}

type BaseQueryOptions struct {
	OnlyConnected bool
	Radius        float64
	MinStructures int
}

type BaseComponentStats struct {
	Components          int
	TotalStructures     int
	LargestComponent    int
	ComponentsAtLeast10 int
	Faults              int
}

func NewBase(save *arksave.Save, mapName string) *BaseAPI {
	if mapName == "" && save.Context != nil {
		mapName = save.Context.MapName
	}
	return &BaseAPI{structures: NewStructure(save), mapName: mapName}
}

func NewBaseFromPath(savePath string, mapName string) (*BaseAPI, func() error, error) {
	save, err := arksave.Open(savePath)
	if err != nil {
		return nil, nil, err
	}
	return NewBase(save, mapName), save.Close, nil
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

func (b *BaseAPI) ComponentStats() (BaseComponentStats, error) {
	infos, faults, err := b.structures.save.SelectedObjectPropertiesWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName) || isPotentialStructureContainer(info.ClassName)
	}, []string{"StructureID", "LinkedStructures", "MyInventoryComponent", "bIsEngram"})
	if err != nil {
		return BaseComponentStats{}, err
	}
	links := make(map[uuid.UUID][]uuid.UUID, len(infos))
	for _, info := range infos {
		container := arkproperty.Container{Properties: info.Properties}
		if _, ok := container.Value("StructureID"); !ok {
			continue
		}
		if isSkippedBaseComponentBlueprint(info.ClassName) {
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
		links[info.UUID] = selectedLinkedStructureUUIDs(container)
	}
	stats := componentStatsFromLinks(links)
	stats.Faults = len(faults)
	return stats, nil
}

func isSkippedBaseComponentBlueprint(className string) bool {
	return strings.Contains(className, "/LostColony/Structures/TekBunker/Structures/BP_Bunker_Base.")
}

func selectedLinkedStructureUUIDs(container arkproperty.Container) []uuid.UUID {
	raw, ok := container.Value("LinkedStructures")
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]uuid.UUID, 0, len(array.Values))
	for _, value := range array.Values {
		ref, ok := value.(arkproperty.ObjectReference)
		if !ok || ref.Type != arkproperty.ObjectReferenceUUID {
			continue
		}
		rawID, ok := ref.Value.(string)
		if !ok {
			continue
		}
		id, err := uuid.Parse(rawID)
		if err == nil {
			out = append(out, id)
		}
	}
	return out
}

func componentStatsFromLinks(links map[uuid.UUID][]uuid.UUID) BaseComponentStats {
	adjacent := make(map[uuid.UUID][]uuid.UUID, len(links))
	for id, linkedIDs := range links {
		if _, ok := adjacent[id]; !ok {
			adjacent[id] = nil
		}
		for _, linkedID := range linkedIDs {
			if _, ok := links[linkedID]; !ok {
				continue
			}
			adjacent[id] = append(adjacent[id], linkedID)
			adjacent[linkedID] = append(adjacent[linkedID], id)
		}
	}
	visited := make(map[uuid.UUID]bool, len(links))
	stats := BaseComponentStats{TotalStructures: len(links)}
	for _, start := range sortedUUIDKeys(links) {
		if visited[start] {
			continue
		}
		stats.Components++
		size := 0
		stack := []uuid.UUID{start}
		visited[start] = true
		for len(stack) > 0 {
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			size++
			for _, next := range adjacent[id] {
				if visited[next] {
					continue
				}
				visited[next] = true
				stack = append(stack, next)
			}
		}
		if size > stats.LargestComponent {
			stats.LargestComponent = size
		}
		if size >= 10 {
			stats.ComponentsAtLeast10++
		}
	}
	return stats
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

func (b *BaseAPI) AllBases(opts BaseQueryOptions) ([]arkobject.Base, error) {
	radius := opts.Radius
	if radius == 0 {
		radius = 0.3
	}
	minStructures := opts.MinStructures
	if minStructures == 0 {
		minStructures = 10
	}
	if opts.OnlyConnected {
		return b.AllWithMinStructures(minStructures)
	}
	all, err := b.structures.All()
	if err != nil {
		return nil, err
	}
	visited := map[uuid.UUID]bool{}
	bases := make([]arkobject.Base, 0)
	for _, id := range sortedUUIDKeys(all) {
		if visited[id] {
			continue
		}
		structure := all[id]
		if structure.Location == nil {
			continue
		}
		base, err := b.At(structure.Location.AsMapCoords(b.mapName), radius, &structure.Owner)
		if err != nil {
			return nil, err
		}
		if base == nil {
			continue
		}
		for structureID := range base.Structures {
			visited[structureID] = true
		}
		if base.StructureCount >= minStructures {
			bases = append(bases, *base)
		}
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
