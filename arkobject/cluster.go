package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

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
	Index       int
	Version     float64
	UploadTime  float64
	RawSize     int
	ObjectCount int
	ClassNames  []string
	ParseError  string
	Properties  arkproperty.Container
}

func ClusterItemFromUpload(item arkcluster.Item, typeName string) ClusterItem {
	return ClusterItem{
		Index:                item.Index,
		Type:                 typeName,
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
		Index:       dino.Index,
		Version:     dino.Version,
		UploadTime:  dino.UploadTime,
		RawSize:     dino.RawSize,
		ObjectCount: objectCount,
		ClassNames:  append([]string(nil), classNames...),
		ParseError:  dino.ParseError,
		Properties:  dino.Properties,
	}
}
