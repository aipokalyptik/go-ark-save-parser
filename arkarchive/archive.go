package arkarchive

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/google/uuid"
)

type Options struct {
	HeaderOnly  bool
	FromStore   bool
	ClusterDino bool
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
}

func Parse(data []byte, opts Options) (*Archive, error) {
	r := arkbinary.NewReader(data, nil)
	version := int32(7)
	var err error
	if !opts.ClusterDino {
		version, err = r.ReadInt32()
		if err != nil {
			return nil, err
		}
	}
	archive := &Archive{Version: version, Legacy: version != 7 && !opts.ClusterDino}
	if opts.HeaderOnly {
		return archive, nil
	}
	if archive.Legacy {
		return nil, fmt.Errorf("legacy archive object parsing is not implemented")
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
	return archive, nil
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
