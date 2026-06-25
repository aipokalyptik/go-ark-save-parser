package arkapi

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/google/uuid"
)

type EquipmentBinaryExport struct {
	ItemCount  int                         `json:"item_count"`
	RowCount   int                         `json:"row_count"`
	FaultCount int                         `json:"fault_count,omitempty"`
	Items      []EquipmentBinaryExportItem `json:"items"`
}

type EquipmentBinaryExportItem struct {
	UUID      string                  `json:"uuid"`
	Blueprint string                  `json:"blueprint"`
	Kind      arkobject.EquipmentKind `json:"kind"`
	Path      string                  `json:"path"`
}

func (e *EquipmentAPI) ExportBinary(outputDir string) (EquipmentBinaryExport, error) {
	if outputDir == "" {
		return EquipmentBinaryExport{}, fmt.Errorf("equipment export output directory is required")
	}
	items, faults, err := e.AllWithFaults()
	if err != nil {
		return EquipmentBinaryExport{}, err
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return EquipmentBinaryExport{}, err
	}
	export := EquipmentBinaryExport{
		FaultCount: len(faults),
		Items:      make([]EquipmentBinaryExportItem, 0, len(items)),
	}
	for _, id := range sortedUUIDKeys(items) {
		item := items[id]
		itemExport, err := e.exportEquipmentBinary(outputDir, id, item)
		if err != nil {
			return EquipmentBinaryExport{}, err
		}
		export.ItemCount++
		export.RowCount++
		export.Items = append(export.Items, itemExport)
	}
	if err := writeBaseExportJSON(filepath.Join(outputDir, "manifest.json"), export); err != nil {
		return EquipmentBinaryExport{}, err
	}
	return export, nil
}

func (e *EquipmentAPI) exportEquipmentBinary(outputDir string, id uuid.UUID, item arkobject.EquipmentItem) (EquipmentBinaryExportItem, error) {
	raw, err := e.save.ObjectBinary(id)
	if err != nil {
		return EquipmentBinaryExportItem{}, fmt.Errorf("read equipment row %s: %w", id, err)
	}
	path := filepath.Join(outputDir, "item_"+id.String()+".bin")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return EquipmentBinaryExportItem{}, err
	}
	return EquipmentBinaryExportItem{
		UUID:      id.String(),
		Blueprint: item.Blueprint,
		Kind:      item.Kind,
		Path:      path,
	}, nil
}
