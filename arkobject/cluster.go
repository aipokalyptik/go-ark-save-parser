package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

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
	DinoID1                      uint32
	DinoID2                      uint32
	TamedName                    string
	IsTamed                      bool
	IsFemale                     bool
	IsBaby                       bool
	IsDead                       bool
	HasStats                     bool
	BaseLevel                    int32
	CurrentLevel                 int32
	ParseError                   string
	Properties                   arkproperty.Container
}
