package arkproperty

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
)

type Type string

const (
	TypeArray  Type = "Array"
	TypeBool   Type = "Boolean"
	TypeDouble Type = "Double"
	TypeFloat  Type = "Float"
	TypeInt    Type = "Int"
	TypeMap    Type = "Map"
	TypeObject Type = "Object"
	TypeSet    Type = "Set"
	TypeString Type = "String"
	TypeStruct Type = "Struct"
	TypeUInt32 Type = "UInt32"
)

type ObjectReferenceType int

const (
	ObjectReferenceUUID ObjectReferenceType = iota
	ObjectReferencePath
	ObjectReferenceID
	ObjectReferenceUnknown
)

type ObjectReference struct {
	Type  ObjectReferenceType
	Value any
}

type Array struct {
	ElementType Type
	Values      []any
}

type MapEntry struct {
	Key   any
	Value any
}

type Map struct {
	KeyType   Type
	ValueType Type
	Entries   []MapEntry
}

type Set struct {
	ElementType Type
	Values      []any
}

type Container struct {
	Properties []Property
}

type UnknownStruct struct {
	TypeName string
	Raw      []byte
}

func (c Container) Value(name string) (any, bool) {
	for _, prop := range c.Properties {
		if prop.Name == name {
			return prop.Value, true
		}
	}
	return nil, false
}

type Property struct {
	Name         string
	Type         Type
	Value        any
	Position     int32
	NameOffset   int
	ValueOffset  int
	DataSize     int32
	UnknownByte  byte
	EncodedBytes []byte
}

func ParseAll(r *arkbinary.Reader, end int) ([]Property, error) {
	var props []Property
	for r.HasMore() && (end < 0 || r.Position() < end) {
		prop, err := ParseOne(r, end)
		if err != nil {
			return nil, err
		}
		if prop == nil {
			break
		}
		props = append(props, *prop)
	}
	return props, nil
}

func ParseOne(r *arkbinary.Reader, structEnd int) (*Property, error) {
	nameOffset := r.Position()
	key, err := r.ReadName("")
	if err != nil {
		return nil, err
	}
	if key == "" || key == "None" {
		if shouldSkipTrailingNoneZeros(r, structEnd) {
			if err := r.Skip(4); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	typeName, err := r.ReadName("")
	if err != nil {
		return nil, err
	}
	dataSize, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}
	position, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}

	prop := &Property{
		Name:       key,
		Position:   position,
		NameOffset: nameOffset,
		DataSize:   dataSize,
	}

	switch typeName {
	case "BoolProperty":
		prop.Type = TypeBool
		prop.ValueOffset = r.Position()
		value, err := r.ReadBool()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "IntProperty":
		prop.Type = TypeInt
		unknown, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		prop.UnknownByte = unknown
		prop.ValueOffset = r.Position()
		value, err := r.ReadInt32()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "DoubleProperty":
		prop.Type = TypeDouble
		unknown, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		prop.UnknownByte = unknown
		prop.ValueOffset = r.Position()
		value, err := r.ReadFloat64()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "FloatProperty":
		prop.Type = TypeFloat
		isPositioned, err := r.ReadBool()
		if err != nil {
			return nil, err
		}
		if isPositioned {
			position, err := r.ReadInt32()
			if err != nil {
				return nil, err
			}
			prop.Position = position
		}
		prop.ValueOffset = r.Position()
		value, err := r.ReadFloat32()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "StrProperty":
		prop.Type = TypeString
		isPositioned, err := r.ReadBool()
		if err != nil {
			return nil, err
		}
		if isPositioned {
			position, err := r.ReadInt32()
			if err != nil {
				return nil, err
			}
			prop.Position = position
		}
		prop.ValueOffset = r.Position()
		value, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		if value != nil {
			prop.Value = *value
		} else {
			prop.Value = ""
		}
	case "ObjectProperty":
		prop.Type = TypeObject
		isPositioned, err := r.ReadBool()
		if err != nil {
			return nil, err
		}
		if isPositioned {
			position, err := r.ReadInt32()
			if err != nil {
				return nil, err
			}
			prop.Position = position
		}
		prop.ValueOffset = r.Position()
		ref, err := readObjectReference(r)
		if err != nil {
			return nil, err
		}
		prop.Value = ref
	case "ArrayProperty":
		prop.Type = TypeArray
		if err := r.SetPosition(r.Position() - 4); err != nil {
			return nil, err
		}
		array, err := readArray(r)
		if err != nil {
			return nil, err
		}
		prop.ValueOffset = r.Position()
		prop.Value = array
	case "MapProperty":
		prop.Type = TypeMap
		if err := r.SetPosition(r.Position() - 4); err != nil {
			return nil, err
		}
		value, err := readMap(r)
		if err != nil {
			return nil, err
		}
		prop.ValueOffset = r.Position()
		prop.Value = value
	case "SetProperty":
		prop.Type = TypeSet
		if err := r.SetPosition(r.Position() - 4); err != nil {
			return nil, err
		}
		value, err := readSet(r)
		if err != nil {
			return nil, err
		}
		prop.ValueOffset = r.Position()
		prop.Value = value
	case "StructProperty":
		prop.Type = TypeStruct
		if err := r.SetPosition(r.Position() - 8); err != nil {
			return nil, err
		}
		value, structType, declaredSize, err := readStruct(r)
		if err != nil {
			return nil, err
		}
		prop.ValueOffset = r.Position()
		prop.DataSize = int32(declaredSize)
		prop.Value = value
		_ = structType
	case "UInt32Property":
		prop.Type = TypeUInt32
		unknown, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		prop.UnknownByte = unknown
		prop.ValueOffset = r.Position()
		value, err := r.ReadUInt32()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	default:
		return nil, fmt.Errorf("unsupported property type %q for %q", typeName, key)
	}

	return prop, nil
}

func readMap(r *arkbinary.Reader) (Map, error) {
	keyTypeName, err := r.ReadName("")
	if err != nil {
		return Map{}, err
	}
	keyType, err := typeFromPropertyName(keyTypeName)
	if err != nil {
		return Map{}, err
	}
	if _, err := r.ReadUInt32(); err != nil {
		return Map{}, err
	}
	valueTypeName, err := r.ReadName("")
	if err != nil {
		return Map{}, err
	}
	valueType, err := typeFromPropertyName(valueTypeName)
	if err != nil {
		return Map{}, err
	}
	structNames, err := r.ReadInt32()
	if err != nil {
		return Map{}, err
	}
	if structNames > 0 {
		if _, err := r.ReadName(""); err != nil {
			return Map{}, err
		}
	}
	dataSize, err := readInlineStructHeader(r, uint32(structNames))
	if err != nil {
		return Map{}, err
	}
	bodyStart := r.Position()
	if _, err := r.ReadUInt32(); err != nil {
		return Map{}, err
	}
	count, err := r.ReadUInt32()
	if err != nil {
		return Map{}, err
	}
	entries := make([]MapEntry, 0, count)
	for i := uint32(0); i < count; i++ {
		key, err := readValue(keyType, r)
		if err != nil {
			return Map{}, err
		}
		value, err := readValue(valueType, r)
		if err != nil {
			return Map{}, err
		}
		entries = append(entries, MapEntry{Key: key, Value: value})
	}
	if err := alignDeclaredBody(r, bodyStart, dataSize); err != nil {
		return Map{}, err
	}
	return Map{KeyType: keyType, ValueType: valueType, Entries: entries}, nil
}

func readSet(r *arkbinary.Reader) (Set, error) {
	valueTypeName, err := r.ReadName("")
	if err != nil {
		return Set{}, err
	}
	elementType, err := typeFromPropertyName(valueTypeName)
	if err != nil {
		return Set{}, err
	}
	zero, err := r.ReadUInt32()
	if err != nil {
		return Set{}, err
	}
	if zero != 0 {
		return Set{}, fmt.Errorf("invalid set header zero %#x", zero)
	}
	dataSize, err := r.ReadInt32()
	if err != nil {
		return Set{}, err
	}
	endByte, err := r.ReadByte()
	if err != nil {
		return Set{}, err
	}
	if endByte != 0 {
		return Set{}, fmt.Errorf("invalid set end byte %#x", endByte)
	}
	bodyStart := r.Position()
	if _, err := r.ReadUInt32(); err != nil {
		return Set{}, err
	}
	count, err := r.ReadInt32()
	if err != nil {
		return Set{}, err
	}
	if count < 0 {
		return Set{}, fmt.Errorf("negative set count %d", count)
	}
	values := make([]any, 0, count)
	for i := int32(0); i < count; i++ {
		value, err := readValue(elementType, r)
		if err != nil {
			return Set{}, err
		}
		values = append(values, value)
	}
	if err := alignDeclaredBody(r, bodyStart, uint32(dataSize)); err != nil {
		return Set{}, err
	}
	return Set{ElementType: elementType, Values: values}, nil
}

func readStruct(r *arkbinary.Reader) (any, string, uint32, error) {
	nrNames, err := r.ReadUInt32()
	if err != nil {
		return nil, "", 0, err
	}
	structType, err := r.ReadName("")
	if err != nil {
		return nil, "", 0, err
	}
	if nrNames != 0 {
		marker, err := r.ReadUInt32()
		if err != nil {
			return nil, "", 0, err
		}
		if marker != 1 {
			return nil, "", 0, fmt.Errorf("invalid struct header marker %#x", marker)
		}
	}
	for i := uint32(0); i < nrNames; i++ {
		if _, err := r.ReadName(""); err != nil {
			return nil, "", 0, err
		}
		zero, err := r.ReadUInt32()
		if err != nil {
			return nil, "", 0, err
		}
		if zero != 0 {
			return nil, "", 0, fmt.Errorf("invalid struct name terminator %#x", zero)
		}
	}
	dataSize, err := r.ReadUInt32()
	if err != nil {
		return nil, "", 0, err
	}
	sizeByte, err := r.ReadByte()
	if err != nil {
		return nil, "", 0, err
	}
	if sizeByte != 0 && sizeByte != 8 {
		if _, err := r.ReadUInt32(); err != nil {
			return nil, "", 0, err
		}
	}
	bodyStart := r.Position()
	bodyEnd := bodyStart + int(dataSize)
	props, err := ParseAll(r, bodyEnd)
	if err != nil {
		if err := r.SetPosition(bodyStart); err != nil {
			return nil, "", 0, err
		}
		raw, readErr := r.ReadBytes(int(dataSize))
		if readErr != nil {
			return nil, "", 0, readErr
		}
		return UnknownStruct{TypeName: structType, Raw: raw}, structType, dataSize, nil
	}
	if r.Position() < bodyEnd {
		if err := r.SetPosition(bodyEnd); err != nil {
			return nil, "", 0, err
		}
	}
	return Container{Properties: props}, structType, dataSize, nil
}

func readInlineStructHeader(r *arkbinary.Reader, nrNames uint32) (uint32, error) {
	if nrNames != 0 {
		marker, err := r.ReadUInt32()
		if err != nil {
			return 0, err
		}
		if marker != 1 {
			return 0, fmt.Errorf("invalid inline struct header marker %#x", marker)
		}
	}
	for i := uint32(0); i < nrNames; i++ {
		if _, err := r.ReadName(""); err != nil {
			return 0, err
		}
		zero, err := r.ReadUInt32()
		if err != nil {
			return 0, err
		}
		if zero != 0 {
			return 0, fmt.Errorf("invalid inline struct name terminator %#x", zero)
		}
	}
	dataSize, err := r.ReadUInt32()
	if err != nil {
		return 0, err
	}
	sizeByte, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	if sizeByte != 0 && sizeByte != 8 {
		if _, err := r.ReadUInt32(); err != nil {
			return 0, err
		}
	}
	return dataSize, nil
}

func alignDeclaredBody(r *arkbinary.Reader, bodyStart int, dataSize uint32) error {
	bodyEnd := bodyStart + int(dataSize)
	if r.Position() < bodyEnd {
		return r.SetPosition(bodyEnd)
	}
	return nil
}

func readObjectReference(r *arkbinary.Reader) (ObjectReference, error) {
	refType, err := r.ReadInt16()
	if err != nil {
		return ObjectReference{}, err
	}
	switch refType {
	case 0:
		id, err := r.ReadUUID()
		if err != nil {
			return ObjectReference{}, err
		}
		return ObjectReference{Type: ObjectReferenceUUID, Value: id.String()}, nil
	case 1, 256:
		name, err := r.ReadName("")
		if err != nil {
			return ObjectReference{}, err
		}
		return ObjectReference{Type: ObjectReferencePath, Value: name}, nil
	case 4:
		id, err := r.ReadInt32()
		if err != nil {
			return ObjectReference{}, err
		}
		return ObjectReference{Type: ObjectReferenceID, Value: id}, nil
	default:
		return ObjectReference{}, fmt.Errorf("unknown object reference type %d", refType)
	}
}

func readArray(r *arkbinary.Reader) (Array, error) {
	arrayTypeName, err := r.ReadName("")
	if err != nil {
		return Array{}, err
	}
	if _, err := r.ReadInt32(); err != nil {
		return Array{}, err
	}
	if _, err := r.ReadUInt32(); err != nil {
		return Array{}, err
	}
	if _, err := r.ReadByte(); err != nil {
		return Array{}, err
	}
	length, err := r.ReadUInt32()
	if err != nil {
		return Array{}, err
	}
	elementType, err := typeFromPropertyName(arrayTypeName)
	if err != nil {
		return Array{}, err
	}
	values := make([]any, 0, length)
	for i := uint32(0); i < length; i++ {
		value, err := readValue(elementType, r)
		if err != nil {
			return Array{}, err
		}
		values = append(values, value)
	}
	return Array{ElementType: elementType, Values: values}, nil
}

func typeFromPropertyName(name string) (Type, error) {
	switch name {
	case "BoolProperty":
		return TypeBool, nil
	case "DoubleProperty":
		return TypeDouble, nil
	case "FloatProperty":
		return TypeFloat, nil
	case "IntProperty":
		return TypeInt, nil
	case "ObjectProperty":
		return TypeObject, nil
	case "StrProperty":
		return TypeString, nil
	case "UInt32Property":
		return TypeUInt32, nil
	default:
		return "", fmt.Errorf("unsupported value type %q", name)
	}
}

func readValue(t Type, r *arkbinary.Reader) (any, error) {
	switch t {
	case TypeBool:
		return r.ReadBool()
	case TypeDouble:
		return r.ReadFloat64()
	case TypeFloat:
		return r.ReadFloat32()
	case TypeInt:
		return r.ReadInt32()
	case TypeObject:
		return readObjectReference(r)
	case TypeString:
		value, err := r.ReadString()
		if err != nil || value == nil {
			return "", err
		}
		return *value, nil
	case TypeUInt32:
		return r.ReadUInt32()
	default:
		return nil, fmt.Errorf("unsupported value type %q", t)
	}
}

func shouldSkipTrailingNoneZeros(r *arkbinary.Reader, structEnd int) bool {
	if structEnd >= 0 && r.Position() >= structEnd {
		return false
	}
	if r.Size()-r.Position() < 4 {
		return false
	}
	value, err := r.PeekUInt32()
	return err == nil && value == 0
}
