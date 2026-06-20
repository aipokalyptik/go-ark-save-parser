package arkproperty

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
)

type Type string

const (
	TypeArray      Type = "Array"
	TypeByte       Type = "Byte"
	TypeBool       Type = "Boolean"
	TypeDouble     Type = "Double"
	TypeEnum       Type = "Enum"
	TypeFloat      Type = "Float"
	TypeInt        Type = "Int"
	TypeInt8       Type = "Int8"
	TypeInt16      Type = "Int16"
	TypeInt64      Type = "Int64"
	TypeMap        Type = "Map"
	TypeName       Type = "Name"
	TypeObject     Type = "Object"
	TypeSet        Type = "Set"
	TypeSoftObject Type = "SoftObject"
	TypeString     Type = "String"
	TypeStruct     Type = "Struct"
	TypeUInt32     Type = "UInt32"
	TypeUInt16     Type = "UInt16"
	TypeUInt64     Type = "UInt64"
)

type ObjectReferenceType int

const (
	ObjectReferenceUUID ObjectReferenceType = iota
	ObjectReferencePath
	ObjectReferencePathNoType
	ObjectReferenceID
	ObjectReferenceUnknown
)

type ObjectReference struct {
	Type  ObjectReferenceType
	Value any
}

type Array struct {
	ElementType Type
	StructType  string
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

type EnumValue struct {
	Name string
}

func (c Container) Value(name string) (any, bool) {
	for _, prop := range c.Properties {
		if prop.Name == name {
			return prop.Value, true
		}
	}
	return nil, false
}

func (c Container) PositionedValue(name string, position int32) (any, bool) {
	for _, prop := range c.Properties {
		if prop.Name == name && prop.Position == position {
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
	return parseAll(r, end, false)
}

func ParseAllPartial(r *arkbinary.Reader, end int) ([]Property, error) {
	return parseAll(r, end, true)
}

func parseAll(r *arkbinary.Reader, end int, keepPartial bool) ([]Property, error) {
	var props []Property
	for r.HasMore() && (end < 0 || r.Position() < end) {
		prop, err := ParseOne(r, end)
		if err != nil {
			if keepPartial && prop != nil {
				props = append(props, *prop)
			}
			if keepPartial {
				return props, err
			}
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
	case "ByteProperty":
		value, valueType, valueOffset, position, unknown, err := readByteProperty(r, dataSize, position)
		if err != nil {
			return nil, err
		}
		prop.Type = valueType
		prop.Position = position
		prop.UnknownByte = unknown
		prop.ValueOffset = valueOffset
		prop.Value = value
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
	case "Int8Property":
		prop.Type = TypeInt8
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
		value, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		prop.Value = int8(value)
	case "Int16Property":
		prop.Type = TypeInt16
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
		value, err := r.ReadInt16()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "Int64Property":
		prop.Type = TypeInt64
		unknown, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		prop.UnknownByte = unknown
		prop.ValueOffset = r.Position()
		value, err := r.ReadInt64()
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
		ref, err := readObjectReference(r, dataSize)
		if err != nil {
			return nil, err
		}
		prop.Value = ref
	case "NameProperty":
		prop.Type = TypeName
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
		value, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "SoftObjectProperty":
		prop.Type = TypeSoftObject
		unknown, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		prop.UnknownByte = unknown
		prop.ValueOffset = r.Position()
		value, err := readSoftObject(r)
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "ArrayProperty":
		prop.Type = TypeArray
		if err := r.SetPosition(r.Position() - 4); err != nil {
			return nil, err
		}
		array, err := readArray(r)
		prop.Value = array
		if err != nil {
			if len(array.Values) > 0 {
				if captureErr := captureEncodedBytes(r, prop); captureErr != nil {
					return prop, captureErr
				}
				return prop, err
			}
			return nil, err
		}
		prop.ValueOffset = r.Position()
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
		prop.ValueOffset = r.Position()
		prop.DataSize = int32(declaredSize)
		prop.Value = value
		_ = structType
		if err != nil {
			if captureErr := captureEncodedBytes(r, prop); captureErr != nil {
				return prop, captureErr
			}
			return prop, err
		}
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
	case "UInt16Property":
		prop.Type = TypeUInt16
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
		value, err := r.ReadUInt16()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	case "UInt64Property":
		prop.Type = TypeUInt64
		unknown, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		prop.UnknownByte = unknown
		prop.ValueOffset = r.Position()
		value, err := r.ReadUInt64()
		if err != nil {
			return nil, err
		}
		prop.Value = value
	default:
		return nil, fmt.Errorf("unsupported property type %q for %q", typeName, key)
	}

	if err := realignPrimitiveProperty(r, prop); err != nil {
		return nil, err
	}
	if err := captureEncodedBytes(r, prop); err != nil {
		return nil, err
	}
	return prop, nil
}

func captureEncodedBytes(r *arkbinary.Reader, prop *Property) error {
	if prop == nil {
		return nil
	}
	encoded, err := r.Slice(prop.NameOffset, r.Position())
	if err != nil {
		return err
	}
	prop.EncodedBytes = encoded
	return nil
}

func realignPrimitiveProperty(r *arkbinary.Reader, prop *Property) error {
	if prop.DataSize <= 0 || prop.ValueOffset == 0 {
		return nil
	}
	switch prop.Type {
	case TypeBool, TypeDouble, TypeFloat, TypeInt, TypeInt8, TypeInt16, TypeInt64, TypeName, TypeObject, TypeSoftObject, TypeString, TypeUInt16, TypeUInt32, TypeUInt64:
	default:
		return nil
	}
	target := prop.ValueOffset + int(prop.DataSize)
	switch current := r.Position(); {
	case current == target:
		return nil
	case current < target:
		return r.Skip(target - current)
	default:
		return fmt.Errorf("property %q read %d bytes past declared payload size %d", prop.Name, current-target, prop.DataSize)
	}
}

func readByteProperty(r *arkbinary.Reader, dataSize int32, position int32) (any, Type, int, int32, byte, error) {
	preReadPosition := r.Position()
	if dataSize == 0 {
		isPositioned, err := r.ReadBool()
		if err != nil {
			return nil, "", 0, 0, 0, err
		}
		if isPositioned {
			updated, err := r.ReadInt32()
			if err != nil {
				return nil, "", 0, 0, 0, err
			}
			position = updated
		} else {
			position = 0
		}
		valueOffset := r.Position()
		value, err := r.ReadByte()
		if err != nil {
			return nil, "", 0, 0, 0, err
		}
		return value, TypeByte, valueOffset, position, 0, nil
	}

	if err := r.SetPosition(preReadPosition - 4); err != nil {
		return nil, "", 0, 0, 0, err
	}
	if _, err := r.ReadName(""); err != nil {
		return nil, "", 0, 0, 0, err
	}
	if _, err := r.ReadInt32(); err != nil {
		return nil, "", 0, 0, 0, err
	}
	if _, err := r.ReadName(""); err != nil {
		return nil, "", 0, 0, 0, err
	}
	zero, err := r.ReadUInt32()
	if err != nil {
		return nil, "", 0, 0, 0, err
	}
	if zero != 0 {
		return nil, "", 0, 0, 0, fmt.Errorf("invalid enum zero %#x", zero)
	}
	unknown, err := r.ReadByte()
	if err != nil {
		return nil, "", 0, 0, 0, err
	}
	zero, err = r.ReadUInt32()
	if err != nil {
		return nil, "", 0, 0, 0, err
	}
	if zero != 0 {
		return nil, "", 0, 0, 0, fmt.Errorf("invalid enum terminator %#x", zero)
	}
	name, err := r.ReadName("")
	if err != nil {
		return nil, "", 0, 0, 0, err
	}
	return EnumValue{Name: name}, TypeEnum, r.Position(), position, unknown, nil
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
	bodyEnd := bodyStart + int(dataSize)
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
		value, err := readMapValue(valueType, r, bodyEnd)
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

func readMapValue(t Type, r *arkbinary.Reader, bodyEnd int) (any, error) {
	if t != TypeStruct {
		return readValue(t, r)
	}
	props, err := ParseAllPartial(r, bodyEnd)
	if err != nil {
		if len(props) > 0 {
			return Container{Properties: props}, err
		}
		return nil, err
	}
	return Container{Properties: props}, nil
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
	bodyStart = r.Position()
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
	props, err := ParseAllPartial(r, bodyEnd)
	if err != nil {
		if len(props) > 0 {
			if err := alignDeclaredBody(r, bodyStart, dataSize); err != nil {
				return nil, "", 0, err
			}
			return Container{Properties: props}, structType, dataSize, err
		}
		if err := r.SetPosition(bodyStart); err != nil {
			return nil, "", 0, err
		}
		raw, readErr := r.ReadBytes(int(dataSize))
		if readErr != nil {
			return nil, "", 0, readErr
		}
		return UnknownStruct{TypeName: structType, Raw: raw}, structType, dataSize, nil
	}
	if err := alignDeclaredBody(r, bodyStart, dataSize); err != nil {
		return nil, "", 0, err
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
	switch current := r.Position(); {
	case current < bodyEnd:
		return r.SetPosition(bodyEnd)
	case current > bodyEnd:
		return fmt.Errorf("read %d bytes past declared compound payload size %d", current-bodyEnd, dataSize)
	default:
		return nil
	}
}

func readObjectReference(r *arkbinary.Reader, dataSize int32) (ObjectReference, error) {
	if !r.HasNameTable() {
		refType, err := r.ReadInt32()
		if err != nil {
			return ObjectReference{}, err
		}
		switch refType {
		case -1:
			return ObjectReference{Type: ObjectReferenceUnknown, Value: nil}, nil
		case 0:
			id, err := r.ReadInt32()
			if err != nil {
				return ObjectReference{}, err
			}
			return ObjectReference{Type: ObjectReferenceID, Value: id}, nil
		case 1:
			if dataSize == 4 {
				return ObjectReference{Type: ObjectReferencePath, Value: "NONE"}, nil
			}
			value, err := r.ReadString()
			if err != nil {
				return ObjectReference{}, err
			}
			if value == nil {
				return ObjectReference{Type: ObjectReferencePath, Value: ""}, nil
			}
			return ObjectReference{Type: ObjectReferencePath, Value: *value}, nil
		default:
			if err := r.SetPosition(r.Position() - 4); err != nil {
				return ObjectReference{}, err
			}
			value, err := r.ReadString()
			if err != nil {
				return ObjectReference{}, err
			}
			if value == nil {
				return ObjectReference{Type: ObjectReferencePathNoType, Value: ""}, nil
			}
			return ObjectReference{Type: ObjectReferencePathNoType, Value: *value}, nil
		}
	}
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

func readSoftObject(r *arkbinary.Reader) ([]string, error) {
	names := make([]string, 0)
	for {
		next, err := r.PeekUInt32()
		if err != nil {
			return nil, err
		}
		if next == 0 {
			if _, err := r.ReadUInt32(); err != nil {
				return nil, err
			}
			return names, nil
		}
		name, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
}

func readArray(r *arkbinary.Reader) (Array, error) {
	arrayTypeName, err := r.ReadName("")
	if err != nil {
		return Array{}, err
	}
	nrNames, err := r.ReadInt32()
	if err != nil {
		return Array{}, err
	}
	if arrayTypeName == "StructProperty" {
		structType, err := r.ReadName("")
		if err != nil {
			return Array{}, err
		}
		dataSize, err := readInlineStructHeader(r, uint32(nrNames))
		if err != nil {
			return Array{}, err
		}
		bodyStart := r.Position()
		count, err := r.ReadUInt32()
		if err != nil {
			return Array{}, err
		}
		arrayEnd := bodyStart + int(dataSize)
		values := make([]any, 0, count)
		for i := uint32(0); i < count; i++ {
			props, err := ParseAllPartial(r, arrayEnd)
			if err != nil {
				if len(props) > 0 {
					values = append(values, Container{Properties: props})
					if r.Position() < arrayEnd {
						if seekErr := r.SetPosition(arrayEnd); seekErr != nil {
							return Array{}, seekErr
						}
					}
					return Array{ElementType: TypeStruct, StructType: structType, Values: values}, err
				}
				return Array{}, err
			}
			values = append(values, Container{Properties: props})
		}
		if err := alignDeclaredBody(r, bodyStart, dataSize); err != nil {
			return Array{}, err
		}
		return Array{ElementType: TypeStruct, StructType: structType, Values: values}, nil
	}
	dataSize, err := r.ReadUInt32()
	if err != nil {
		return Array{}, err
	}
	if _, err := r.ReadByte(); err != nil {
		return Array{}, err
	}
	length, err := r.ReadUInt32()
	if err != nil {
		return Array{}, err
	}
	bodyStart := r.Position()
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
	if err := alignArrayBody(r, bodyStart, dataSize); err != nil {
		return Array{}, err
	}
	return Array{ElementType: elementType, Values: values}, nil
}

func alignArrayBody(r *arkbinary.Reader, elementBodyStart int, dataSize uint32) error {
	if dataSize >= 4 && r.Position() == elementBodyStart-4+int(dataSize) {
		return nil
	}
	return alignDeclaredBody(r, elementBodyStart, dataSize)
}

func typeFromPropertyName(name string) (Type, error) {
	switch name {
	case "ByteProperty":
		return TypeByte, nil
	case "BoolProperty":
		return TypeBool, nil
	case "DoubleProperty":
		return TypeDouble, nil
	case "FloatProperty":
		return TypeFloat, nil
	case "IntProperty":
		return TypeInt, nil
	case "Int8Property":
		return TypeInt8, nil
	case "Int16Property":
		return TypeInt16, nil
	case "Int64Property":
		return TypeInt64, nil
	case "NameProperty":
		return TypeName, nil
	case "ObjectProperty":
		return TypeObject, nil
	case "SoftObjectProperty":
		return TypeSoftObject, nil
	case "StrProperty":
		return TypeString, nil
	case "StructProperty":
		return TypeStruct, nil
	case "UInt32Property":
		return TypeUInt32, nil
	case "UInt16Property":
		return TypeUInt16, nil
	case "UInt64Property":
		return TypeUInt64, nil
	default:
		return "", fmt.Errorf("unsupported value type %q", name)
	}
}

func readValue(t Type, r *arkbinary.Reader) (any, error) {
	switch t {
	case TypeByte:
		return r.ReadByte()
	case TypeBool:
		return r.ReadBool()
	case TypeDouble:
		return r.ReadFloat64()
	case TypeFloat:
		return r.ReadFloat32()
	case TypeInt:
		return r.ReadInt32()
	case TypeInt8:
		value, err := r.ReadByte()
		return int8(value), err
	case TypeInt16:
		return r.ReadInt16()
	case TypeInt64:
		return r.ReadInt64()
	case TypeName:
		return r.ReadName("")
	case TypeObject:
		return readObjectReference(r, -1)
	case TypeSoftObject:
		return readSoftObject(r)
	case TypeString:
		value, err := r.ReadString()
		if err != nil || value == nil {
			return "", err
		}
		return *value, nil
	case TypeUInt32:
		return r.ReadUInt32()
	case TypeUInt16:
		return r.ReadUInt16()
	case TypeUInt64:
		return r.ReadUInt64()
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
