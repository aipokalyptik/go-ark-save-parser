package arkapi

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type StructureAPI struct {
	save *arksave.Save
}

type StructureOwnerSummary struct {
	Structures            int
	WithTribeID           int
	WithPlayerID          int
	WithTribeName         int
	WithPlayerName        int
	WithOriginalPlacerID  int
	UniqueTribes          int
	UniquePlayers         int
	UniqueOriginalPlacers int
}

type StructureTribeOwnershipSummary struct {
	TribeID    int32
	Structures int
}

type StructureLocationSummary struct {
	Structures int
	Connected  int
}

type StructureHealthSummary struct {
	Structures           int
	WithHealth           int
	Damaged              int
	FullyRepaired        int
	WithoutMaxHealth     int
	TotalMaxHealth       float64
	TotalCurrentHealth   float64
	AverageHealthPercent float64
	MinimumHealthPercent float64
	MaximumHealthPercent float64
}

type StructureOwnerLocationExport struct {
	Structures          int                          `json:"structures"`
	Owners              int                          `json:"owners"`
	Cells               int                          `json:"cells"`
	NamedCells          int                          `json:"named_cells"`
	MultiStructureCells int                          `json:"multi_structure_cells"`
	FaultCount          int                          `json:"fault_count"`
	OwnersByLocation    []StructureOwnerLocationData `json:"owners_by_location"`
}

type StructureOwnerLocationData struct {
	Owner string                       `json:"owner"`
	Cells []StructureOwnerLocationCell `json:"cells"`
}

type StructureOwnerLocationCell struct {
	Coords string `json:"coords"`
	Name   string `json:"name,omitempty"`
	Count  int    `json:"count,omitempty"`
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
	containers, err := s.missedInventoryContainers()
	if err != nil {
		return nil, err
	}
	objects = append(objects, containers...)
	out := map[uuid.UUID]arkobject.Structure{}
	for _, info := range objects {
		if _, exists := out[info.UUID]; exists {
			continue
		}
		if !isParsedStructure(info.Object) {
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
	containers, containerFaults, err := s.missedInventoryContainersWithFaults()
	if err != nil {
		return nil, nil, err
	}
	objects = append(objects, containers...)
	faults = append(faults, containerFaults...)
	out := map[uuid.UUID]arkobject.Structure{}
	for _, info := range objects {
		if _, exists := out[info.UUID]; exists {
			continue
		}
		if !isParsedStructure(info.Object) {
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

func (s *StructureAPI) missedInventoryContainers() ([]arksave.ParsedObjectInfo, error) {
	infos, err := s.save.ObjectClassInfosWithAnyProperty([]string{"MyInventoryComponent"})
	if err != nil {
		return nil, err
	}
	objects := make([]arksave.ParsedObjectInfo, 0, len(infos))
	for _, info := range infos {
		if isStructureBlueprint(info.ClassName) || !isPotentialStructureContainer(info.ClassName) {
			continue
		}
		object, err := s.save.ParsedObject(info.UUID)
		if err != nil {
			return nil, err
		}
		objects = append(objects, arksave.ParsedObjectInfo{UUID: info.UUID, ClassName: info.ClassName, Object: object})
	}
	return objects, nil
}

func (s *StructureAPI) missedInventoryContainersWithFaults() ([]arksave.ParsedObjectInfo, []arksave.FaultyObjectInfo, error) {
	infos, err := s.save.ObjectClassInfosWithAnyProperty([]string{"MyInventoryComponent"})
	if err != nil {
		return nil, nil, err
	}
	objects := make([]arksave.ParsedObjectInfo, 0, len(infos))
	var faults []arksave.FaultyObjectInfo
	for _, info := range infos {
		if isStructureBlueprint(info.ClassName) || !isPotentialStructureContainer(info.ClassName) {
			continue
		}
		object, err := s.save.ParsedObject(info.UUID)
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: info.UUID, ClassName: info.ClassName, Err: err})
			continue
		}
		objects = append(objects, arksave.ParsedObjectInfo{UUID: info.UUID, ClassName: info.ClassName, Object: object})
	}
	return objects, faults, nil
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

func (s *StructureAPI) OwnedByWithFaults(owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.Structure, []arksave.FaultyObjectInfo, error) {
	all, faults, err := s.AllWithFaults()
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for id, structure := range all {
		if structure.IsOwnedBy(owner) {
			out[id] = structure
		}
	}
	return out, faults, nil
}

func (s *StructureAPI) CountOwnedByTribe(tribeID int32) (int, error) {
	structures, err := s.OwnedBy(arkobject.ObjectOwner{TribeID: tribeID})
	if err != nil {
		return 0, err
	}
	return len(structures), nil
}

func (s *StructureAPI) CountOwnedByTribeWithFaults(tribeID int32) (int, []arksave.FaultyObjectInfo, error) {
	infos, faults, err := s.save.SelectedObjectPropertiesWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName) || isPotentialStructureContainer(info.ClassName)
	}, []string{"StructureID", "MyInventoryComponent", "TargetingTeam", "bIsEngram"})
	if err != nil {
		return 0, nil, err
	}
	count := 0
	for _, info := range infos {
		container := arkproperty.Container{Properties: info.Properties}
		if _, ok := container.Value("StructureID"); !ok {
			continue
		}
		if selectedBoolProperty(container, "bIsEngram") {
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
		if selectedInt32(container, "TargetingTeam") == tribeID {
			count++
		}
	}
	return count, faults, nil
}

func (s *StructureAPI) TribeOwnershipSummaryWithFaults(tribeID int32) (StructureTribeOwnershipSummary, []arksave.FaultyObjectInfo, error) {
	count, faults, err := s.CountOwnedByTribeWithFaults(tribeID)
	if err != nil {
		return StructureTribeOwnershipSummary{}, nil, err
	}
	return StructureTribeOwnershipSummary{TribeID: tribeID, Structures: count}, faults, nil
}

func (s *StructureAPI) OwnerSummaryWithFaults() (StructureOwnerSummary, []arksave.FaultyObjectInfo, error) {
	structures, faults, err := s.selectedStructureIndexWithFaults()
	if err != nil {
		return StructureOwnerSummary{}, nil, err
	}
	summary := StructureOwnerSummary{Structures: len(structures)}
	tribes := map[int32]struct{}{}
	players := map[int32]struct{}{}
	placers := map[int32]struct{}{}
	for _, structure := range structures {
		owner := structure.Owner
		if owner.TribeID != 0 {
			summary.WithTribeID++
			tribes[owner.TribeID] = struct{}{}
		}
		if owner.PlayerID != 0 {
			summary.WithPlayerID++
			players[owner.PlayerID] = struct{}{}
		}
		if owner.TribeName != "" {
			summary.WithTribeName++
		}
		if owner.PlayerName != "" {
			summary.WithPlayerName++
		}
		if owner.OriginalPlacerID != 0 {
			summary.WithOriginalPlacerID++
			placers[owner.OriginalPlacerID] = struct{}{}
		}
	}
	summary.UniqueTribes = len(tribes)
	summary.UniquePlayers = len(players)
	summary.UniqueOriginalPlacers = len(placers)
	return summary, faults, nil
}

func (s *StructureAPI) HealthSummaryWithFaults() (StructureHealthSummary, []arksave.FaultyObjectInfo, error) {
	infos, faults, err := s.save.SelectedObjectPropertiesWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName) || isPotentialStructureContainer(info.ClassName)
	}, []string{"StructureID", "MyInventoryComponent", "MaxHealth", "Health", "bIsEngram"})
	if err != nil {
		return StructureHealthSummary{}, nil, err
	}
	summary := StructureHealthSummary{Structures: len(infos)}
	for _, info := range infos {
		container := arkproperty.Container{Properties: info.Properties}
		if _, ok := container.Value("StructureID"); !ok {
			summary.Structures--
			continue
		}
		if selectedBoolProperty(container, "bIsEngram") {
			summary.Structures--
			continue
		}
		if !isStructureBlueprint(info.ClassName) {
			if !isPotentialStructureContainer(info.ClassName) {
				summary.Structures--
				continue
			}
			if _, ok := container.Value("MyInventoryComponent"); !ok {
				summary.Structures--
				continue
			}
		}
		maxHealth := selectedFloat64(container, "MaxHealth")
		currentHealth := selectedFloat64(container, "Health")
		if currentHealth == 0 {
			currentHealth = maxHealth
		}
		summary.addHealth(maxHealth, currentHealth)
	}
	return summary, faults, nil
}

func (s *StructureAPI) HealthSummaryForStructures(structures map[uuid.UUID]arkobject.Structure) StructureHealthSummary {
	summary := StructureHealthSummary{Structures: len(structures)}
	for _, structure := range structures {
		summary.addHealth(structure.MaxHealth, structure.CurrentHealth)
	}
	return summary
}

func (s *StructureHealthSummary) addHealth(maxHealth float64, currentHealth float64) {
	if maxHealth <= 0 {
		s.WithoutMaxHealth++
		return
	}
	s.WithHealth++
	s.TotalMaxHealth += maxHealth
	s.TotalCurrentHealth += currentHealth
	healthPercent := currentHealth / maxHealth * 100
	if s.WithHealth == 1 || healthPercent < s.MinimumHealthPercent {
		s.MinimumHealthPercent = healthPercent
	}
	if s.WithHealth == 1 || healthPercent > s.MaximumHealthPercent {
		s.MaximumHealthPercent = healthPercent
	}
	if currentHealth < maxHealth {
		s.Damaged++
	} else {
		s.FullyRepaired++
	}
	if s.TotalMaxHealth > 0 {
		s.AverageHealthPercent = s.TotalCurrentHealth / s.TotalMaxHealth * 100
	}
}

func (s *StructureAPI) OwnerLocationsWithFaults(mapName string, digits int, playerAPI *PlayerAPI) (StructureOwnerLocationExport, []arksave.FaultyObjectInfo, error) {
	structures, faults, err := s.selectedStructureIndexWithFaults()
	if err != nil {
		return StructureOwnerLocationExport{}, nil, err
	}
	if mapName == "" && s.save != nil && s.save.Context != nil {
		mapName = s.save.Context.MapName
	}
	if digits < 0 {
		digits = 0
	}
	scale := math.Pow10(digits)
	ownerNames, err := ownerLocationTribeNames(playerAPI)
	if err != nil {
		return StructureOwnerLocationExport{}, nil, err
	}
	byOwner := map[string]map[string]StructureOwnerLocationCell{}
	export := StructureOwnerLocationExport{Structures: len(structures), FaultCount: len(faults)}
	for _, id := range sortedUUIDKeys(structures) {
		structure := structures[id]
		if structure.Owner.TribeID == 0 || structure.Location == nil {
			continue
		}
		owner := structure.Owner.TribeName
		if owner == "" {
			owner = ownerNames[structure.Owner.TribeID]
		}
		if owner == "" {
			owner = fmt.Sprintf("%d", structure.Owner.TribeID)
		}
		coords := structure.Location.AsMapCoords(mapName)
		lat := math.Round(coords.Lat*scale) / scale
		lon := math.Round(coords.Long*scale) / scale
		coordKey := fmt.Sprintf("%.*f,%.*f", digits, lat, digits, lon)
		cells := byOwner[owner]
		if cells == nil {
			cells = map[string]StructureOwnerLocationCell{}
			byOwner[owner] = cells
		}
		cell, ok := cells[coordKey]
		if !ok {
			cell = StructureOwnerLocationCell{Coords: coordKey, Name: arkobject.ShortNameFromBlueprint(structure.Blueprint)}
		} else {
			cell.Name = ""
			if cell.Count == 0 {
				cell.Count = 1
			}
			cell.Count++
		}
		cells[coordKey] = cell
	}
	owners := make([]string, 0, len(byOwner))
	for owner := range byOwner {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	export.Owners = len(owners)
	for _, owner := range owners {
		cellMap := byOwner[owner]
		coordKeys := make([]string, 0, len(cellMap))
		for coord := range cellMap {
			coordKeys = append(coordKeys, coord)
		}
		sort.Strings(coordKeys)
		cells := make([]StructureOwnerLocationCell, 0, len(coordKeys))
		for _, coord := range coordKeys {
			cell := cellMap[coord]
			if cell.Name != "" {
				export.NamedCells++
			}
			if cell.Count > 1 {
				export.MultiStructureCells++
			}
			cells = append(cells, cell)
		}
		export.Cells += len(cells)
		export.OwnersByLocation = append(export.OwnersByLocation, StructureOwnerLocationData{Owner: owner, Cells: cells})
	}
	return export, faults, nil
}

func ownerLocationTribeNames(playerAPI *PlayerAPI) (map[int32]string, error) {
	names := map[int32]string{}
	if playerAPI == nil {
		return names, nil
	}
	tribes, faults, err := playerAPI.TribeDetailsWithFaults()
	if err != nil {
		return nil, err
	}
	if len(faults) > 0 {
		return nil, faults[0].Err
	}
	for _, tribe := range tribes {
		if tribe.TribeID != 0 && tribe.Name != "" {
			names[tribe.TribeID] = tribe.Name
		}
	}
	return names, nil
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

func (s *StructureAPI) ByClassOwnedBy(blueprints []string, owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.Structure, error) {
	structures, err := s.ByClass(blueprints)
	if err != nil {
		return nil, err
	}
	return s.FilterByOwner(structures, &owner, 0, false)
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
	index, _, err := s.selectedStructureIndexWithFaults()
	if err != nil {
		return nil, err
	}
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
			linked, ok := index[linkedID]
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
	all, _, err := s.selectedStructureIndexWithFaults()
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

func (s *StructureAPI) AtLocationWithFaults(mapName string, coords arkobject.MapCoords, radius float64, blueprints []string) (map[uuid.UUID]arkobject.Structure, []arksave.FaultyObjectInfo, error) {
	all, faults, err := s.AllWithFaults()
	if err != nil {
		return nil, nil, err
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
	return out, faults, nil
}

func (s *StructureAPI) AtLocationSummaryWithFaults(mapName string, coords arkobject.MapCoords, radius float64, blueprints []string) (StructureLocationSummary, []arksave.FaultyObjectInfo, error) {
	found, faults, err := s.AtLocationWithFaults(mapName, coords, radius, blueprints)
	if err != nil {
		return StructureLocationSummary{}, nil, err
	}
	connected, err := s.ConnectedStructures(found)
	if err != nil {
		return StructureLocationSummary{}, faults, err
	}
	return StructureLocationSummary{Structures: len(found), Connected: len(connected)}, faults, nil
}

func (s *StructureAPI) selectedStructureIndexWithFaults() (map[uuid.UUID]arkobject.Structure, []arksave.FaultyObjectInfo, error) {
	infos, faults, err := s.save.SelectedObjectPropertiesWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName) || isPotentialStructureContainer(info.ClassName)
	}, []string{
		"StructureID",
		"LinkedStructures",
		"MyInventoryComponent",
		"CurrentItemCount",
		"MaxItemCount",
		"OriginalPlacerPlayerID",
		"OwnerName",
		"OwningPlayerName",
		"OwningPlayerID",
		"TargetingTeam",
		"bIsEngram",
	})
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.Structure{}
	for _, info := range infos {
		container := arkproperty.Container{Properties: info.Properties}
		if _, ok := container.Value("StructureID"); !ok {
			continue
		}
		if selectedBoolProperty(container, "bIsEngram") {
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
		var location *arkobject.ActorTransform
		if transform, ok := s.save.ActorTransform(info.UUID); ok {
			location = &transform
		}
		structure := arkobject.Structure{
			UUID:                 info.UUID,
			Blueprint:            info.ClassName,
			Owner:                arkobject.ObjectOwnerFromContainer(container),
			ID:                   selectedInt32(container, "StructureID"),
			Location:             location,
			ItemCount:            selectedInt32(container, "CurrentItemCount"),
			MaxItemCount:         selectedInt32(container, "MaxItemCount"),
			LinkedStructureUUIDs: selectedLinkedStructureUUIDs(container),
		}
		if inventoryID, ok := selectedObjectReferenceUUID(container, "MyInventoryComponent"); ok {
			structure.InventoryUUID = &inventoryID
		}
		out[info.UUID] = structure
	}
	return out, faults, nil
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

func (s *StructureAPI) Heatmap(mapName string, resolution int, structures map[uuid.UUID]arkobject.Structure, blueprints []string, owner *arkobject.ObjectOwner, minInSection int) ([][]int, error) {
	if resolution <= 0 {
		return nil, fmt.Errorf("resolution must be positive")
	}
	if mapName == "" && s.save != nil && s.save.Context != nil {
		mapName = s.save.Context.MapName
	}
	var err error
	if structures == nil {
		if len(blueprints) > 0 {
			structures, err = s.ByClass(blueprints)
		} else {
			structures, err = s.All()
		}
		if err != nil {
			return nil, err
		}
	}
	allowed := blueprintSet(blueprints)
	heatmap := make([][]int, resolution)
	for i := range heatmap {
		heatmap[i] = make([]int, resolution)
	}
	for _, structure := range structures {
		if structure.Location == nil {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[structure.Blueprint]; !ok {
				continue
			}
		}
		if owner != nil && !structure.IsOwnedBy(*owner) {
			continue
		}
		coords := structure.Location.AsMapCoords(mapName)
		x := int(math.Floor(coords.Lat))
		y := int(math.Floor(coords.Long))
		if x < 0 || x >= resolution || y < 0 || y >= resolution {
			continue
		}
		heatmap[x][y]++
	}
	for i := range heatmap {
		for j := range heatmap[i] {
			if heatmap[i][j] < minInSection {
				heatmap[i][j] = 0
			}
		}
	}
	return heatmap, nil
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

func blueprintSet(blueprints []string) map[string]struct{} {
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	return allowed
}

func isParsedStructure(object *arkobject.GameObject) bool {
	if object == nil {
		return false
	}
	if boolProperty(object, "bIsEngram") {
		return false
	}
	_, ok := object.Value("StructureID")
	return ok
}

func isPotentialStructureContainer(name string) bool {
	if name == "" {
		return false
	}
	return !strings.Contains(name, "PlayerPawn") &&
		!strings.Contains(name, "/Dinos/") &&
		!strings.Contains(name, "Character_BP")
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

func selectedInt32(properties arkproperty.Container, name string) int32 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int32:
		return v
	case uint32:
		return int32(v)
	case int16:
		return int32(v)
	case uint16:
		return int32(v)
	case int8:
		return int32(v)
	case uint8:
		return int32(v)
	case int:
		return int32(v)
	case float32:
		return int32(v)
	case float64:
		return int32(v)
	default:
		return 0
	}
}

func selectedFloat64(properties arkproperty.Container, name string) float64 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int32:
		return float64(v)
	case uint32:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}
