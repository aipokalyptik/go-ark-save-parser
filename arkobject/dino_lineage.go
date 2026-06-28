package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

func babyStageForPercent(percent float64) BabyStage {
	if percent < 10 {
		return BabyStageBaby
	}
	if percent < 50 {
		return BabyStageJuvenile
	}
	return BabyStageAdolescent
}

func ancestorIDs(properties arkproperty.Container) []DinoID {
	var out []DinoID
	for _, name := range []string{"DinoAncestors", "DinoAncestorsMale"} {
		value, ok := properties.Value(name)
		if !ok {
			continue
		}
		array, ok := value.(arkproperty.Array)
		if !ok {
			continue
		}
		for _, entry := range array.Values {
			container, ok := entry.(arkproperty.Container)
			if !ok {
				continue
			}
			female := DinoID{
				ID1: uint32Value(container, "FemaleDinoID1"),
				ID2: uint32Value(container, "FemaleDinoID2"),
			}
			if !female.IsZero() {
				out = append(out, female)
			}
			male := DinoID{
				ID1: uint32Value(container, "MaleDinoID1"),
				ID2: uint32Value(container, "MaleDinoID2"),
			}
			if !male.IsZero() {
				out = append(out, male)
			}
		}
	}
	return out
}

func (id DinoID) IsZero() bool {
	return id.ID1 == 0 && id.ID2 == 0
}
