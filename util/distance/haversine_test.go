package distance

import (
	"fmt"
	"testing"
)

func TestDistance(t *testing.T) {
	p := Coord{
		Lat: 32.191,
		Lon: 119.238,
	}
	q := Coord{
		Lat: 32.177,
		Lon: 119.373,
	}
	mi, km := Distance(p, q)
	fmt.Println(mi, km)
}

func TestDistanceMeter(t *testing.T) {
	p := Coord{
		Lat: 29.568356,
		Lon: 106.470959,
	}
	q := Coord{
		Lat: 29.552642,
		Lon: 106.568955,
	}
	meter := DistanceMeter(p, q)
	fmt.Println(meter)
}
