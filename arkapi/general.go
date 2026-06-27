package arkapi

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
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

func NewGeneral(save *arksave.Save) *GeneralAPI {
	return &GeneralAPI{save: save}
}

func (g *GeneralAPI) ObjectIDs() ([]uuid.UUID, error) {
	return g.save.ObjectIDs()
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
