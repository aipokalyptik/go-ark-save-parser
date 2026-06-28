package arkapi

import (
	"encoding/json"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

type JSONAPI struct {
	save *arksave.Save
}

func NewJSON(save *arksave.Save) *JSONAPI {
	return &JSONAPI{save: save}
}

func NewJSONFromPath(savePath string) (*JSONAPI, func() error, error) {
	save, err := arksave.Open(savePath)
	if err != nil {
		return nil, nil, err
	}
	return NewJSON(save), save.Close, nil
}

func ExportSaveInfoFromPath(savePath string) (SaveInfo, error) {
	api, closeAPI, err := NewJSONFromPath(savePath)
	if err != nil {
		return SaveInfo{}, err
	}
	defer closeAPI()

	return api.ExportSaveInfo()
}

type SaveInfo struct {
	MapName      string       `json:"map_name"`
	SaveVersion  int16        `json:"save_version"`
	GameTime     float64      `json:"game_time"`
	ObjectCount  int          `json:"object_count"`
	SectionCount int          `json:"section_count"`
	NameCount    int          `json:"name_count"`
	Objects      []ObjectInfo `json:"objects"`
}

type ObjectInfo struct {
	UUID      string   `json:"uuid"`
	ClassName string   `json:"class_name"`
	Names     []string `json:"names,omitempty"`
	Section   string   `json:"section,omitempty"`
}

func (j *JSONAPI) ExportSaveInfo() (SaveInfo, error) {
	objects, err := j.save.ObjectClassInfos()
	if err != nil {
		return SaveInfo{}, err
	}

	info := SaveInfo{
		MapName:      j.save.Context.MapName,
		SaveVersion:  j.save.Context.SaveVersion,
		GameTime:     j.save.Context.GameTime,
		ObjectCount:  len(objects),
		SectionCount: len(j.save.Context.Sections),
		NameCount:    len(j.save.Context.Names),
		Objects:      make([]ObjectInfo, 0, len(objects)),
	}
	for _, object := range objects {
		info.Objects = append(info.Objects, ObjectInfo{UUID: object.UUID.String(), ClassName: object.ClassName})
	}
	return info, nil
}

func (j *JSONAPI) ExportSaveInfoJSON() ([]byte, error) {
	info, err := j.ExportSaveInfo()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(info, "", "  ")
}
