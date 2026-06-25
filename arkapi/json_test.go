package arkapi

import (
	"encoding/json"
	"testing"
)

func TestJSONAPIExportSaveInfoSummarizesLocalSave(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewJSON(save)
	info, err := api.ExportSaveInfo()
	if err != nil {
		t.Fatalf("ExportSaveInfo() error = %v", err)
	}
	if info.MapName != "Valguero_WP" || info.SaveVersion != 12 || info.ObjectCount != 1 {
		t.Fatalf("SaveInfo = %#v", info)
	}
	if len(info.Objects) != 1 || info.Objects[0].ClassName != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("SaveInfo.Objects = %#v", info.Objects)
	}
}

func TestJSONAPIExportSaveInfoJSONIsDeterministic(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewJSON(save)
	data, err := api.ExportSaveInfoJSON()
	if err != nil {
		t.Fatalf("ExportSaveInfoJSON() error = %v", err)
	}
	var decoded SaveInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, data)
	}
	if decoded.ObjectCount != 1 || len(decoded.Objects) != 1 {
		t.Fatalf("decoded SaveInfo = %#v", decoded)
	}
}
