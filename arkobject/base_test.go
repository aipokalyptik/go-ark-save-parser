package arkobject

import (
	"testing"

	"github.com/google/uuid"
)

func TestBaseFromStructuresUsesKeystoneOwnerAndLocation(t *testing.T) {
	keystoneID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	otherID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	structures := map[uuid.UUID]Structure{
		keystoneID: {
			UUID:      keystoneID,
			Blueprint: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			Owner:     ObjectOwner{TribeID: 555, TribeName: "Builders"},
			Location:  &ActorTransform{X: 10, Y: 20, Z: 30},
		},
		otherID: {
			UUID:      otherID,
			Blueprint: "Blueprint'/Game/Structures/Stone/PrimalStructure_Foundation_Stone.PrimalStructure_Foundation_Stone_C'",
			Owner:     ObjectOwner{TribeID: 555, TribeName: "Builders"},
			Location:  &ActorTransform{X: 30, Y: 40, Z: 50},
		},
	}

	base := BaseFromStructures(keystoneID, structures)

	if base.KeystoneUUID != keystoneID || base.Owner.TribeID != 555 || base.StructureCount != 2 {
		t.Fatalf("Base = %#v", base)
	}
	if base.Location == nil || base.Location.X != 10 || base.Location.Y != 20 {
		t.Fatalf("Base.Location = %#v, want keystone location", base.Location)
	}
	if base.AverageLocation == nil || base.AverageLocation.X != 20 || base.AverageLocation.Y != 30 || base.AverageLocation.Z != 40 {
		t.Fatalf("Base.AverageLocation = %#v", base.AverageLocation)
	}
}

func TestBaseFromStructuresCountsTurrets(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	base := BaseFromStructures(id, map[uuid.UUID]Structure{
		id: {
			UUID:      id,
			Blueprint: "Blueprint'/Game/PrimalEarth/Structures/Turrets/PrimalStructureTurretHeavy.PrimalStructureTurretHeavy_C'",
			Location:  &ActorTransform{},
		},
	})

	if base.TurretCount != 1 {
		t.Fatalf("TurretCount = %f, want 1", base.TurretCount)
	}
}
