package arkobject

type ObjectCrafter struct {
	CharacterName string
	TribeName     string
}

func (c ObjectCrafter) Valid() bool {
	return c.CharacterName != "" || c.TribeName != ""
}
