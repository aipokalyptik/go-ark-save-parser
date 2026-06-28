package arkobject

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

type Tribe struct {
	Name      string
	OwnerID   int32
	TribeID   int32
	Members   []string
	MemberIDs []int32
	TribeLog  []string
	LogIndex  int32
	NumDinos  int32
}

func (t Tribe) HasName() bool {
	return t.Name != ""
}

func (t Tribe) MemberCount() int {
	if len(t.MemberIDs) > len(t.Members) {
		return len(t.MemberIDs)
	}
	return len(t.Members)
}

func TribeFromContainer(properties arkproperty.Container) (Tribe, error) {
	raw, ok := properties.Value("TribeData")
	if !ok {
		return Tribe{}, fmt.Errorf("missing TribeData")
	}
	tribeData, ok := raw.(arkproperty.Container)
	if !ok {
		return Tribe{}, fmt.Errorf("TribeData has type %T, want arkproperty.Container", raw)
	}
	return Tribe{
		Name:      stringValue(tribeData, "TribeName"),
		OwnerID:   int32Value(tribeData, "OwnerPlayerDataId"),
		TribeID:   int32Value(tribeData, "TribeID"),
		Members:   stringArrayValue(tribeData, "MembersPlayerName"),
		MemberIDs: int32ArrayValue(tribeData, "MembersPlayerDataID"),
		TribeLog:  stringArrayValue(tribeData, "TribeLog"),
		LogIndex:  int32Value(tribeData, "LogIndex"),
		NumDinos:  int32Value(tribeData, "NumTribeDinos"),
	}, nil
}
