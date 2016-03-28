/*
Package geo is a geometry/geography library in Go. Its purpose is to allow for
basic point, line and path operations in the context of online mapping.
*/
package geo

import (
	"fmt"
	"math"
	"strconv"

	"github.com/paulmach/go.geojson"
)

// A Point is a simple X/Y or Lng/Lat 2d point. [X, Y] or [Lng, Lat]
type Point [2]float64

// NewPoint creates a new point
func NewPoint(x, y float64) Point {
	return Point{x, y}
}

// NewPointFromQuadkey creates a new point from a quadkey.
// See http://msdn.microsoft.com/en-us/library/bb259689.aspx for more information
// about this coordinate system.
func NewPointFromQuadkey(key int64, level int) Point {
	var x, y int64

	var i uint
	for i = 0; i < uint(level); i++ {
		x |= (key & (1 << (2 * i))) >> i
		y |= (key & (1 << (2*i + 1))) >> (i + 1)
	}

	lng, lat := scalarMercatorInverse(uint64(x), uint64(y), uint64(level))
	return Point{lng, lat}
}

// NewPointFromQuadkeyString creates a new point from a quadkey string.
func NewPointFromQuadkeyString(key string) Point {
	i, _ := strconv.ParseInt(key, 4, 64)
	return NewPointFromQuadkey(i, len(key))
}

// NewPointFromGeoHash creates a new point at the center of the geohash range.
func NewPointFromGeoHash(hash string) Point {
	west, east, south, north := geoHash2ranges(hash)
	return NewPoint((west+east)/2.0, (north+south)/2.0)
}

// NewPointFromGeoHashInt64 creates a new point at the center of the
// integer version of a geohash range. bits indicates the precision of the hash.
func NewPointFromGeoHashInt64(hash int64, bits int) Point {
	west, east, south, north := geoHashInt2ranges(hash, bits)
	return NewPoint((west+east)/2.0, (north+south)/2.0)
}

// Point returns itself, so it implements the pointer interface.
func (p Point) Point() Point {
	return p
}

// DistanceFrom returns the Euclidean distance between the points.
func (p Point) DistanceFrom(point Point) float64 {
	d0 := (point[0] - p[0])
	d1 := (point[1] - p[1])
	return math.Sqrt(d0*d0 + d1*d1)
}

// SquaredDistanceFrom returns the squared Euclidean distance between the points.
// This avoids a sqrt computation.
func (p Point) SquaredDistanceFrom(point Point) float64 {
	d0 := (point[0] - p[0])
	d1 := (point[1] - p[1])
	return d0*d0 + d1*d1
}

// GeoDistanceFrom returns the geodesic distance in meters.
func (p Point) GeoDistanceFrom(point Point, haversine ...bool) float64 {
	dLat := deg2rad(point.Lat() - p.Lat())
	dLng := deg2rad(point.Lng() - p.Lng())

	if yesHaversine(haversine) {
		// yes trig functions
		dLat2Sin := math.Sin(dLat / 2)
		dLng2Sin := math.Sin(dLng / 2)
		a := dLat2Sin*dLat2Sin + math.Cos(deg2rad(p.Lat()))*math.Cos(deg2rad(point.Lat()))*dLng2Sin*dLng2Sin

		return 2.0 * EarthRadius * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	}

	// fast way using pythagorean theorem on an equirectangular projection
	x := dLng * math.Cos(deg2rad((p.Lat()+point.Lat())/2.0))
	return math.Sqrt(dLat*dLat+x*x) * EarthRadius
}

// BearingTo computes the direction one must start traveling on earth
// to be heading to the given point.
func (p Point) BearingTo(point Point) float64 {
	dLng := deg2rad(point.Lng() - p.Lng())

	pLatRad := deg2rad(p.Lat())
	pointLatRad := deg2rad(point.Lat())

	y := math.Sin(dLng) * math.Cos(pointLatRad)
	x := math.Cos(pLatRad)*math.Sin(pointLatRad) - math.Sin(pLatRad)*math.Cos(pointLatRad)*math.Cos(dLng)

	return rad2deg(math.Atan2(y, x))
}

// Quadkey returns the quad key for the given point at the provided level.
// See http://msdn.microsoft.com/en-us/library/bb259689.aspx for more information
// about this coordinate system.
func (p Point) Quadkey(level int) int64 {
	x, y := scalarMercatorProject(p.Lng(), p.Lat(), uint64(level))

	var i uint
	var result uint64
	for i = 0; i < uint(level); i++ {
		result |= (x & (1 << i)) << i
		result |= (y & (1 << i)) << (i + 1)
	}

	return int64(result)
}

// QuadkeyString returns the quad key for the given point at the provided level in string form
// See http://msdn.microsoft.com/en-us/library/bb259689.aspx for more information
// about this coordinate system.
func (p Point) QuadkeyString(level int) string {
	s := strconv.FormatInt(p.Quadkey(level), 4)

	// for zero padding
	zeros := "000000000000000000000000000000"
	return zeros[:((level+1)-len(s))/2] + s
}

const base32 = "0123456789bcdefghjkmnpqrstuvwxyz"

// GeoHash returns the geohash string of a point representing a lng/lat location.
// The resulting hash will be `GeoHashPrecision` characters long, default is 12.
// Optionally one can include their required number of chars precision.
func (p Point) GeoHash(chars ...int) string {
	precision := GeoHashPrecision
	if len(chars) > 0 {
		precision = chars[0]
	}

	// 15 must be greater than GeoHashPrecision. If not, panic!!
	var result [15]byte

	hash := p.GeoHashInt64(5 * precision)
	for i := 1; i <= precision; i++ {
		result[precision-i] = byte(base32[hash&0x1F])
		hash >>= 5
	}

	return string(result[:precision])
}

// GeoHashInt64 returns the integer version of the geohash
// down to the given number of bits.
// The main usecase for this function is to be able to do integer based ordering of points.
// In that case the number of bits should be the same for all encodings.
func (p Point) GeoHashInt64(bits int) (hash int64) {
	// This code was inspired by https://github.com/broady/gogeohash

	latMin, latMax := -90.0, 90.0
	lngMin, lngMax := -180.0, 180.0

	for i := 0; i < bits; i++ {
		hash <<= 1

		// interleave bits
		if i%2 == 0 {
			mid := (lngMin + lngMax) / 2.0
			if p[0] > mid {
				lngMin = mid
				hash |= 1
			} else {
				lngMax = mid
			}
		} else {
			mid := (latMin + latMax) / 2.0
			if p[1] > mid {
				latMin = mid
				hash |= 1
			} else {
				latMax = mid
			}
		}
	}

	return
}

// Add a point to the given point.
func (p Point) Add(vector Vector) Point {
	p[0] += vector[0]
	p[1] += vector[1]

	return p
}

// Subtract a point from the given point.
func (p Point) Subtract(point Point) Vector {
	return Vector{
		p[0] - point[0],
		p[1] - point[1],
	}
}

// ToArray casts the data to a [2]float64.
func (p Point) ToArray() [2]float64 {
	return [2]float64(p)
}

// Equal checks if the point represents the same point or vector.
func (p Point) Equal(point Point) bool {
	return p[0] == point[0] && p[1] == point[1]
}

// Lat returns the latitude/vertical component of the point.
func (p Point) Lat() float64 {
	return p[1]
}

// Lng returns the longitude/horizontal component of the point.
func (p Point) Lng() float64 {
	return p[0]
}

// X returns the x/horizontal component of the point.
func (p Point) X() float64 {
	return p[0]
}

// Y returns the y/vertical component of the point.
func (p Point) Y() float64 {
	return p[1]
}

// ToGeoJSON creates a new geojson feature with a point geometry.
func (p Point) ToGeoJSON() *geojson.Geometry {
	return geojson.NewPointGeometry([]float64{p[0], p[1]})
}

// ToWKT returns the point in WKT format, eg. POINT(30.5 10.5)
func (p Point) ToWKT() string {
	return p.String()
}

// String returns a string representation of the point.
// The format is WKT, e.g. POINT(30.5 10.5)
func (p Point) String() string {
	return fmt.Sprintf("POINT(%g %g)", p[0], p[1])
}
