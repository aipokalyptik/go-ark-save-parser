package arkapi

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type GeneralAPI struct {
	save *arksave.Save
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
