package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/google/uuid"
)

func ClusterDinoFromUpload(dino arkcluster.Dino, classNames []string) ClusterDino {
	objectCount := 0
	if dino.Archive != nil {
		objectCount = len(dino.Archive.Objects)
	}
	out := ClusterDino{
		Index:                        dino.Index,
		Version:                      dino.Version,
		UploadTime:                   dino.UploadTime,
		RawSize:                      dino.RawSize,
		ObjectCount:                  objectCount,
		ParsedArchive:                dino.Archive != nil,
		ClassNames:                   append([]string(nil), classNames...),
		StatusComponentClassNames:    clusterClassNamesContainingAny(classNames, "CharacterStatusComponent", "DinoCharacterStatus", "CharacterStatus"),
		AIControllerClassNames:       clusterClassNamesContaining(classNames, "AIController"),
		InventoryComponentClassNames: clusterClassNamesContaining(classNames, "InventoryComponent"),
		ParseError:                   dino.ParseError,
		Properties:                   dino.Properties,
	}
	if parsed, ok := clusterEmbeddedDino(dino.Archive); ok {
		out.DinoID1 = parsed.ID1
		out.DinoID2 = parsed.ID2
		out.TamedName = parsed.TamedName
		out.IsTamed = parsed.IsTamed
		out.IsFemale = parsed.IsFemale
		out.IsBaby = parsed.IsBaby
		out.IsDead = parsed.IsDead
		if parsed.Stats != nil {
			out.HasStats = true
			out.BaseLevel = parsed.Stats.BaseLevel
			out.CurrentLevel = parsed.Stats.CurrentLevel
		}
	}
	return out
}

func (d ClusterDino) Parsed() bool {
	return d.ParsedArchive && d.ParseError == ""
}

func (d ClusterDino) HasParseError() bool {
	return d.ParseError != ""
}

func (d ClusterDino) SupportedVersion() bool {
	return d.Version == 7
}

func (d ClusterDino) UnsupportedVersion() bool {
	return d.Version != 0 && !d.SupportedVersion()
}

func (d ClusterDino) ParseStatus() ClusterDinoParseStatus {
	switch {
	case d.HasParseError():
		return ClusterDinoParseStatusParseError
	case d.UnsupportedVersion():
		return ClusterDinoParseStatusUnsupportedVersion
	case d.ParsedArchive:
		return ClusterDinoParseStatusParsed
	default:
		return ClusterDinoParseStatusUnparsed
	}
}

func (d ClusterDino) PrimaryClassName() string {
	for _, className := range d.ClassNames {
		if !clusterDinoComponentClass(className) {
			return className
		}
	}
	if len(d.ClassNames) == 0 {
		return ""
	}
	return d.ClassNames[0]
}

func (d ClusterDino) ShortName() string {
	return ShortNameFromBlueprint(d.PrimaryClassName())
}

func clusterEmbeddedDino(archive *arkarchive.Archive) (Dino, bool) {
	if archive == nil {
		return Dino{}, false
	}
	var dinoObject *GameObject
	statusObjects := make(map[uuid.UUID]*GameObject)
	var firstStatus *GameObject
	for i := range archive.Objects {
		object := gameObjectFromArchiveObject(archive.Objects[i])
		if clusterDinoComponentClass(object.Blueprint) {
			statusObjects[object.UUID] = object
			if firstStatus == nil {
				firstStatus = object
			}
			continue
		}
		if dinoObject == nil {
			dinoObject = object
		}
	}
	if dinoObject == nil {
		return Dino{}, false
	}
	parsed := DinoFromObject(dinoObject, nil)
	var statusObject *GameObject
	if parsed.StatusComponentUUID != nil {
		statusObject = statusObjects[*parsed.StatusComponentUUID]
	}
	if statusObject == nil {
		statusObject = firstStatus
	}
	if statusObject != nil {
		parsed = DinoFromObjectWithStatus(dinoObject, statusObject, nil)
	}
	return parsed, true
}
