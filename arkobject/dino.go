package arkobject

import (
	"strconv"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

type BabyStage string

const (
	BabyStageUnknown    BabyStage = ""
	BabyStageBaby       BabyStage = "Baby"
	BabyStageJuvenile   BabyStage = "Juvenile"
	BabyStageAdolescent BabyStage = "Adolescent"
)

type Dino struct {
	UUID                   uuid.UUID
	Blueprint              string
	Object                 *GameObject
	ID1                    uint32
	ID2                    uint32
	IsFemale               bool
	IsTamed                bool
	IsBaby                 bool
	IsDead                 bool
	IsCryopodded           bool
	Generation             int
	AncestorIDs            []DinoID
	MaturationPercent      float64
	BabyStage              BabyStage
	StatusComponentUUID    *uuid.UUID
	InventoryUUID          *uuid.UUID
	TamedName              string
	IsNeutered             bool
	ColorSetIndices        [6]int
	ColorSetNames          [6]string
	UploadedFromServerName string
	Stats                  *DinoStats
	Owner                  DinoOwner
	GeneTraits             []string
	ParsedGeneTraits       []GeneTrait
	Location               *ActorTransform
}

type GeneTrait struct {
	Raw   string
	Name  string
	Level int
}

type DinoID struct {
	ID1 uint32
	ID2 uint32
}

func DinoFromObject(object *GameObject, location *ActorTransform) Dino {
	dino := Dino{Object: object, Location: location}
	if object == nil {
		return dino
	}
	properties := arkproperty.Container{Properties: object.Properties}
	dino.UUID = object.UUID
	dino.Blueprint = object.Blueprint
	dino.ID1 = uint32Value(properties, "DinoID1")
	dino.ID2 = uint32Value(properties, "DinoID2")
	dino.IsFemale = boolValue(properties, "bIsFemale")
	dino.IsDead = boolValue(properties, "bIsDead")
	dino.IsBaby = boolValue(properties, "bIsBaby")
	if dino.IsBaby {
		dino.MaturationPercent = float64Value(properties, "BabyAge") * 100
		dino.BabyStage = babyStageForPercent(dino.MaturationPercent)
	}
	_, dino.IsTamed = properties.Value("TamedTimeStamp")
	if dino.IsTamed {
		dino.Generation = max(arrayLength(properties, "DinoAncestors"), arrayLength(properties, "DinoAncestorsMale")) + 1
	}
	dino.AncestorIDs = ancestorIDs(properties)
	dino.StatusComponentUUID = objectReferenceUUID(properties, "MyCharacterStatusComponent")
	dino.InventoryUUID = objectReferenceUUID(properties, "MyInventoryComponent")
	dino.TamedName = stringValue(properties, "TamedName")
	dino.IsNeutered = boolValue(properties, "bNeutered")
	dino.ColorSetIndices = colorSetIndices(properties)
	dino.ColorSetNames = colorSetNames(properties)
	dino.UploadedFromServerName = uploadedFromServerName(properties)
	dino.Owner = DinoOwnerFromContainer(properties)
	dino.GeneTraits = stringArrayValue(properties, "GeneTraits")
	dino.ParsedGeneTraits = parseGeneTraits(dino.GeneTraits)
	if dino.Location == nil {
		dino.Location = actorTransformValue(properties, "SavedBaseWorldLocation")
	}
	return dino
}

func DinoFromObjectWithStatus(object *GameObject, statusObject *GameObject, location *ActorTransform) Dino {
	dino := DinoFromObject(object, location)
	if statusObject != nil {
		stats := DinoStatsFromObject(statusObject)
		dino.Stats = &stats
	}
	return dino
}

func (d Dino) ShortName() string {
	if d.Object != nil {
		return d.Object.ShortName()
	}
	return ShortNameFromBlueprint(d.Blueprint)
}

func (d Dino) IsWildTamed() bool {
	if !d.IsTamed {
		return false
	}
	if d.Object != nil {
		_, hasFemaleAncestors := (arkproperty.Container{Properties: d.Object.Properties}).Value("DinoAncestors")
		return !hasFemaleAncestors
	}
	return len(d.AncestorIDs) == 0
}

func colorSetIndices(properties arkproperty.Container) [6]int {
	var values [6]int
	for i := range values {
		value, ok := properties.PositionedValue("ColorSetIndices", int32(i))
		if !ok {
			continue
		}
		switch v := value.(type) {
		case int8:
			values[i] = int(v)
		case int32:
			values[i] = int(v)
		case int:
			values[i] = v
		}
	}
	return values
}

func colorSetNames(properties arkproperty.Container) [6]string {
	var values [6]string
	for i := range values {
		values[i] = "None"
		value, ok := properties.PositionedValue("ColorSetNames", int32(i))
		if !ok {
			continue
		}
		if name, ok := value.(string); ok && name != "" {
			values[i] = name
		}
	}
	return values
}

func uploadedFromServerName(properties arkproperty.Container) string {
	return strings.TrimPrefix(stringValue(properties, "UploadedFromServerName"), "\n")
}

func parseGeneTraits(values []string) []GeneTrait {
	if len(values) == 0 {
		return nil
	}
	out := make([]GeneTrait, 0, len(values))
	for _, raw := range values {
		out = append(out, parseGeneTrait(raw))
	}
	return out
}

func parseGeneTrait(raw string) GeneTrait {
	trait := GeneTrait{Raw: raw, Name: raw}
	open := strings.Index(raw, "[")
	if open < 0 || !strings.HasSuffix(raw, "]") {
		return trait
	}
	trait.Name = raw[:open]
	level, err := strconv.Atoi(raw[open+1 : len(raw)-1])
	if err == nil {
		trait.Level = level
	}
	return trait
}

func babyStageForPercent(percent float64) BabyStage {
	if percent < 10 {
		return BabyStageBaby
	}
	if percent < 50 {
		return BabyStageJuvenile
	}
	return BabyStageAdolescent
}

func uint32Value(properties arkproperty.Container, name string) uint32 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case uint32:
		return v
	case int32:
		return uint32(v)
	case int:
		return uint32(v)
	default:
		return 0
	}
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

func arrayLength(properties arkproperty.Container, name string) int {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	array, ok := value.(arkproperty.Array)
	if !ok {
		return 0
	}
	return len(array.Values)
}

func objectReferenceUUID(properties arkproperty.Container, name string) *uuid.UUID {
	value, ok := properties.Value(name)
	if !ok {
		return nil
	}
	ref, ok := value.(arkproperty.ObjectReference)
	if !ok || ref.Type != arkproperty.ObjectReferenceUUID {
		return nil
	}
	id, ok := uuidFromReferenceValue(ref.Value)
	if !ok {
		return nil
	}
	return &id
}

func actorTransformValue(properties arkproperty.Container, name string) *ActorTransform {
	value, ok := properties.Value(name)
	if !ok {
		return nil
	}
	switch v := value.(type) {
	case ActorTransform:
		return &v
	case *ActorTransform:
		return v
	default:
		return nil
	}
}
