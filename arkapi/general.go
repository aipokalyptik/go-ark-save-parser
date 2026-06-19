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

func (g *GeneralAPI) Object(id uuid.UUID) (*arkobject.GameObject, error) {
	return g.save.Object(id)
}
