package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkcluster"

func (i ClusterItem) IsDinoUpload() bool {
	return i.ItemType() == ClusterItemTypeDino
}

func (i ClusterItem) IsEquipmentUpload() bool {
	return i.ItemType() == ClusterItemTypeEquipment
}

func (i ClusterItem) IsOtherUpload() bool {
	return i.ItemType() == ClusterItemTypeOther
}

// ItemType normalizes the known upload type strings and treats any unknown
// string as "other"; the raw Type field remains available for compatibility.
func (i ClusterItem) ItemType() ClusterItemType {
	switch ClusterItemType(i.Type) {
	case ClusterItemTypeDino:
		return ClusterItemTypeDino
	case ClusterItemTypeEquipment:
		return ClusterItemTypeEquipment
	default:
		return ClusterItemTypeOther
	}
}

func (i ClusterItem) SupportedVersion() bool {
	return i.Version == 7
}

func (i ClusterItem) UnsupportedVersion() bool {
	return i.Version != 0 && !i.SupportedVersion()
}

func (i ClusterItem) Crafter() ObjectCrafter {
	return ObjectCrafter{
		CharacterName: i.CrafterCharacterName,
		TribeName:     i.CrafterTribeName,
	}
}

func (i ClusterItem) IsCrafted() bool {
	return i.Crafter().Valid()
}

func (i ClusterItem) ShortName() string {
	return ShortNameFromBlueprint(i.Blueprint)
}

func ClusterItemFromUpload(item arkcluster.Item, itemType ClusterItemType) ClusterItem {
	return ClusterItem{
		Index:                item.Index,
		Type:                 itemType.String(),
		Version:              item.Version,
		UploadTime:           item.UploadTime,
		Blueprint:            item.Blueprint,
		Quantity:             item.Quantity,
		Rating:               item.Rating,
		Quality:              item.Quality,
		CrafterCharacterName: item.CrafterCharacterName,
		CrafterTribeName:     item.CrafterTribeName,
		Properties:           item.Properties,
	}
}
