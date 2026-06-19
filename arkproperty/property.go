package arkproperty

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
)

type Type string

const (
	TypeBool   Type = "Boolean"
	TypeInt    Type = "Int"
	TypeString Type = "String"
)

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
	default:
		return nil, fmt.Errorf("unsupported property type %q for %q", typeName, key)
	}

	return prop, nil
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
