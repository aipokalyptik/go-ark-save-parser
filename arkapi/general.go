package arkapi

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type GeneralAPI struct {
	save *arksave.Save
}

type ObjectSummary struct {
	Exists     bool
	Bytes      int
	Properties int
}

type ClassPropertySummary struct {
	Objects    int
	Properties int
}

type ClassLookupSummary struct {
	Objects int
	Classes int
}

type PropertyFilterSummary struct {
	Objects int
	Classes int
}

type PropertyPositionSummary struct {
	Exists       bool
	Properties   int
	NameOffsets  int
	ValueOffsets int
	Encoded      int
	Positioned   int
	OffsetsOK    int
}

type ParseSummary struct {
	Objects int
	Parsed  int
	Faults  int
}

type generalPathError struct {
	op  string
	err error
}

func (e generalPathError) Error() string {
	return e.op + ": " + e.err.Error()
}

func (e generalPathError) Unwrap() error {
	return e.err
}

func NewGeneral(save *arksave.Save) *GeneralAPI {
	return &GeneralAPI{save: save}
}

func NewGeneralFromPath(savePath string) (*GeneralAPI, func() error, error) {
	save, err := arksave.Open(savePath)
	if err != nil {
		return nil, nil, err
	}
	return NewGeneral(save), save.Close, nil
}

func GeneralClassesFromPath(savePath string) ([]string, error) {
	api, closeAPI, err := NewGeneralFromPath(savePath)
	if err != nil {
		return nil, err
	}
	defer closeAPI()

	return api.Classes()
}

func GeneralParseSummaryFromPath(savePath string) (ParseSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewGeneralFromPath(savePath)
	if err != nil {
		return ParseSummary{}, nil, generalPathError{op: "open save", err: err}
	}
	defer closeAPI()

	summary, faults, err := api.ParseSummaryWithFaults()
	if err != nil {
		return ParseSummary{}, nil, generalPathError{op: "parse objects", err: err}
	}
	return summary, faults, nil
}

func (g *GeneralAPI) SaveInfo() (SaveInfo, error) {
	return NewJSON(g.save).ExportSaveInfo()
}

func (g *GeneralAPI) ObjectIDs() ([]uuid.UUID, error) {
	return g.save.ObjectIDs()
}

func (g *GeneralAPI) Classes() ([]string, error) {
	return g.save.Classes()
}

func (g *GeneralAPI) ParseSummaryWithFaults() (ParseSummary, []arksave.FaultyObjectInfo, error) {
	ids, err := g.ObjectIDs()
	if err != nil {
		return ParseSummary{}, nil, err
	}
	objects, faults, err := g.ObjectsWithFaults()
	if err != nil {
		return ParseSummary{}, nil, err
	}
	return ParseSummary{Objects: len(ids), Parsed: len(objects), Faults: len(faults)}, faults, nil
}

func (g *GeneralAPI) Objects() ([]*arkobject.GameObject, error) {
	ids, err := g.ObjectIDs()
	if err != nil {
		return nil, err
	}

	objects := make([]*arkobject.GameObject, 0, len(ids))
	for _, id := range ids {
		obj, err := g.Object(id)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func (g *GeneralAPI) ObjectsWithFaults() ([]*arkobject.GameObject, []arksave.FaultyObjectInfo, error) {
	infos, faults, err := g.save.ParsedObjectsWithFaults(nil)
	if err != nil {
		return nil, nil, err
	}
	return parsedObjects(infos), faults, nil
}

func (g *GeneralAPI) Object(id uuid.UUID) (*arkobject.GameObject, error) {
	return g.save.Object(id)
}

func (g *GeneralAPI) ObjectSummary(id uuid.UUID) (ObjectSummary, error) {
	raw, err := g.save.ObjectBinary(id)
	if errors.Is(err, sql.ErrNoRows) {
		return ObjectSummary{}, nil
	}
	if err != nil {
		return ObjectSummary{}, err
	}
	object, err := g.save.Object(id)
	if err != nil {
		return ObjectSummary{}, err
	}
	return ObjectSummary{Exists: true, Bytes: len(raw), Properties: len(object.Properties)}, nil
}

func (g *GeneralAPI) ClassPropertySummaryWithFaults(classSubstring string) (ClassPropertySummary, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := g.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return strings.Contains(info.ClassName, classSubstring)
	})
	if err != nil {
		return ClassPropertySummary{}, nil, err
	}
	properties := map[string]struct{}{}
	for _, info := range objects {
		for _, property := range info.Object.Properties {
			properties[property.Name] = struct{}{}
		}
	}
	return ClassPropertySummary{Objects: len(objects), Properties: len(properties)}, faults, nil
}

func (g *GeneralAPI) ClassLookupSummaryWithFaults(classSubstrings []string) (ClassLookupSummary, []arksave.FaultyObjectInfo, error) {
	infos, faults, err := g.save.SelectedObjectPropertiesWithFaults(func(info arksave.ObjectClassInfo) bool {
		for _, substr := range classSubstrings {
			if strings.Contains(info.ClassName, substr) {
				return true
			}
		}
		return false
	}, []string{"StructureID", "bIsEngram"})
	if err != nil {
		return ClassLookupSummary{}, nil, err
	}
	objects := 0
	classes := map[string]struct{}{}
	for _, info := range infos {
		container := arkproperty.Container{Properties: info.Properties}
		if _, ok := container.Value("StructureID"); !ok {
			continue
		}
		if selectedBoolProperty(container, "bIsEngram") {
			continue
		}
		objects++
		classes[info.ClassName] = struct{}{}
	}
	return ClassLookupSummary{Objects: objects, Classes: len(classes)}, faults, nil
}

func (g *GeneralAPI) PropertyFilterSummary(propertyNames []string) (PropertyFilterSummary, error) {
	objects, err := g.save.ObjectClassInfosWithAnyProperty(propertyNames)
	if err != nil {
		return PropertyFilterSummary{}, err
	}
	classes := map[string]struct{}{}
	for _, object := range objects {
		classes[object.ClassName] = struct{}{}
	}
	return PropertyFilterSummary{Objects: len(objects), Classes: len(classes)}, nil
}

func (g *GeneralAPI) PropertyPositionSummary(id uuid.UUID) (PropertyPositionSummary, error) {
	object, err := g.save.Object(id)
	if errors.Is(err, sql.ErrNoRows) {
		return PropertyPositionSummary{}, nil
	}
	if err != nil {
		return PropertyPositionSummary{}, err
	}
	summary := PropertyPositionSummary{Exists: true, Properties: len(object.Properties)}
	for _, property := range object.Properties {
		if property.NameOffset > 0 {
			summary.NameOffsets++
		}
		if property.ValueOffset > 0 {
			summary.ValueOffsets++
		}
		if len(property.EncodedBytes) > 0 {
			summary.Encoded++
		}
		if property.Position != 0 {
			summary.Positioned++
		}
		if property.NameOffset >= 0 && property.ValueOffset > property.NameOffset && len(property.EncodedBytes) > 0 {
			summary.OffsetsOK++
		}
	}
	return summary, nil
}

func (g *GeneralAPI) ObjectsWithAnyProperty(names []string) ([]*arkobject.GameObject, error) {
	infos, err := g.save.ParsedObjectsWithAnyProperty(names)
	if err != nil {
		return nil, err
	}
	return parsedObjects(infos), nil
}

func (g *GeneralAPI) ObjectsWithAnyPropertyWithFaults(names []string) ([]*arkobject.GameObject, []arksave.FaultyObjectInfo, error) {
	infos, faults, err := g.save.ParsedObjectsWithAnyPropertyWithFaults(names)
	if err != nil {
		return nil, nil, err
	}
	return parsedObjects(infos), faults, nil
}

func parsedObjects(infos []arksave.ParsedObjectInfo) []*arkobject.GameObject {
	objects := make([]*arkobject.GameObject, 0, len(infos))
	for _, info := range infos {
		objects = append(objects, info.Object)
	}
	return objects
}
