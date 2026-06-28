package arkapi

import (
	"encoding/json"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
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
	Count   int                       `json:"count"`
	Faults  []ClusterFileFaultInfo    `json:"faults,omitempty"`
	Summary ClusterDirectoryAggregate `json:"summary"`
	Files   []ClusterDataInfo         `json:"files"`
}

type ClusterFileFaultInfo struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

type ClusterItemInfo struct {
	Index                int     `json:"index"`
	Type                 string  `json:"type"`
	Version              float64 `json:"version"`
	SupportedVersion     bool    `json:"supported_version"`
	UnsupportedVersion   bool    `json:"unsupported_version"`
	UploadTime           float64 `json:"upload_time"`
	Blueprint            string  `json:"blueprint,omitempty"`
	ShortName            string  `json:"short_name,omitempty"`
	Quantity             int32   `json:"quantity"`
	Rating               float64 `json:"rating,omitempty"`
	Quality              int32   `json:"quality,omitempty"`
	CrafterCharacterName string  `json:"crafter_character_name,omitempty"`
	CrafterTribeName     string  `json:"crafter_tribe_name,omitempty"`
}

type ClusterDinoInfo struct {
	Index                        int      `json:"index"`
	Version                      float64  `json:"version"`
	SupportedVersion             bool     `json:"supported_version"`
	UnsupportedVersion           bool     `json:"unsupported_version"`
	UploadTime                   float64  `json:"upload_time"`
	RawSize                      int      `json:"raw_size"`
	ObjectCount                  int      `json:"object_count"`
	ParsedArchive                bool     `json:"parsed_archive"`
	ParseStatus                  string   `json:"parse_status"`
	PrimaryClassName             string   `json:"primary_class_name,omitempty"`
	ShortName                    string   `json:"short_name,omitempty"`
	ClassNames                   []string `json:"class_names,omitempty"`
	StatusComponentClassNames    []string `json:"status_component_class_names,omitempty"`
	AIControllerClassNames       []string `json:"ai_controller_class_names,omitempty"`
	InventoryComponentClassNames []string `json:"inventory_component_class_names,omitempty"`
	ParseError                   string   `json:"parse_error,omitempty"`
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
		typed := arkobject.ClusterItemFromUpload(item, clusterItemType(item))
		info.Items = append(info.Items, ClusterItemInfo{
			Index:                typed.Index,
			Type:                 typed.Type,
			Version:              typed.Version,
			SupportedVersion:     typed.SupportedVersion(),
			UnsupportedVersion:   typed.UnsupportedVersion(),
			UploadTime:           typed.UploadTime,
			Blueprint:            typed.Blueprint,
			ShortName:            typed.ShortName(),
			Quantity:             typed.Quantity,
			Rating:               typed.Rating,
			Quality:              typed.Quality,
			CrafterCharacterName: typed.CrafterCharacterName,
			CrafterTribeName:     typed.CrafterTribeName,
		})
	}
	for _, dino := range data.Dinos {
		var classNames []string
		if dino.Archive != nil {
			classNames = archiveClassNames(dino.Archive)
		}
		typed := arkobject.ClusterDinoFromUpload(dino, classNames)
		info.Dinos = append(info.Dinos, ClusterDinoInfo{
			Index:                        typed.Index,
			Version:                      typed.Version,
			SupportedVersion:             typed.SupportedVersion(),
			UnsupportedVersion:           typed.UnsupportedVersion(),
			UploadTime:                   typed.UploadTime,
			RawSize:                      typed.RawSize,
			ObjectCount:                  typed.ObjectCount,
			ParsedArchive:                typed.ParsedArchive,
			ParseStatus:                  typed.ParseStatus().String(),
			PrimaryClassName:             typed.PrimaryClassName(),
			ShortName:                    typed.ShortName(),
			ClassNames:                   typed.ClassNames,
			StatusComponentClassNames:    typed.StatusComponentClassNames,
			AIControllerClassNames:       typed.AIControllerClassNames,
			InventoryComponentClassNames: typed.InventoryComponentClassNames,
			ParseError:                   typed.ParseError,
		})
	}
	return info
}

func archiveClassNames(archive *arkarchive.Archive) []string {
	if archive == nil {
		return nil
	}
	names := make([]string, 0, len(archive.Objects))
	for _, object := range archive.Objects {
		if object.ClassName != "" {
			names = append(names, object.ClassName)
		}
	}
	return names
}

func clusterItemType(item arkcluster.Item) arkobject.ClusterItemType {
	switch {
	case NewDino(nil).IsApplicableBlueprint(item.Blueprint) && hasCustomItemDatas(item.Properties):
		return arkobject.ClusterItemTypeDino
	case NewEquipment(nil).IsApplicableBlueprint(item.Blueprint):
		return arkobject.ClusterItemTypeEquipment
	default:
		return arkobject.ClusterItemTypeOther
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
	return ExportClusterDirectoryDataWithFaults(entries, nil)
}

func ExportClusterDirectoryDataWithFaults(entries []*arkcluster.Data, faults []arkcluster.FileFault) ClusterDirectoryInfo {
	info := ClusterDirectoryInfo{
		Count:   len(entries),
		Summary: ClusterDirectorySummary(entries),
		Files:   make([]ClusterDataInfo, 0, len(entries)),
		Faults:  make([]ClusterFileFaultInfo, 0, len(faults)),
	}
	for _, entry := range entries {
		info.Files = append(info.Files, ExportClusterData(entry))
	}
	for _, fault := range faults {
		if fault.Err == nil {
			continue
		}
		info.Faults = append(info.Faults, ClusterFileFaultInfo{
			Path:  fault.Path,
			Error: fault.Err.Error(),
		})
	}
	return info
}

func ExportClusterDirectoryDataJSON(entries []*arkcluster.Data) ([]byte, error) {
	return ExportClusterDirectoryDataWithFaultsJSON(entries, nil)
}

func ExportClusterDirectoryDataWithFaultsJSON(entries []*arkcluster.Data, faults []arkcluster.FileFault) ([]byte, error) {
	return json.MarshalIndent(ExportClusterDirectoryDataWithFaults(entries, faults), "", "  ")
}
