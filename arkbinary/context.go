package arkbinary

import "fmt"

// Context carries the save-level name tables needed while reading names.
type Context struct {
	names           map[uint32]string
	constantNames   map[uint32]string
	GenerateUnknown bool
}

func NewContext() *Context {
	return &Context{
		names:         map[uint32]string{},
		constantNames: map[uint32]string{},
	}
}

func (c *Context) SetNames(names map[uint32]string) {
	c.names = make(map[uint32]string, len(names))
	for k, v := range names {
		c.names[k] = v
	}
}

func (c *Context) UseConstantNameTable(names map[uint32]string) {
	c.constantNames = make(map[uint32]string, len(names))
	for k, v := range names {
		c.constantNames[k] = v
	}
}

func (c *Context) HasNameTable() bool {
	return len(c.names) > 0 || len(c.constantNames) > 0
}

func (c *Context) Name(id uint32) (string, bool) {
	if name, ok := c.names[id]; ok {
		return name, true
	}
	if name, ok := c.constantNames[id]; ok {
		return name, true
	}
	if c.GenerateUnknown {
		name := fmt.Sprintf("Unknown_%d", id)
		c.names[id] = name
		return name, true
	}
	return "", false
}
