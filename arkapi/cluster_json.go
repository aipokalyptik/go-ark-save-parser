package arkapi

import (
	"encoding/json"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
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
	Index      int     `json:"index"`
	Version    float64 `json:"version"`
	UploadTime float64 `json:"upload_time"`
	Blueprint  string  `json:"blueprint,omitempty"`
	Quantity   int32   `json:"quantity"`
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
			Index:      item.Index,
			Version:    item.Version,
			UploadTime: item.UploadTime,
			Blueprint:  item.Blueprint,
			Quantity:   item.Quantity,
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
