package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

func CryopodPayloadsFromObject(object *GameObject) [][]byte {
	if object == nil {
		return nil
	}
	value, ok := object.Value("CustomItemDatas")
	if !ok {
		return nil
	}
	customDatas, ok := value.(arkproperty.Array)
	if !ok || customDatas.ElementType != arkproperty.TypeStruct {
		return nil
	}
	var payloads [][]byte
	for _, customDataValue := range customDatas.Values {
		customData, ok := customDataValue.(arkproperty.Container)
		if !ok {
			continue
		}
		payloads = append(payloads, byteArraysFromCustomItemData(customData)...)
	}
	return payloads
}

func DinoFromCryopodObject(object *GameObject, maxInflatedBytes int64) (Dino, bool, error) {
	payloads := CryopodPayloadsFromObject(object)
	if len(payloads) == 0 {
		return Dino{}, false, nil
	}
	archive, err := arkarchive.ParseEmbeddedCryopodPayload(payloads[0], maxInflatedBytes)
	if err != nil {
		return Dino{}, false, err
	}
	if archive == nil || len(archive.Objects) < 2 {
		return Dino{}, false, nil
	}
	dinoArchiveObject, statusArchiveObject, ok := embeddedDinoAndStatusObjects(archive.Objects)
	if !ok {
		return Dino{}, false, nil
	}
	dinoObject := gameObjectFromArchiveObject(dinoArchiveObject)
	statusObject := gameObjectFromArchiveObject(statusArchiveObject)
	location := &ActorTransform{InCryopod: true}
	dino := DinoFromObjectWithStatus(dinoObject, statusObject, location)
	dino.IsCryopodded = true
	if dino.Location == nil {
		dino.Location = location
	} else {
		dino.Location.InCryopod = true
	}
	return dino, true, nil
}

func embeddedDinoAndStatusObjects(objects []arkarchive.Object) (arkarchive.Object, arkarchive.Object, bool) {
	dinoIndex := -1
	statusIndex := -1
	for i, object := range objects {
		container := arkproperty.Container{Properties: object.Properties}
		if dinoIndex < 0 && isEmbeddedDinoObject(container) {
			dinoIndex = i
			continue
		}
		if statusIndex < 0 && isEmbeddedStatusObject(container) {
			statusIndex = i
		}
	}
	if dinoIndex < 0 || statusIndex < 0 || dinoIndex == statusIndex {
		return arkarchive.Object{}, arkarchive.Object{}, false
	}
	return objects[dinoIndex], objects[statusIndex], true
}

func isEmbeddedDinoObject(properties arkproperty.Container) bool {
	_, hasID1 := properties.Value("DinoID1")
	_, hasID2 := properties.Value("DinoID2")
	return hasID1 || hasID2
}

func isEmbeddedStatusObject(properties arkproperty.Container) bool {
	for _, name := range []string{
		"BaseCharacterLevel",
		"NumberOfLevelUpPointsApplied",
		"NumberOfLevelUpPointsAppliedTamed",
		"NumberOfMutationsAppliedTamed",
		"CurrentStatusValues",
	} {
		if _, ok := properties.Value(name); ok {
			return true
		}
	}
	return false
}

func gameObjectFromArchiveObject(object arkarchive.Object) *GameObject {
	return &GameObject{
		UUID:       object.UUID,
		Blueprint:  object.ClassName,
		Names:      object.Names,
		Properties: object.Properties,
	}
}

func byteArraysFromCustomItemData(customData arkproperty.Container) [][]byte {
	value, ok := customData.Value("CustomDataBytes")
	if !ok {
		return nil
	}
	customDataBytes, ok := value.(arkproperty.Container)
	if !ok {
		return nil
	}
	value, ok = customDataBytes.Value("ByteArrays")
	if !ok {
		return nil
	}
	byteArrays, ok := value.(arkproperty.Array)
	if !ok || byteArrays.ElementType != arkproperty.TypeStruct {
		return nil
	}
	var payloads [][]byte
	for _, byteArrayValue := range byteArrays.Values {
		byteArray, ok := byteArrayValue.(arkproperty.Container)
		if !ok {
			continue
		}
		payload, ok := bytesFromCustomItemByteArray(byteArray)
		if ok && len(payload) > 0 {
			payloads = append(payloads, payload)
		}
	}
	return payloads
}

func bytesFromCustomItemByteArray(byteArray arkproperty.Container) ([]byte, bool) {
	value, ok := byteArray.Value("Bytes")
	if !ok {
		return nil, false
	}
	bytesArray, ok := value.(arkproperty.Array)
	if !ok || bytesArray.ElementType != arkproperty.TypeByte {
		return nil, false
	}
	payload := make([]byte, 0, len(bytesArray.Values))
	for _, byteValue := range bytesArray.Values {
		switch v := byteValue.(type) {
		case byte:
			payload = append(payload, v)
		}
	}
	return payload, true
}
