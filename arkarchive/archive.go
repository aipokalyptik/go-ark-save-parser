package arkarchive

import (
	"errors"
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

type Format int

const (
	FormatAuto Format = iota
	FormatModern
	FormatLegacy
	FormatClusterDino
)

var ErrLegacyArchiveUnsupported = errors.New("legacy archive object parsing is not implemented")

type Options struct {
	HeaderOnly       bool
	FromStore        bool
	ClusterDino      bool
	Format           Format
	StrictProperties bool
}

type Archive struct {
	Version int32
	Legacy  bool
	Objects []Object
}

type Object struct {
	UUID             uuid.UUID
	ClassName        string
	Item             uint32
	Names            []string
	FromDataFile     uint32
	DataFileIndex    int32
	PropertiesOffset int32
	Properties       []arkproperty.Property
	PropertyError    error
}

func Parse(data []byte, opts Options) (*Archive, error) {
	format := opts.Format
	if opts.ClusterDino {
		format = FormatClusterDino
	}
	r := arkbinary.NewReader(data, nil)
	version := int32(7)
	var err error
	if format != FormatClusterDino {
		version, err = r.ReadInt32()
		if err != nil {
			return nil, err
		}
	}
	legacy := format == FormatLegacy || (format == FormatAuto && version != 7)
	if format == FormatModern && version != 7 {
		return nil, fmt.Errorf("modern archive format requires version 7, got %d", version)
	}
	archive := &Archive{Version: version, Legacy: legacy}
	if opts.HeaderOnly {
		return archive, nil
	}
	if archive.Legacy {
		return nil, ErrLegacyArchiveUnsupported
	}
	if _, err := r.ReadInt32(); err != nil {
		return nil, err
	}
	if _, err := r.ReadInt32(); err != nil {
		return nil, err
	}
	count, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}
	if count < 0 {
		return nil, fmt.Errorf("negative archive object count %d", count)
	}
	archive.Objects = make([]Object, 0, count)
	for i := int32(0); i < count; i++ {
		obj, err := readObject(r, opts.ClusterDino)
		if err != nil {
			return nil, err
		}
		archive.Objects = append(archive.Objects, obj)
	}
	if err := readObjectProperties(r, archive, format, opts.StrictProperties); err != nil {
		return nil, err
	}
	return archive, nil
}

func readObjectProperties(r *arkbinary.Reader, archive *Archive, format Format, strict bool) error {
	if format == FormatClusterDino {
		return nil
	}
	for i := range archive.Objects {
		start := int(archive.Objects[i].PropertiesOffset) + 1
		if start < 0 || start >= r.Size() {
			continue
		}
		end := r.Size()
		if i+1 < len(archive.Objects) {
			next := int(archive.Objects[i+1].PropertiesOffset) + 1
			if next > start && next <= r.Size() {
				end = next
			}
		}
		if err := r.SetPosition(start); err != nil {
			return err
		}
		props, err := arkproperty.ParseAllPartial(r, end)
		archive.Objects[i].Properties = props
		if err != nil {
			err = fmt.Errorf("parse archive object %s properties at %d: %w", archive.Objects[i].UUID, start, err)
			archive.Objects[i].PropertyError = err
			if strict {
				return err
			}
			continue
		}
	}
	return nil
}

func readObject(r *arkbinary.Reader, clusterDino bool) (Object, error) {
	id, err := r.ReadUUID()
	if err != nil {
		return Object{}, err
	}
	className, err := r.ReadString()
	if err != nil {
		return Object{}, err
	}
	item, err := r.ReadUInt32()
	if err != nil {
		return Object{}, err
	}
	names, err := readStringArray(r)
	if err != nil {
		return Object{}, err
	}
	fromDataFile, err := r.ReadUInt32()
	if err != nil {
		return Object{}, err
	}
	dataFileIndex, err := r.ReadInt32()
	if err != nil {
		return Object{}, err
	}
	hasTransform, err := r.ReadUInt32()
	if err != nil {
		return Object{}, err
	}
	if hasTransform != 0 {
		if clusterDino {
			if err := r.Skip(16); err != nil {
				return Object{}, err
			}
		} else if err := r.Skip(28); err != nil {
			return Object{}, err
		}
	}
	propertiesOffset, err := r.ReadInt32()
	if err != nil {
		return Object{}, err
	}
	zero, err := r.ReadUInt32()
	if err != nil {
		return Object{}, err
	}
	if zero != 0 {
		return Object{}, fmt.Errorf("expected zero after archive object properties offset, got %#x", zero)
	}
	value := ""
	if className != nil {
		value = *className
	}
	return Object{
		UUID:             id,
		ClassName:        value,
		Item:             item,
		Names:            names,
		FromDataFile:     fromDataFile,
		DataFileIndex:    dataFileIndex,
		PropertiesOffset: propertiesOffset,
	}, nil
}

func readStringArray(r *arkbinary.Reader) ([]string, error) {
	count, err := r.ReadUInt32()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, count)
	for i := uint32(0); i < count; i++ {
		value, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		if value == nil {
			out = append(out, "")
		} else {
			out = append(out, *value)
		}
	}
	return out, nil
}
