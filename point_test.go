package geo

import (
	"math"
	"strings"
	"testing"
)

var citiesGeoHash = [][3]interface{}{
	{57.09700, 9.85000, "u4phb4hw"},
	{49.03000, -122.32000, "c29nbt9k3q"},
	{39.23500, -76.17490, "dqcz4we0k"},
	{-34.7666, 138.53670, "r1fd0qzmg"},
}

func TestNewPoint(t *testing.T) {
	p := NewPoint(1, 2)
	if p.X() != 1 || p.Lng() != 1 {
		t.Errorf("point, expected 1, got %f", p.X())
	}

	if p.Y() != 2 || p.Lat() != 2 {
		t.Errorf("point, expected 2, got %f", p.Y())
	}
}

func TestPointQuadkey(t *testing.T) {
	p := Point{
		-87.65005229999997,
		41.850033,
	}

	if k := p.Quadkey(15); k != 212521785 {
		t.Errorf("point quadkey, incorrect got %d", k)
	}

	// default level
	level := 30
	for _, city := range cities {
		p := Point{
			city[1],
			city[0],
		}
		key := p.Quadkey(level)

		p = NewPointFromQuadkey(key, level)

		if math.Abs(p.Lat()-city[0]) > epsilon {
			t.Errorf("point quadkey, latitude miss match: %f != %f", p.Lat(), city[0])
		}

		if math.Abs(p.Lng()-city[1]) > epsilon {
			t.Errorf("point quadkey, longitude miss match: %f != %f", p.Lng(), city[1])
		}
	}
}

func TestPointQuadkeyString(t *testing.T) {
	p := Point{
		-87.65005229999997,
		41.850033,
	}

	if k := p.QuadkeyString(15); k != "030222231030321" {
		t.Errorf("point quadkey string, incorrect got %s", k)
	}

	// default level
	level := 30
	for _, city := range cities {
		p := Point{
			city[1],
			city[0],
		}

		key := p.QuadkeyString(level)

		p = NewPointFromQuadkeyString(key)
		if math.Abs(p.Lat()-city[0]) > epsilon {
			t.Errorf("point quadkey, latitude miss match: %f != %f", p.Lat(), city[0])
		}

		if math.Abs(p.Lng()-city[1]) > epsilon {
			t.Errorf("point quadkey, longitude miss match: %f != %f", p.Lng(), city[1])
		}
	}
}

func TestNewPointFromGeoHash(t *testing.T) {
	for _, c := range citiesGeoHash {
		p := NewPointFromGeoHash(c[2].(string))
		if d := p.GeoDistanceFrom(NewPoint(c[1].(float64), c[0].(float64))); d > 10 {
			t.Errorf("point, new from geohash expected distance %f", d)
		}
	}
}

func TestNewPointFromGeoHashInt64(t *testing.T) {
	for _, c := range citiesGeoHash {
		var hash int64
		for _, r := range c[2].(string) {
			hash <<= 5
			hash |= int64(strings.Index("0123456789bcdefghjkmnpqrstuvwxyz", string(r)))
		}

		p := NewPointFromGeoHashInt64(hash, 5*len(c[2].(string)))
		if d := p.GeoDistanceFrom(NewPoint(c[1].(float64), c[0].(float64))); d > 10 {
			t.Errorf("point, new from geohash expected distance %f", d)
		}
	}
}

func TestPointDistanceFrom(t *testing.T) {
	p1 := NewPoint(0, 0)
	p2 := NewPoint(3, 4)

	if d := p1.DistanceFrom(p2); d != 5 {
		t.Errorf("point, distanceFrom expected 5, got %f", d)
	}

	if d := p2.DistanceFrom(p1); d != 5 {
		t.Errorf("point, distanceFrom expected 5, got %f", d)
	}
}

func TestPointSquaredDistanceFrom(t *testing.T) {
	p1 := NewPoint(0, 0)
	p2 := NewPoint(3, 4)

	if d := p1.SquaredDistanceFrom(p2); d != 25 {
		t.Errorf("point, squaredDistanceFrom expected 25, got %f", d)
	}

	if d := p2.SquaredDistanceFrom(p1); d != 25 {
		t.Errorf("point, squaredDistanceFrom expected 25, got %f", d)
	}
}

func TestPointGeoDistanceFrom(t *testing.T) {
	p1 := NewPoint(-1.8444, 53.1506)
	p2 := NewPoint(0.1406, 52.2047)

	if d := p1.GeoDistanceFrom(p2, true); math.Abs(d-170389.801924) > epsilon {
		t.Errorf("incorrect geodistance, got %v", d)
	}

	if d := p1.GeoDistanceFrom(p2, false); math.Abs(d-170400.503437) > epsilon {
		t.Errorf("incorrect geodistance, got %v", d)
	}
}

func TestPointBearingTo(t *testing.T) {
	p1 := NewPoint(0, 0)
	p2 := NewPoint(0, 1)

	if d := p1.BearingTo(p2); d != 0 {
		t.Errorf("point, bearingTo expected 0, got %f", d)
	}

	if d := p2.BearingTo(p1); d != 180 {
		t.Errorf("point, bearingTo expected 180, got %f", d)
	}

	p1 = NewPoint(0, 0)
	p2 = NewPoint(1, 0)

	if d := p1.BearingTo(p2); d != 90 {
		t.Errorf("point, bearingTo expected 90, got %f", d)
	}

	if d := p2.BearingTo(p1); d != -90 {
		t.Errorf("point, bearingTo expected -90, got %f", d)
	}

	p1 = NewPoint(-1.8444, 53.1506)
	p2 = NewPoint(0.1406, 52.2047)

	if d := p1.BearingTo(p2); math.Abs(127.373351-d) > epsilon {
		t.Errorf("point, bearingTo got %f", d)
	}
}

func TestPointAdd(t *testing.T) {
	p := NewPoint(1, 2)
	v := NewVector(3, 4)

	answer := NewPoint(4, 6)
	p2 := p.Add(v)
	if !p2.Equal(answer) {
		t.Errorf("point, add expect %v == %v", p2, answer)
	}
}

func TestPointSubtract(t *testing.T) {
	p1 := NewPoint(3, 4)
	p2 := NewPoint(1, 3)

	answer := NewVector(2, 1)
	v := p1.Subtract(p2)
	if !v.Equal(answer) {
		t.Errorf("point, subtract expect %v == %v", v, answer)
	}
}

func TestPointGeoHash(t *testing.T) {
	for _, c := range citiesGeoHash {
		hash := NewPoint(c[1].(float64), c[0].(float64)).GeoHash()
		if !strings.HasPrefix(hash, c[2].(string)) {
			t.Errorf("point, geohash expected %s, got %s", c[2].(string), hash)
		}
	}

	for _, c := range citiesGeoHash {
		hash := NewPoint(c[1].(float64), c[0].(float64)).GeoHash(len(c[2].(string)))
		if hash != c[2].(string) {
			t.Errorf("point, geohash expected %s, got %s", c[2].(string), hash)
		}
	}
}

func TestPointEqual(t *testing.T) {
	p1 := NewPoint(1, 0)
	p2 := NewPoint(1, 0)

	p3 := NewPoint(2, 3)
	p4 := NewPoint(2, 4)

	if !p1.Equal(p2) {
		t.Errorf("point, equals expect %v == %v", p1, p2)
	}

	if p2.Equal(p3) {
		t.Errorf("point, equals expect %v != %v", p2, p3)
	}

	if p3.Equal(p4) {
		t.Errorf("point, equals expect %v != %v", p3, p4)
	}
}

func TestPointToGeoJSON(t *testing.T) {
	// TODO: fix
	// p := NewPoint(1, 2.5)

	// f := p.ToGeoJSON()
	// if !f.Geometry.IsPoint() {
	// 	t.Errorf("point, should be point geometry")
	// }
}

func TestPointToWKT(t *testing.T) {
	p := NewPoint(1, 2.5)

	answer := "POINT(1 2.5)"
	if s := p.ToWKT(); s != answer {
		t.Errorf("point, string expected %s, got %s", answer, s)
	}
}

func TestPointString(t *testing.T) {
	p := NewPoint(1, 2.5)

	answer := "POINT(1 2.5)"
	if s := p.String(); s != answer {
		t.Errorf("point, string expected %s, got %s", answer, s)
	}
}
