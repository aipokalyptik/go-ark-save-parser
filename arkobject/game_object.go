package arkobject

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

type GameObject struct {
	UUID       uuid.UUID
	Blueprint  string
	Names      []string
	Section    string
	Unknown    int16
	Properties []arkproperty.Property
}

func (g *GameObject) Value(name string) (any, bool) {
	for _, prop := range g.Properties {
		if prop.Name == name {
			return prop.Value, true
		}
	}
	return nil, false
}

func (g *GameObject) Container() arkproperty.Container {
	if g == nil {
		return arkproperty.Container{}
	}
	return arkproperty.Container{Properties: g.Properties}
}

func ParseGameObject(id uuid.UUID, data []byte, ctx *arkbinary.Context, sections []string) (*GameObject, error) {
	r := arkbinary.NewReader(data, ctx)
	blueprint, err := r.ReadName("")
	if err != nil {
		return nil, err
	}
	zero, err := r.ReadUInt32()
	if err != nil {
		return nil, err
	}
	if zero != 0 {
		return nil, fmt.Errorf("expected zero after blueprint name, got %#x", zero)
	}

	nameCount, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}
	if nameCount < 0 {
		return nil, fmt.Errorf("negative object name count %d", nameCount)
	}
	names := make([]string, 0, nameCount)
	for i := int32(0); i < nameCount; i++ {
		name, err := readObjectLocalName(r)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	partIndex, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}
	var section string
	if partIndex >= 0 && int(partIndex) < len(sections) {
		section = sections[partIndex]
	}

	unknown, err := r.ReadInt16()
	if err != nil {
		return nil, err
	}

	props, err := arkproperty.ParseAll(r, r.Size())
	if err != nil {
		return nil, err
	}

	return &GameObject{
		UUID:       id,
		Blueprint:  blueprint,
		Names:      names,
		Section:    section,
		Unknown:    unknown,
		Properties: props,
	}, nil
}

func readObjectLocalName(r *arkbinary.Reader) (string, error) {
	pos := r.Position()
	name, err := r.ReadName("")
	if err == nil {
		return name, nil
	}
	if setErr := r.SetPosition(pos); setErr != nil {
		return "", err
	}
	value, stringErr := r.ReadString()
	if stringErr != nil {
		return "", err
	}
	if value == nil {
		return "", nil
	}
	return *value, nil
}
