package arkapi

import (
	"bytes"

	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
)

func writeIntProperty(buf *bytes.Buffer, name uint32, value int32) {
	testfixtures.WriteIntPropertyID(buf, name, 0x10000003, value)
}

func writeFloatProperty(buf *bytes.Buffer, name uint32, value float32) {
	testfixtures.WriteFloatPropertyID(buf, name, 0x1000000a, value)
}

func writeBoolProperty(buf *bytes.Buffer, name uint32, value bool) {
	testfixtures.WriteBoolPropertyID(buf, name, 0x1000000e, value)
}
