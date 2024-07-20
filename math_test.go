package bkit

import (
	"fmt"
	"testing"
)

func TestMath_Distance(t *testing.T) {
	p := Coord{
		Lat: 32.191,
		Lon: 119.238,
	}
	q := Coord{
		Lat: 32.177,
		Lon: 119.373,
	}
	mi, km := Math.Distance(p, q)
	fmt.Println(mi, km)
}
