package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

func TestObjectOwnerFromContainerReadsOwnershipFields(t *testing.T) {
	owner := ObjectOwnerFromContainer(arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "OriginalPlacerPlayerID", Type: arkproperty.TypeInt, Value: int32(10)},
		{Name: "OwnerName", Type: arkproperty.TypeString, Value: "Builders"},
		{Name: "OwningPlayerName", Type: arkproperty.TypeString, Value: "Ada"},
		{Name: "OwningPlayerID", Type: arkproperty.TypeInt, Value: int32(20)},
		{Name: "TargetingTeam", Type: arkproperty.TypeInt, Value: int32(30)},
	}})

	if owner.OriginalPlacerID != 10 || owner.TribeName != "Builders" || owner.PlayerName != "Ada" || owner.PlayerID != 20 || owner.TribeID != 30 {
		t.Fatalf("ObjectOwner = %#v", owner)
	}
}

func TestObjectOwnerEqualMatchesWhenKnownIDsAgree(t *testing.T) {
	left := ObjectOwner{PlayerID: 20, TribeID: 30}
	right := ObjectOwner{PlayerID: 20, TribeID: 30}
	if !left.Equal(right) {
		t.Fatalf("Equal() = false, want true")
	}
	right.TribeID = 31
	if left.Equal(right) {
		t.Fatalf("Equal() = true for different tribe IDs")
	}
}
