package caches

import (
	"strconv"
	"testing"
	"time"
)

func TestNewLRUMemory(t *testing.T) {
	type val struct {
		Val int
	}
	max := 100
	m := NewLRUMemory(max)
	c := 110
	total := c - (c - max)
	for c > 0 {
		c--
		if c%3 == 0 {
			err := m.SetWithTTL(strconv.Itoa(c), &val{
				Val: c,
			}, time.Second)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			err := m.Set(strconv.Itoa(c), &val{
				Val: c,
			})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	time.Sleep(2 * time.Second)
	i := 1
	for i < total {
		v, err := m.Get(strconv.Itoa(i))
		if err != nil {
			if i%3 == 0 {
				if !IsNotFoundErr(err) {
					t.Fatal(err)
				}
			} else {
				t.Fatal(err)
			}
		} else {
			if v.(*val).Val != i {
				t.Fatal("should equal")
			}
		}

		i++
	}
}
