package arkobject

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
)

func TestReadActorTransformReadsSevenDoubles(t *testing.T) {
	var buf bytes.Buffer
	for _, value := range []float64{1, 2, 3, 4, 5, 6, 7} {
		_ = binary.Write(&buf, binary.LittleEndian, value)
	}

	got, err := ReadActorTransform(arkbinary.NewReader(buf.Bytes(), nil))
	if err != nil {
		t.Fatalf("ReadActorTransform() error = %v", err)
	}
	if got.X != 1 || got.Y != 2 || got.Z != 3 || got.Pitch != 4 || got.Roll != 5 || got.Yaw != 6 || got.Quaternion != 7 {
		t.Fatalf("ActorTransform = %#v", got)
	}
}

func TestActorTransformDistanceAndFoundationChecks(t *testing.T) {
	left := ActorTransform{X: 0, Y: 0, Z: 0}
	right := ActorTransform{X: 300, Y: 400, Z: 0}

	if got := left.DistanceTo(right); got != 500 {
		t.Fatalf("DistanceTo() = %f, want 500", got)
	}
	if !left.IsWithinDistance(right, 510, 0, 10) {
		t.Fatalf("IsWithinDistance(distance) = false, want true")
	}
	if !left.IsWithinDistance(right, 0, 2, 10) {
		t.Fatalf("IsWithinDistance(foundations) = false, want true")
	}
}

func TestActorTransformMapCoordinatesUseUpstreamFormula(t *testing.T) {
	transform := ActorTransform{X: 0, Y: 0, Z: 0}

	coords := transform.AsMapCoords("Valguero")
	if math.Abs(coords.Lat-50) > 0.0001 || math.Abs(coords.Long-50) > 0.0001 {
		t.Fatalf("AsMapCoords(Valguero) = %#v, want center coordinates", coords)
	}
	back := coords.AsActorTransform("Valguero")
	if math.Abs(back.X) > 0.0001 || math.Abs(back.Y) > 0.0001 {
		t.Fatalf("AsActorTransform(Valguero) = %#v, want origin", back)
	}
}
