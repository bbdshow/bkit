package caches

import (
	"testing"
)

func TestLimitMemoryCache(t *testing.T) {
	m := NewLimitMemoryCache(-1)
	type val struct {
		Val int
	}
	v := &val{
		Val: 1,
	}
	m.Set("1", v)
	v1, err := m.Get("1")
	if err != nil {
		t.Fatal(err)
	}
	if v1.(*val).Val != 1 {
		t.Fatal("should equal")
	}
}
