package arkapi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/google/uuid"
)

type StructureBinaryExport struct {
	StructureCount int                         `json:"structure_count"`
	RowCount       int                         `json:"row_count"`
	FaultCount     int                         `json:"fault_count,omitempty"`
	Files          []StructureBinaryExportFile `json:"files"`
}

type StructureBinaryExportFile struct {
	Kind        string         `json:"kind"`
	UUID        string         `json:"uuid"`
	Blueprint   string         `json:"blueprint,omitempty"`
	Path        string         `json:"path"`
	Owner       OwnerInfo      `json:"owner"`
	Location    *LocationInfo  `json:"location,omitempty"`
	MapLocation *MapCoordsInfo `json:"map_location,omitempty"`
}

func (s *StructureAPI) ExportBinary(outputDir string) (StructureBinaryExport, error) {
	if outputDir == "" {
		return StructureBinaryExport{}, errors.New("structure export output directory is required")
	}
	structures, faults, err := s.AllWithFaults()
	if err != nil {
		return StructureBinaryExport{}, err
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return StructureBinaryExport{}, err
	}
	export := StructureBinaryExport{
		StructureCount: len(structures),
		FaultCount:     len(faults),
		Files:          make([]StructureBinaryExportFile, 0, len(structures)),
	}
	mapName := ""
	if s.save.Context != nil {
		mapName = s.save.Context.MapName
	}
	for _, id := range sortedUUIDKeys(structures) {
		file, err := s.exportStructureBinary(outputDir, id, structures[id], mapName)
		if err != nil {
			return StructureBinaryExport{}, err
		}
		export.RowCount++
		export.Files = append(export.Files, file)
	}
	manifestPath := filepath.Join(outputDir, "manifest.json")
	if err := writeBaseExportJSON(manifestPath, export); err != nil {
		return StructureBinaryExport{}, err
	}
	return export, nil
}

func ExportStructureBinaryFromPath(savePath string, outputDir string) (StructureBinaryExport, error) {
	api, closeAPI, err := NewStructureFromPath(savePath)
	if err != nil {
		return StructureBinaryExport{}, err
	}
	defer closeAPI()

	return api.ExportBinary(outputDir)
}

func (s *StructureAPI) exportStructureBinary(outputDir string, id uuid.UUID, structure arkobject.Structure, mapName string) (StructureBinaryExportFile, error) {
	raw, err := s.save.ObjectBinary(id)
	if err != nil {
		return StructureBinaryExportFile{}, fmt.Errorf("read structure %s: %w", id, err)
	}
	binPath := filepath.Join(outputDir, "str_"+id.String()+".bin")
	if err := os.WriteFile(binPath, raw, 0o600); err != nil {
		return StructureBinaryExportFile{}, err
	}
	file := StructureBinaryExportFile{
		Kind:      "structure_binary",
		UUID:      id.String(),
		Blueprint: structure.Blueprint,
		Path:      binPath,
		Owner:     ownerInfo(structure.Owner),
	}
	if location := structure.Location; location != nil {
		file.Location = locationInfo(location)
		file.MapLocation = mapCoordsInfo(location, mapName)
		locationPath := filepath.Join(outputDir, "str_"+id.String()+"_location.json")
		if err := writeBaseExportJSON(locationPath, file.Location); err != nil {
			return StructureBinaryExportFile{}, err
		}
	}
	return file, nil
}
