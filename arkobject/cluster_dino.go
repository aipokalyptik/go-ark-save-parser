package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkcluster"

func ClusterDinoFromUpload(dino arkcluster.Dino, classNames []string) ClusterDino {
	objectCount := 0
	if dino.Archive != nil {
		objectCount = len(dino.Archive.Objects)
	}
	return ClusterDino{
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
