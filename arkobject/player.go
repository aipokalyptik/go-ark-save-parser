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
