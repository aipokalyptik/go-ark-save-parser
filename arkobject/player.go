package arkobject

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

type Player struct {
	PlayerDataVersion int32
	PlayerDataID      uint64
	CharacterName     string
	PlayerName        string
	UniqueID          string
	IPAddress         string
	FirstSpawned      bool
	TribeID           int32
	NumDeaths         int32
	LastTimeDied      float64
	LoginTime         float64
	Level             int32
	Experience        float64
	EngramPoints      int32
	UnlockedEngrams   []string
}

func (p Player) DisplayName() string {
	if p.CharacterName != "" {
		return p.CharacterName
	}
	return p.PlayerName
}

func (p Player) HasName() bool {
	return p.DisplayName() != ""
}

func (p Player) InTribe() bool {
	return p.TribeID != 0
}

func PlayerFromContainer(properties arkproperty.Container) (Player, error) {
	var player Player
	player.PlayerDataVersion = int32Value(properties, "SavedPlayerDataVersion")

	raw, ok := properties.Value("MyData")
	if !ok {
		return Player{}, fmt.Errorf("missing MyData")
	}
	myData, ok := raw.(arkproperty.Container)
	if !ok {
		return Player{}, fmt.Errorf("MyData has type %T, want arkproperty.Container", raw)
	}

	stats := properties
	if nestedStats, ok := containerValue(myData, "MyPersistentCharacterStats"); ok {
		stats = nestedStats
	}
	player.Level = 1 + int32Value(stats, "CharacterStatusComponent_ExtraCharacterLevel")
	player.Experience = float64Value(stats, "CharacterStatusComponent_ExperiencePoints")
	player.EngramPoints = int32Value(stats, "PlayerState_TotalEngramPoints")
	player.UnlockedEngrams = objectReferenceStringArrayValue(stats, "PlayerState_EngramBlueprints")

	player.PlayerDataID = uint64Value(myData, "PlayerDataID")
	player.CharacterName = stringValue(myData, "PlayerCharacterName")
	if player.CharacterName == "" {
		if characterConfig, ok := containerValue(myData, "MyPlayerCharacterConfig"); ok {
			player.CharacterName = stringValue(characterConfig, "PlayerCharacterName")
		}
	}
	player.PlayerName = stringValue(myData, "PlayerName")
	player.IPAddress = stringValue(myData, "SavedNetworkAddress")
	player.FirstSpawned = boolValue(myData, "bFirstSpawned")
	player.TribeID = int32Value(myData, "TribeID")
	player.NumDeaths = int32Value(myData, "NumOfDeaths")
	player.LastTimeDied = float64Value(myData, "LastTimeDiedToEnemyTeam")
	player.LoginTime = float64Value(myData, "LoginTime")
	player.UniqueID = uniqueIDValue(myData)
	return player, nil
}

func containerValue(properties arkproperty.Container, name string) (arkproperty.Container, bool) {
	value, ok := properties.Value(name)
	if !ok {
		return arkproperty.Container{}, false
	}
	container, ok := value.(arkproperty.Container)
	return container, ok
}

func uint64Value(properties arkproperty.Container, name string) uint64 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case uint64:
		return v
	case uint32:
		return uint64(v)
	case int32:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int:
		if v < 0 {
			return 0
		}
		return uint64(v)
	default:
		return 0
	}
}

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

func uniqueIDValue(properties arkproperty.Container) string {
	value, ok := properties.Value("UniqueID")
	if !ok {
		return ""
	}
	switch v := value.(type) {
	case arkproperty.ObjectReference:
		text, _ := v.Value.(string)
		return text
	case string:
		return v
	default:
		return ""
	}
}

func stringArrayValue(properties arkproperty.Container, name string) []string {
	raw, ok := properties.Value(name)
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(array.Values))
	for _, value := range array.Values {
		if text, ok := value.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func objectReferenceStringArrayValue(properties arkproperty.Container, name string) []string {
	raw, ok := properties.Value(name)
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(array.Values))
	for _, value := range array.Values {
		switch v := value.(type) {
		case arkproperty.ObjectReference:
			if text, ok := v.Value.(string); ok && text != "" {
				out = append(out, text)
			}
		case string:
			if v != "" {
				out = append(out, v)
			}
		}
	}
	return out
}

func int32ArrayValue(properties arkproperty.Container, name string) []int32 {
	raw, ok := properties.Value(name)
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]int32, 0, len(array.Values))
	for _, value := range array.Values {
		switch v := value.(type) {
		case int32:
			out = append(out, v)
		case uint32:
			out = append(out, int32(v))
		case int:
			out = append(out, int32(v))
		}
	}
	return out
}
