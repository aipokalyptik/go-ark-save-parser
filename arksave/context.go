package arksave

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/google/uuid"
)

type Context struct {
	SaveVersion             int16
	GameTime                float64
	MapName                 string
	UnknownValue            uint32
	Sections                []HeaderLocation
	Names                   map[uint32]string
	ActorTransforms         map[uuid.UUID]arkobject.ActorTransform
	ActorTransformPositions map[uuid.UUID]int
}

func (c *Context) Name(id uint32) string {
	return c.Names[id]
}

type HeaderLocation struct {
	Raw string
}
