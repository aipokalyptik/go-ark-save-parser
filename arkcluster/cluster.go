package arkcluster

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

type File struct {
	ID   string
	Path string
}

type Data struct {
	ID      string
	Path    string
	Archive *arkarchive.Archive
	Items   []Item
	Dinos   []Dino
}

type Item struct {
	Index      int
	Version    float64
	UploadTime float64
	Blueprint  string
	Quantity   int32
	Properties arkproperty.Container
}

type Dino struct {
	Index      int
	Version    float64
	UploadTime float64
	RawSize    int
	Archive    *arkarchive.Archive
	Properties arkproperty.Container
}

func Discover(dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isClusterFileName(name) {
			continue
		}
		files = append(files, File{ID: name, Path: filepath.Join(dir, name)})
	}
	sort.Slice(files, func(i int, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func Open(path string) (*Data, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	archive, err := arkarchive.Parse(raw, arkarchive.Options{FromStore: false, Format: arkarchive.FormatAuto})
	if err != nil {
		return nil, err
	}
	data := &Data{ID: filepath.Base(path), Path: path, Archive: archive}
	if err := data.parsePayload(); err != nil {
		return nil, err
	}
	return data, nil
}

func OpenDirectory(dir string) ([]*Data, error) {
	files, err := Discover(dir)
	if err != nil {
		return nil, err
	}
	out := make([]*Data, 0, len(files))
	for _, file := range files {
		data, err := Open(file.Path)
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return out, nil
}

func (d *Data) parsePayload() error {
	payload, ok, err := clusterPayload(d.Archive)
	if err != nil || !ok {
		return err
	}
	if raw, ok := payload.Value("ArkItems"); ok {
		items, err := parseItems(raw)
		if err != nil {
			return err
		}
		d.Items = items
	}
	if raw, ok := payload.Value("ArkTamedDinosData"); ok {
		dinos, err := parseDinos(raw)
		if err != nil {
			return err
		}
		d.Dinos = dinos
	}
	return nil
}

func clusterPayload(archive *arkarchive.Archive) (arkproperty.Container, bool, error) {
	if archive == nil {
		return arkproperty.Container{}, false, nil
	}
	for _, object := range archive.Objects {
		if object.ClassName != "/Script/ShooterGame.ArkCloudInventoryData" && object.ClassName != "/Script/ShooterGame.PrimalLocalProfile" {
			continue
		}
		container := arkproperty.Container{Properties: object.Properties}
		if object.ClassName == "/Script/ShooterGame.PrimalLocalProfile" {
			raw, ok := container.Value("MyArkData")
			if !ok {
				return arkproperty.Container{}, false, nil
			}
			return payloadContainer(raw)
		}
		if raw, ok := container.Value("MyArkData"); ok {
			return payloadContainer(raw)
		}
		return container, true, nil
	}
	return arkproperty.Container{}, false, nil
}

func payloadContainer(raw any) (arkproperty.Container, bool, error) {
	switch value := raw.(type) {
	case arkproperty.Container:
		return value, true, nil
	case arkproperty.UnknownStruct:
		reader := arkbinary.NewReader(value.Raw, nil)
		props, err := arkproperty.ParseAllPartial(reader, -1)
		if err != nil && len(props) == 0 {
			return arkproperty.Container{}, false, err
		}
		return arkproperty.Container{Properties: props}, true, nil
	default:
		return arkproperty.Container{}, false, fmt.Errorf("cluster payload has type %T, want property container", raw)
	}
}

func parseItems(raw any) ([]Item, error) {
	values := arrayValues(raw)
	items := make([]Item, 0, len(values))
	for i, value := range values {
		container, ok := value.(arkproperty.Container)
		if !ok {
			return nil, fmt.Errorf("cluster item %d has type %T, want property container", i, value)
		}
		item := Item{Index: i, Properties: container}
		item.Version = numberValue(container, "Version")
		item.UploadTime = numberValue(container, "UploadTime")
		itemProperties := itemPayloadProperties(container)
		item.Blueprint = blueprintValue(itemProperties, "ItemArchetype")
		item.Quantity = int32Value(itemProperties, "ItemQuantity")
		items = append(items, item)
	}
	return items, nil
}

func itemPayloadProperties(container arkproperty.Container) arkproperty.Container {
	if raw, ok := container.Value("ArkTributeItem"); ok {
		if nested, ok := raw.(arkproperty.Container); ok {
			return nested
		}
	}
	return container
}

func parseDinos(raw any) ([]Dino, error) {
	values := arrayValues(raw)
	dinos := make([]Dino, 0, len(values))
	for i, value := range values {
		container, ok := value.(arkproperty.Container)
		if !ok {
			return nil, fmt.Errorf("cluster dino %d has type %T, want property container", i, value)
		}
		dino := Dino{Index: i, Properties: container}
		dino.Version = numberValue(container, "Version")
		dino.UploadTime = numberValue(container, "UploadTime")
		if rawData, ok := byteSliceValue(container, "DinoData"); ok {
			dino.RawSize = len(rawData)
			archive, err := arkarchive.Parse(rawData, arkarchive.Options{FromStore: false, ClusterDino: true, Format: arkarchive.FormatClusterDino})
			if err == nil {
				dino.Archive = archive
			}
		}
		dinos = append(dinos, dino)
	}
	return dinos, nil
}

func arrayValues(raw any) []any {
	switch value := raw.(type) {
	case arkproperty.Array:
		return value.Values
	case []any:
		return value
	default:
		return nil
	}
}

func numberValue(container arkproperty.Container, name string) float64 {
	raw, ok := container.Value(name)
	if !ok {
		return 0
	}
	switch value := raw.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int32:
		return float64(value)
	case uint32:
		return float64(value)
	case int64:
		return float64(value)
	case uint64:
		return float64(value)
	case int:
		return float64(value)
	default:
		return 0
	}
}

func int32Value(container arkproperty.Container, name string) int32 {
	raw, ok := container.Value(name)
	if !ok {
		return 0
	}
	switch value := raw.(type) {
	case int32:
		return value
	case uint32:
		return int32(value)
	case float64:
		return int32(value)
	case float32:
		return int32(value)
	default:
		return 0
	}
}

func blueprintValue(container arkproperty.Container, name string) string {
	raw, ok := container.Value(name)
	if !ok {
		return ""
	}
	switch value := raw.(type) {
	case arkproperty.ObjectReference:
		if text, ok := value.Value.(string); ok {
			return trimBlueprintPrefix(text)
		}
	case string:
		return trimBlueprintPrefix(value)
	}
	return ""
}

func trimBlueprintPrefix(value string) string {
	fields := strings.Fields(value)
	if len(fields) >= 2 && (fields[0] == "BlueprintGeneratedClass" || fields[0] == "Class") {
		return fields[1]
	}
	return value
}

func byteSliceValue(container arkproperty.Container, name string) ([]byte, bool) {
	raw, ok := container.Value(name)
	if !ok {
		return nil, false
	}
	switch value := raw.(type) {
	case []byte:
		return value, true
	case arkproperty.Array:
		if value.ElementType != arkproperty.TypeByte {
			return nil, false
		}
		out := make([]byte, 0, len(value.Values))
		for _, item := range value.Values {
			switch b := item.(type) {
			case byte:
				out = append(out, b)
			default:
				return nil, false
			}
		}
		return out, true
	default:
		return nil, false
	}
}

func isClusterFileName(name string) bool {
	if name == "" || strings.HasPrefix(name, ".") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case "":
		return isExtensionlessClusterID(name)
	default:
		return false
	}
}

func isExtensionlessClusterID(name string) bool {
	if strings.HasPrefix(name, "EOS_") && len(name) > len("EOS_") {
		for _, r := range name[len("EOS_"):] {
			if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-') {
				return false
			}
		}
		return true
	}
	if len(name) != 32 {
		return false
	}
	for _, r := range name {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
