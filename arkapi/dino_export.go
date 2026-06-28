package arkapi

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/google/uuid"
)

type DinoBinaryExport struct {
	DinoCount  int                    `json:"dino_count"`
	RowCount   int                    `json:"row_count"`
	FaultCount int                    `json:"fault_count,omitempty"`
	Dinos      []DinoBinaryExportDino `json:"dinos"`
}

type DinoBinaryExportDino struct {
	UUID      string                 `json:"uuid"`
	Blueprint string                 `json:"blueprint"`
	Directory string                 `json:"directory"`
	Files     []DinoBinaryExportFile `json:"files"`
}

type DinoBinaryExportFile struct {
	Kind string `json:"kind"`
	UUID string `json:"uuid,omitempty"`
	Path string `json:"path"`
}

func (d *DinoAPI) ExportBinary(outputDir string) (DinoBinaryExport, error) {
	if outputDir == "" {
		return DinoBinaryExport{}, fmt.Errorf("dino export output directory is required")
	}
	dinos, faults, err := d.AllWithFaults()
	if err != nil {
		return DinoBinaryExport{}, err
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return DinoBinaryExport{}, err
	}
	export := DinoBinaryExport{
		FaultCount: len(faults),
		Dinos:      make([]DinoBinaryExportDino, 0, len(dinos)),
	}
	for _, id := range sortedUUIDKeys(dinos) {
		dino := dinos[id]
		if dino.IsCryopodded {
			continue
		}
		dinoExport, err := d.exportDinoBinary(outputDir, id, dino)
		if err != nil {
			return DinoBinaryExport{}, err
		}
		export.DinoCount++
		export.RowCount += countDinoBinaryRows(dinoExport.Files)
		export.Dinos = append(export.Dinos, dinoExport)
	}
	if err := writeBaseExportJSON(filepath.Join(outputDir, "manifest.json"), export); err != nil {
		return DinoBinaryExport{}, err
	}
	return export, nil
}

func ExportDinoBinaryFromPath(savePath string, outputDir string) (DinoBinaryExport, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return DinoBinaryExport{}, err
	}
	defer closeAPI()

	return api.ExportBinary(outputDir)
}

func (d *DinoAPI) exportDinoBinary(outputDir string, id uuid.UUID, dino arkobject.Dino) (DinoBinaryExportDino, error) {
	dinoDir := filepath.Join(outputDir, "dino_"+id.String())
	if err := os.MkdirAll(dinoDir, 0o700); err != nil {
		return DinoBinaryExportDino{}, err
	}
	out := DinoBinaryExportDino{
		UUID:      id.String(),
		Blueprint: dino.Blueprint,
		Directory: dinoDir,
	}
	if err := d.writeLinkedRow(dinoDir, "dino", id, &out.Files); err != nil {
		return DinoBinaryExportDino{}, err
	}
	if dino.StatusComponentUUID != nil {
		if err := d.writeLinkedRow(dinoDir, "status", *dino.StatusComponentUUID, &out.Files); err != nil {
			return DinoBinaryExportDino{}, err
		}
	}
	if dino.InventoryUUID != nil {
		if err := d.writeLinkedRow(dinoDir, "inv", *dino.InventoryUUID, &out.Files); err != nil {
			return DinoBinaryExportDino{}, err
		}
	}
	if dino.Location != nil {
		locationPath := filepath.Join(dinoDir, "dino_"+id.String()+"_location.json")
		if err := writeBaseExportJSON(locationPath, locationInfo(dino.Location)); err != nil {
			return DinoBinaryExportDino{}, err
		}
		out.Files = append(out.Files, DinoBinaryExportFile{Kind: "dino_location", UUID: id.String(), Path: locationPath})
	}
	return out, nil
}

func (d *DinoAPI) writeLinkedRow(dir string, kind string, id uuid.UUID, files *[]DinoBinaryExportFile) error {
	raw, err := d.save.ObjectBinary(id)
	if err != nil {
		return fmt.Errorf("read %s row %s: %w", kind, id, err)
	}
	path := filepath.Join(dir, kind+"_"+id.String()+".bin")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return err
	}
	*files = append(*files, DinoBinaryExportFile{Kind: kind + "_binary", UUID: id.String(), Path: path})
	return nil
}

func countDinoBinaryRows(files []DinoBinaryExportFile) int {
	count := 0
	for _, file := range files {
		if file.Kind == "dino_binary" || file.Kind == "status_binary" || file.Kind == "inv_binary" {
			count++
		}
	}
	return count
}
