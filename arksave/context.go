package arksave

type Context struct {
	SaveVersion  int16
	GameTime     float64
	MapName      string
	UnknownValue uint32
	Sections     []HeaderLocation
	Names        map[uint32]string
}

func (c *Context) Name(id uint32) string {
	return c.Names[id]
}

type HeaderLocation struct {
	Raw string
}
