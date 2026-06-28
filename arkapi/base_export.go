package arkapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

type BaseBinaryExport struct {
	BaseCount      int                    `json:"base_count"`
	StructureCount int                    `json:"structure_count"`
	FaultCount     int                    `json:"fault_count,omitempty"`
	Bases          []BaseBinaryExportBase `json:"bases"`
}

type BaseBinaryExportBase struct {
	KeystoneUUID   string                 `json:"keystone_uuid"`
	StructureCount int                    `json:"structure_count"`
	Directory      string                 `json:"directory"`
	Files          []BaseBinaryExportFile `json:"files"`
}

type BaseBinaryExportFile struct {
	Kind string `json:"kind"`
	UUID string `json:"uuid,omitempty"`
	Path string `json:"path"`
}

func (b *BaseAPI) ExportBinary(outputDir string) (BaseBinaryExport, error) {
	if outputDir == "" {
		return BaseBinaryExport{}, errors.New("base export output directory is required")
	}
	bases, faults, err := b.AllWithFaults()
	if err != nil {
		return BaseBinaryExport{}, err
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return BaseBinaryExport{}, err
	}
	export := BaseBinaryExport{
		BaseCount:  len(bases),
		FaultCount: len(faults),
		Bases:      make([]BaseBinaryExportBase, 0, len(bases)),
	}
	for _, base := range bases {
		baseExport, err := b.exportBaseBinary(outputDir, base)
		if err != nil {
			return BaseBinaryExport{}, err
		}
		export.StructureCount += baseExport.StructureCount
		export.Bases = append(export.Bases, baseExport)
	}
	manifestPath := filepath.Join(outputDir, "manifest.json")
	if err := writeBaseExportJSON(manifestPath, export); err != nil {
		return BaseBinaryExport{}, err
	}
	return export, nil
}

func ExportBaseBinaryFromPath(savePath string, mapName string, outputDir string) (BaseBinaryExport, error) {
	api, closeAPI, err := NewBaseFromPath(savePath, mapName)
	if err != nil {
		return BaseBinaryExport{}, err
	}
	defer closeAPI()

	return api.ExportBinary(outputDir)
}

func (b *BaseAPI) exportBaseBinary(outputDir string, base arkobject.Base) (BaseBinaryExportBase, error) {
	baseDir := filepath.Join(outputDir, "base_"+base.KeystoneUUID.String())
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return BaseBinaryExportBase{}, err
	}
	out := BaseBinaryExportBase{
		KeystoneUUID:   base.KeystoneUUID.String(),
		StructureCount: base.StructureCount,
		Directory:      baseDir,
	}
	info := BaseInfo{
		KeystoneUUID:       base.KeystoneUUID.String(),
		StructureUUIDs:     sortedBaseStructureUUIDs(base),
		StructureCount:     base.StructureCount,
		Owner:              ownerInfo(base.Owner),
		Location:           locationInfo(base.Location),
		AverageLocation:    locationInfo(base.AverageLocation),
		MapLocation:        mapCoordsInfo(base.Location, b.mapName),
		AverageMapLocation: mapCoordsInfo(base.AverageLocation, b.mapName),
		TurretCount:        base.TurretCount,
	}
	baseJSONPath := filepath.Join(baseDir, "base.json")
	if err := writeBaseExportJSON(baseJSONPath, info); err != nil {
		return BaseBinaryExportBase{}, err
	}
	out.Files = append(out.Files, BaseBinaryExportFile{Kind: "base_json", Path: baseJSONPath})

	for _, id := range sortedUUIDKeys(base.Structures) {
		raw, err := b.structures.save.ObjectBinary(id)
		if err != nil {
			return BaseBinaryExportBase{}, fmt.Errorf("read structure %s: %w", id, err)
		}
		binPath := filepath.Join(baseDir, "str_"+id.String()+".bin")
		if err := os.WriteFile(binPath, raw, 0o600); err != nil {
			return BaseBinaryExportBase{}, err
		}
		out.Files = append(out.Files, BaseBinaryExportFile{Kind: "structure_binary", UUID: id.String(), Path: binPath})
		if structure := base.Structures[id]; structure.Location != nil {
			locationPath := filepath.Join(baseDir, "str_"+id.String()+"_location.json")
			if err := writeBaseExportJSON(locationPath, locationInfo(structure.Location)); err != nil {
				return BaseBinaryExportBase{}, err
			}
			out.Files = append(out.Files, BaseBinaryExportFile{Kind: "structure_location", UUID: id.String(), Path: locationPath})
		}
	}
	return out, nil
}

func writeBaseExportJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}
