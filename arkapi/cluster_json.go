package arkapi

import (
	"encoding/json"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

type ClusterDataInfo struct {
	ID             string            `json:"id"`
	Path           string            `json:"path"`
	ArchiveVersion int32             `json:"archive_version"`
	ObjectCount    int               `json:"object_count"`
	ItemCount      int               `json:"item_count"`
	DinoCount      int               `json:"dino_count"`
	Items          []ClusterItemInfo `json:"items"`
	Dinos          []ClusterDinoInfo `json:"dinos"`
}

type ClusterDirectoryInfo struct {
	Count int               `json:"count"`
	Files []ClusterDataInfo `json:"files"`
}

type ClusterItemInfo struct {
	Index                int     `json:"index"`
	Type                 string  `json:"type"`
	Version              float64 `json:"version"`
	UploadTime           float64 `json:"upload_time"`
	Blueprint            string  `json:"blueprint,omitempty"`
	Quantity             int32   `json:"quantity"`
	Rating               float64 `json:"rating,omitempty"`
	Quality              int32   `json:"quality,omitempty"`
	CrafterCharacterName string  `json:"crafter_character_name,omitempty"`
	CrafterTribeName     string  `json:"crafter_tribe_name,omitempty"`
}

type ClusterDinoInfo struct {
	Index       int     `json:"index"`
	Version     float64 `json:"version"`
	UploadTime  float64 `json:"upload_time"`
	RawSize     int     `json:"raw_size"`
	ObjectCount int     `json:"object_count"`
	ParseError  string  `json:"parse_error,omitempty"`
}

func ExportClusterData(data *arkcluster.Data) ClusterDataInfo {
	info := ClusterDataInfo{
		ID:        data.ID,
		Path:      data.Path,
		ItemCount: len(data.Items),
		DinoCount: len(data.Dinos),
		Items:     make([]ClusterItemInfo, 0, len(data.Items)),
		Dinos:     make([]ClusterDinoInfo, 0, len(data.Dinos)),
	}
	if data.Archive != nil {
		info.ArchiveVersion = data.Archive.Version
		info.ObjectCount = len(data.Archive.Objects)
	}
	for _, item := range data.Items {
		info.Items = append(info.Items, ClusterItemInfo{
			Index:                item.Index,
			Type:                 clusterItemType(item),
			Version:              item.Version,
			UploadTime:           item.UploadTime,
			Blueprint:            item.Blueprint,
			Quantity:             item.Quantity,
			Rating:               item.Rating,
			Quality:              item.Quality,
			CrafterCharacterName: item.CrafterCharacterName,
			CrafterTribeName:     item.CrafterTribeName,
		})
	}
	for _, dino := range data.Dinos {
		objectCount := 0
		if dino.Archive != nil {
			objectCount = len(dino.Archive.Objects)
		}
		info.Dinos = append(info.Dinos, ClusterDinoInfo{
			Index:       dino.Index,
			Version:     dino.Version,
			UploadTime:  dino.UploadTime,
			RawSize:     dino.RawSize,
			ObjectCount: objectCount,
			ParseError:  dino.ParseError,
		})
	}
	return info
}

func clusterItemType(item arkcluster.Item) string {
	switch {
	case NewDino(nil).IsApplicableBlueprint(item.Blueprint) && hasCustomItemDatas(item.Properties):
		return "dino"
	case NewEquipment(nil).IsApplicableBlueprint(item.Blueprint):
		return "equipment"
	default:
		return "other"
	}
}

func hasCustomItemDatas(container arkproperty.Container) bool {
	raw, ok := container.Value("CustomItemDatas")
	if !ok {
		return false
	}
	switch value := raw.(type) {
	case arkproperty.Array:
		return len(value.Values) > 0
	case []any:
		return len(value) > 0
	default:
		return false
	}
}

func ExportClusterDataJSON(data *arkcluster.Data) ([]byte, error) {
	return json.MarshalIndent(ExportClusterData(data), "", "  ")
}

func ExportClusterDirectoryData(entries []*arkcluster.Data) ClusterDirectoryInfo {
	info := ClusterDirectoryInfo{
		Count: len(entries),
		Files: make([]ClusterDataInfo, 0, len(entries)),
	}
	for _, entry := range entries {
		info.Files = append(info.Files, ExportClusterData(entry))
	}
	return info
}

func ExportClusterDirectoryDataJSON(entries []*arkcluster.Data) ([]byte, error) {
	return json.MarshalIndent(ExportClusterDirectoryData(entries), "", "  ")
}
