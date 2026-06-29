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

func TestNewJSONFromPathOpensLocalSave(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api, closeAPI, err := NewJSONFromPath(save.Path())
	if err != nil {
		t.Fatalf("NewJSONFromPath() error = %v", err)
	}
	defer closeAPI()

	info, err := api.ExportSaveInfo()
	if err != nil {
		t.Fatalf("ExportSaveInfo() error = %v", err)
	}
	if info.MapName != "Valguero_WP" || info.ObjectCount != 1 {
		t.Fatalf("SaveInfo = %#v, want synthetic save info", info)
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

func TestExportSaveInfoFromPathSummarizesLocalSave(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	info, err := ExportSaveInfoFromPath(save.Path())
	if err != nil {
		t.Fatalf("ExportSaveInfoFromPath() error = %v", err)
	}
	if info.MapName != "Valguero_WP" || info.SaveVersion != 12 || info.ObjectCount != 1 {
		t.Fatalf("SaveInfo = %#v", info)
	}
	if len(info.Objects) != 1 || info.Objects[0].ClassName != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("SaveInfo.Objects = %#v", info.Objects)
	}
}

func TestExportSaveInfoJSONFromPathHelpersMarshalLocalSave(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	raw, err := ExportSaveInfoJSONFromPath(save.Path())
	if err != nil {
		t.Fatalf("ExportSaveInfoJSONFromPath() error = %v", err)
	}
	var decoded SaveInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ObjectCount != 1 || len(decoded.Objects) != 1 {
		t.Fatalf("SaveInfo = %#v, want object detail", decoded)
	}

	redacted, err := ExportRedactedSaveInfoJSONFromPath(save.Path())
	if err != nil {
		t.Fatalf("ExportRedactedSaveInfoJSONFromPath() error = %v", err)
	}
	decoded = SaveInfo{}
	if err := json.Unmarshal(redacted, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(redacted) error = %v; data = %s", err, redacted)
	}
	if decoded.ObjectCount != 1 || len(decoded.Objects) != 0 {
		t.Fatalf("redacted SaveInfo = %#v, want count without object details", decoded)
	}
}
