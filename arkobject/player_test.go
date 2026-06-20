package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

func TestPlayerFromContainerReadsMyDataFields(t *testing.T) {
	myData := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "PlayerDataID", Type: arkproperty.TypeInt, Value: int32(42)},
		{Name: "PlayerCharacterName", Type: arkproperty.TypeString, Value: "Survivor"},
		{Name: "PlayerName", Type: arkproperty.TypeString, Value: "PlatformName"},
		{Name: "TribeID", Type: arkproperty.TypeInt, Value: int32(777)},
		{Name: "SavedNetworkAddress", Type: arkproperty.TypeString, Value: "127.0.0.1"},
		{Name: "bFirstSpawned", Type: arkproperty.TypeBool, Value: true},
		{Name: "NumOfDeaths", Type: arkproperty.TypeInt, Value: int32(3)},
		{Name: "LastTimeDiedToEnemyTeam", Type: arkproperty.TypeDouble, Value: float64(12.5)},
		{Name: "LoginTime", Type: arkproperty.TypeDouble, Value: float64(99.25)},
		{Name: "UniqueID", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{
			Type:  arkproperty.ObjectReferencePath,
			Value: "EOS:abc123",
		}},
	}}
	props := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "SavedPlayerDataVersion", Type: arkproperty.TypeInt, Value: int32(17)},
		{Name: "CharacterStatusComponent_ExtraCharacterLevel", Type: arkproperty.TypeInt, Value: int32(4)},
		{Name: "CharacterStatusComponent_ExperiencePoints", Type: arkproperty.TypeFloat, Value: float32(123.5)},
		{Name: "PlayerState_TotalEngramPoints", Type: arkproperty.TypeInt, Value: int32(12)},
		{Name: "PlayerState_EngramBlueprints", Type: arkproperty.TypeArray, Value: arkproperty.Array{
			ElementType: arkproperty.TypeObject,
			Values: []any{
				arkproperty.ObjectReference{Type: arkproperty.ObjectReferencePath, Value: "Blueprint'/Game/Engrams/EngramA.EngramA_C'"},
				"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
			},
		}},
		{Name: "MyData", Type: arkproperty.TypeStruct, Value: myData},
	}}

	player, err := PlayerFromContainer(props)
	if err != nil {
		t.Fatalf("PlayerFromContainer() error = %v", err)
	}
	if player.PlayerDataID != 42 || player.CharacterName != "Survivor" || player.PlayerName != "PlatformName" || player.TribeID != 777 {
		t.Fatalf("Player = %#v", player)
	}
	if player.UniqueID != "EOS:abc123" || player.NumDeaths != 3 || !player.FirstSpawned {
		t.Fatalf("Player secondary fields = %#v", player)
	}
	if player.LastTimeDied != 12.5 || player.LoginTime != 99.25 || player.PlayerDataVersion != 17 {
		t.Fatalf("Player time/version fields = %#v", player)
	}
	if player.Level != 5 || player.Experience != 123.5 || player.EngramPoints != 12 {
		t.Fatalf("Player stat fields = %#v", player)
	}
	if len(player.UnlockedEngrams) != 2 || player.UnlockedEngrams[0] != "Blueprint'/Game/Engrams/EngramA.EngramA_C'" || player.UnlockedEngrams[1] != "Blueprint'/Game/Engrams/EngramB.EngramB_C'" {
		t.Fatalf("UnlockedEngrams = %#v", player.UnlockedEngrams)
	}
}

func TestPlayerFromContainerReadsNestedLocalProfileStats(t *testing.T) {
	characterConfig := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "PlayerCharacterName", Type: arkproperty.TypeString, Value: "NestedSurvivor"},
	}}
	persistentStats := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "CharacterStatusComponent_ExtraCharacterLevel", Type: arkproperty.TypeUInt16, Value: uint16(9)},
		{Name: "CharacterStatusComponent_ExperiencePoints", Type: arkproperty.TypeFloat, Value: float32(456.25)},
		{Name: "PlayerState_TotalEngramPoints", Type: arkproperty.TypeInt, Value: int32(22)},
		{Name: "PlayerState_EngramBlueprints", Type: arkproperty.TypeArray, Value: arkproperty.Array{
			ElementType: arkproperty.TypeObject,
			Values: []any{
				arkproperty.ObjectReference{Type: arkproperty.ObjectReferencePath, Value: "Blueprint'/Game/Engrams/EngramA.EngramA_C'"},
			},
		}},
	}}
	myData := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "PlayerDataID", Type: arkproperty.TypeUInt64, Value: uint64(42)},
		{Name: "PlayerName", Type: arkproperty.TypeString, Value: "PlatformName"},
		{Name: "MyPlayerCharacterConfig", Type: arkproperty.TypeStruct, Value: characterConfig},
		{Name: "MyPersistentCharacterStats", Type: arkproperty.TypeStruct, Value: persistentStats},
		{Name: "NumOfDeaths", Type: arkproperty.TypeFloat, Value: float32(7)},
	}}
	props := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "SavedPlayerDataVersion", Type: arkproperty.TypeInt, Value: int32(17)},
		{Name: "MyData", Type: arkproperty.TypeStruct, Value: myData},
	}}

	player, err := PlayerFromContainer(props)
	if err != nil {
		t.Fatalf("PlayerFromContainer() error = %v", err)
	}
	if player.CharacterName != "NestedSurvivor" || player.NumDeaths != 7 {
		t.Fatalf("nested profile identity fields = %#v", player)
	}
	if player.Level != 10 || player.Experience != 456.25 || player.EngramPoints != 22 {
		t.Fatalf("nested profile stat fields = %#v", player)
	}
	if len(player.UnlockedEngrams) != 1 || player.UnlockedEngrams[0] != "Blueprint'/Game/Engrams/EngramA.EngramA_C'" {
		t.Fatalf("nested profile engrams = %#v", player.UnlockedEngrams)
	}
}

func TestTribeFromContainerReadsTribeDataFields(t *testing.T) {
	tribeData := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "TribeName", Type: arkproperty.TypeString, Value: "Builders"},
		{Name: "OwnerPlayerDataId", Type: arkproperty.TypeUInt32, Value: uint32(42)},
		{Name: "TribeID", Type: arkproperty.TypeInt, Value: int32(777)},
		{Name: "MembersPlayerName", Type: arkproperty.TypeArray, Value: arkproperty.Array{
			ElementType: arkproperty.TypeString,
			Values:      []any{"Ada", "Grace"},
		}},
		{Name: "MembersPlayerDataID", Type: arkproperty.TypeArray, Value: arkproperty.Array{
			ElementType: arkproperty.TypeInt,
			Values:      []any{int32(42), int32(43)},
		}},
		{Name: "TribeLog", Type: arkproperty.TypeArray, Value: arkproperty.Array{
			ElementType: arkproperty.TypeString,
			Values:      []any{"created", "built"},
		}},
		{Name: "LogIndex", Type: arkproperty.TypeInt, Value: int32(2)},
		{Name: "NumTribeDinos", Type: arkproperty.TypeInt, Value: int32(9)},
	}}
	props := arkproperty.Container{Properties: []arkproperty.Property{
		{Name: "TribeData", Type: arkproperty.TypeStruct, Value: tribeData},
	}}

	tribe, err := TribeFromContainer(props)
	if err != nil {
		t.Fatalf("TribeFromContainer() error = %v", err)
	}
	if tribe.Name != "Builders" || tribe.OwnerID != 42 || tribe.TribeID != 777 || tribe.NumDinos != 9 || tribe.LogIndex != 2 {
		t.Fatalf("Tribe = %#v", tribe)
	}
	if len(tribe.Members) != 2 || tribe.Members[0] != "Ada" || len(tribe.MemberIDs) != 2 || tribe.MemberIDs[1] != 43 {
		t.Fatalf("Tribe members = %#v ids=%#v", tribe.Members, tribe.MemberIDs)
	}
	if len(tribe.TribeLog) != 2 || tribe.TribeLog[1] != "built" {
		t.Fatalf("TribeLog = %#v", tribe.TribeLog)
	}
}
