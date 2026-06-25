package arkobject

import (
	"math"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
)

const FoundationDistance = 300.0

type ActorTransform struct {
	X          float64
	Y          float64
	Z          float64
	Pitch      float64
	Roll       float64
	Yaw        float64
	Quaternion float64
	InCryopod  bool
}

func ReadActorTransform(r *arkbinary.Reader) (ActorTransform, error) {
	x, err := r.ReadFloat64()
	if err != nil {
		return ActorTransform{}, err
	}
	y, err := r.ReadFloat64()
	if err != nil {
		return ActorTransform{}, err
	}
	z, err := r.ReadFloat64()
	if err != nil {
		return ActorTransform{}, err
	}
	pitch, err := r.ReadFloat64()
	if err != nil {
		return ActorTransform{}, err
	}
	roll, err := r.ReadFloat64()
	if err != nil {
		return ActorTransform{}, err
	}
	yaw, err := r.ReadFloat64()
	if err != nil {
		return ActorTransform{}, err
	}
	quat, err := r.ReadFloat64()
	if err != nil {
		return ActorTransform{}, err
	}
	return ActorTransform{X: x, Y: y, Z: z, Pitch: pitch, Roll: roll, Yaw: yaw, Quaternion: quat}, nil
}

func (a ActorTransform) DistanceTo(other ActorTransform) float64 {
	if a.InCryopod || other.InCryopod {
		return math.Inf(1)
	}
	return math.Sqrt(math.Pow(a.X-other.X, 2) + math.Pow(a.Y-other.Y, 2) + math.Pow(a.Z-other.Z, 2))
}

func (a ActorTransform) IsWithinDistance(other ActorTransform, distance float64, foundations int, tolerance float64) bool {
	if a.InCryopod || other.InCryopod {
		return false
	}
	if distance != 0 {
		return a.DistanceTo(other)+tolerance <= distance
	}
	if foundations != 0 {
		return a.DistanceTo(other)+tolerance <= float64(foundations)*FoundationDistance
	}
	return false
}

func (a ActorTransform) AsMapCoords(mapName string) MapCoords {
	lat, long := mapCoordinateParametersFor(mapName).TransformTo(a.X, a.Y)
	return MapCoords{Lat: lat, Long: long, InCryopod: a.InCryopod}
}

func (a ActorTransform) IsAtMapCoordinate(mapName string, coords MapCoords, tolerance float64) bool {
	if a.InCryopod {
		return false
	}
	own := a.AsMapCoords(mapName)
	return math.Abs(own.Lat-coords.Lat) <= tolerance && math.Abs(own.Long-coords.Long) <= tolerance
}

type MapCoords struct {
	Lat       float64
	Long      float64
	InCryopod bool
}

func (m MapCoords) DistanceTo(other MapCoords) float64 {
	if m.InCryopod || other.InCryopod {
		return math.Inf(1)
	}
	return math.Sqrt(math.Pow(m.Lat-other.Lat, 2) + math.Pow(m.Long-other.Long, 2))
}

func (m MapCoords) AsActorTransform(mapName string) ActorTransform {
	vector := mapCoordinateParametersFor(mapName).TransformFrom(m.Lat, m.Long)
	return ActorTransform{X: vector.X, Y: vector.Y, Z: vector.Z, InCryopod: m.InCryopod}
}

type mapVector struct {
	X float64
	Y float64
	Z float64
}

type mapCoordinateParameters struct {
	OriginMinX float64
	OriginMinY float64
	OriginMinZ float64
	OriginMaxX float64
	OriginMaxY float64
	OriginMaxZ float64
}

func (p mapCoordinateParameters) TransformTo(x float64, y float64) (float64, float64) {
	yMaxDiff := y - p.OriginMaxY
	xMaxDiff := x - p.OriginMaxX
	originYDiff := p.OriginMinY - p.OriginMaxY
	originXDiff := p.OriginMinX - p.OriginMaxX
	latRatio := yMaxDiff / originYDiff
	longRatio := xMaxDiff / originXDiff
	return lerp(100, 0, latRatio), lerp(100, 0, longRatio)
}

func (p mapCoordinateParameters) TransformFrom(lat float64, long float64) mapVector {
	originYDiff := p.OriginMinY - p.OriginMaxY
	originXDiff := p.OriginMinX - p.OriginMaxX
	latRatio := invLerp(100, 0, lat)
	longRatio := invLerp(100, 0, long)
	yMaxDiff := latRatio * originYDiff
	xMaxDiff := longRatio * originXDiff
	return mapVector{X: xMaxDiff + p.OriginMaxX, Y: yMaxDiff + p.OriginMaxY}
}

func mapCoordinateParametersFor(mapName string) mapCoordinateParameters {
	switch normalizedMapName(mapName) {
	case "SCORCHEDEARTH":
		return mapCoordinateParameters{-393650, -393650, -25515, 393750, 393750, 66645}
	case "THECENTER":
		return mapCoordinateParameters{-524364, -337215, -171880.46875, 513040, 700189, 101159.6875}
	case "ABERRATION":
		return mapCoordinateParameters{-400000, -400000, -15000, 400000, 400000, 54695}
	case "EXTINCTION":
		return mapCoordinateParameters{-342900, -342900, -15000, 342900, 342900, 54695}
	case "RAGNAROK":
		return mapCoordinateParameters{-655000, -655000, -655000, 655000, 655000, 54695}
	case "ASTRAEOS":
		return mapCoordinateParameters{-800000, -800000, -15000, 800000, 800000, 54695}
	case "SVARTALFHEIM":
		return mapCoordinateParameters{-203250, -203250, -15000, 203250, 203250, 54695}
	case "VALGUERO":
		return mapCoordinateParameters{-408000, -408000, -655000, 408000, 408000, 54695}
	case "CLUBARK":
		return mapCoordinateParameters{-12812, -15121, -12500, 12078, 9770, 12500}
	case "LOSTCOLONY":
		return mapCoordinateParameters{-408000, -408000, -15000, 408000, 408000, 54695}
	default:
		return mapCoordinateParameters{-342900, -342900, -15000, 342900, 342900, 54695}
	}
}

func normalizedMapName(value string) string {
	value = strings.ToUpper(value)
	value = strings.TrimSuffix(value, "_WP")
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}

func lerp(a float64, b float64, t float64) float64 {
	return (1-t)*a + t*b
}

func invLerp(a float64, b float64, v float64) float64 {
	return (v - a) / (b - a)
}
