package arkobject

import (
	"strings"

	"github.com/google/uuid"
)

type Base struct {
	KeystoneUUID    uuid.UUID
	Structures      map[uuid.UUID]Structure
	StructureCount  int
	Owner           ObjectOwner
	Location        *ActorTransform
	AverageLocation *ActorTransform
	TurretCount     float64
}

func BaseFromStructures(keystone uuid.UUID, structures map[uuid.UUID]Structure) Base {
	base := Base{
		KeystoneUUID:   keystone,
		Structures:     structures,
		StructureCount: len(structures),
		TurretCount:    countTurrets(structures),
	}
	if structure, ok := structures[keystone]; ok {
		base.Owner = structure.Owner
		base.Location = structure.Location
	}
	base.AverageLocation = averageStructureLocation(structures)
	return base
}

func averageStructureLocation(structures map[uuid.UUID]Structure) *ActorTransform {
	var totalX, totalY, totalZ float64
	var count float64
	for _, structure := range structures {
		if structure.Location == nil {
			continue
		}
		totalX += structure.Location.X
		totalY += structure.Location.Y
		totalZ += structure.Location.Z
		count++
	}
	if count == 0 {
		return nil
	}
	return &ActorTransform{X: totalX / count, Y: totalY / count, Z: totalZ / count}
}

func countTurrets(structures map[uuid.UUID]Structure) float64 {
	var count float64
	for _, structure := range structures {
		blueprint := strings.ToLower(structure.Blueprint)
		switch {
		case strings.Contains(blueprint, "turretheavy") || strings.Contains(blueprint, "turrettek"):
			count++
		case strings.Contains(blueprint, "turret") && strings.Contains(blueprint, "auto"):
			count += 0.25
		}
	}
	return count
}
