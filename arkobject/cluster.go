package arkobject

import (
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

type ClusterItemType string

const (
	ClusterItemTypeDino      ClusterItemType = "dino"
	ClusterItemTypeEquipment ClusterItemType = "equipment"
	ClusterItemTypeOther     ClusterItemType = "other"
)

type ClusterDinoParseStatus string

const (
	ClusterDinoParseStatusParsed             ClusterDinoParseStatus = "parsed"
	ClusterDinoParseStatusUnsupportedVersion ClusterDinoParseStatus = "unsupported_version"
	ClusterDinoParseStatusParseError         ClusterDinoParseStatus = "parse_error"
	ClusterDinoParseStatusUnparsed           ClusterDinoParseStatus = "unparsed"
)

func (t ClusterItemType) String() string {
	return string(t)
}

func (s ClusterDinoParseStatus) String() string {
	return string(s)
}

type ClusterItem struct {
	Index                int
	Type                 string
	Version              float64
	UploadTime           float64
	Blueprint            string
	Quantity             int32
	Rating               float64
	Quality              int32
	CrafterCharacterName string
	CrafterTribeName     string
	Properties           arkproperty.Container
}

type ClusterDino struct {
	Index                        int
	Version                      float64
	UploadTime                   float64
	RawSize                      int
	ObjectCount                  int
	ParsedArchive                bool
	ClassNames                   []string
	StatusComponentClassNames    []string
	AIControllerClassNames       []string
	InventoryComponentClassNames []string
	ParseError                   string
	Properties                   arkproperty.Container
}

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

func ClusterDinoFromUpload(dino arkcluster.Dino, classNames []string) ClusterDino {
	objectCount := 0
	if dino.Archive != nil {
		objectCount = len(dino.Archive.Objects)
	}
	return ClusterDino{
		Index:                        dino.Index,
		Version:                      dino.Version,
		UploadTime:                   dino.UploadTime,
		RawSize:                      dino.RawSize,
		ObjectCount:                  objectCount,
		ParsedArchive:                dino.Archive != nil,
		ClassNames:                   append([]string(nil), classNames...),
		StatusComponentClassNames:    clusterClassNamesContainingAny(classNames, "CharacterStatusComponent", "DinoCharacterStatus", "CharacterStatus"),
		AIControllerClassNames:       clusterClassNamesContaining(classNames, "AIController"),
		InventoryComponentClassNames: clusterClassNamesContaining(classNames, "InventoryComponent"),
		ParseError:                   dino.ParseError,
		Properties:                   dino.Properties,
	}
}

func (d ClusterDino) Parsed() bool {
	return d.ParsedArchive && d.ParseError == ""
}

func (d ClusterDino) HasParseError() bool {
	return d.ParseError != ""
}

func (d ClusterDino) SupportedVersion() bool {
	return d.Version == 7
}

func (d ClusterDino) UnsupportedVersion() bool {
	return d.Version != 0 && !d.SupportedVersion()
}

func (d ClusterDino) ParseStatus() ClusterDinoParseStatus {
	switch {
	case d.HasParseError():
		return ClusterDinoParseStatusParseError
	case d.UnsupportedVersion():
		return ClusterDinoParseStatusUnsupportedVersion
	case d.ParsedArchive:
		return ClusterDinoParseStatusParsed
	default:
		return ClusterDinoParseStatusUnparsed
	}
}

func clusterClassNamesContaining(classNames []string, token string) []string {
	out := make([]string, 0)
	for _, className := range classNames {
		if strings.Contains(className, token) {
			out = append(out, className)
		}
	}
	return out
}

func clusterClassNamesContainingAny(classNames []string, tokens ...string) []string {
	out := make([]string, 0)
	for _, className := range classNames {
		for _, token := range tokens {
			if strings.Contains(className, token) {
				out = append(out, className)
				break
			}
		}
	}
	return out
}
