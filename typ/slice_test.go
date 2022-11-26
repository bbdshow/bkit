package typ

import (
	"fmt"
	"testing"
)

func TestSliceRemoveDuplicate(t *testing.T) {
	type Info struct {
		Name string
	}
	type Attr struct {
		Age int
	}
	testCases := []struct {
		In  []interface{}
		Out []interface{}
	}{
		{
			In:  []interface{}{"a", "b", "a", "c", "D", "b", "C", "1", "1", "2", "a", "b"},
			Out: []interface{}{"a", "b", "c", "D", "C", "1", "2"},
		},
		{
			In:  []interface{}{"a", "b", "a"},
			Out: []interface{}{"a", "b"},
		},
		{
			In:  []interface{}{"aaa", "aaa", "a"},
			Out: []interface{}{"aaa", "a"},
		},
		{
			In:  []interface{}{"a", "aaa", "aaa"},
			Out: []interface{}{"a", "aaa"},
		},
		{
			In:  []interface{}{&Info{Name: "info"}, &Info{Name: "info1"}, &Info{Name: "info"}, nil},
			Out: []interface{}{&Info{Name: "info"}, &Info{Name: "info1"}},
		},
		{
			In:  []interface{}{Info{Name: "info"}, Info{Name: "info"}, Info{Name: "info"}},
			Out: []interface{}{Info{Name: "info"}},
		},
		{
			In:  []interface{}{Info{Name: "info"}, Info{Name: "info"}, Attr{Age: 1}, Attr{Age: 1}},
			Out: []interface{}{Info{Name: "info"}, Attr{Age: 1}},
		},
	}

	for i, tc := range testCases {
		in := tc.In
		SliceRemoveDuplicate(&in, func(i, j int) bool {
			if in[i] == nil || in[j] == nil {
				// 直接刪除
				return true
			}
			switch in[i].(type) {
			case *Info:
				switch in[j].(type) {
				case *Info:
					return in[i].(*Info).Name == in[j].(*Info).Name
				default:
					return false
				}
			case Attr:
				switch in[j].(type) {
				case Attr:
					return in[i].(Attr).Age == in[j].(Attr).Age
				default:
					return false
				}
			default:
				return in[i] == in[j]
			}
		})

		if i == 4 {
			if len(in) != len(tc.Out) {
				fmt.Println(tc.In, in, tc.Out)
				t.Fatalf("case index %d", i)
			}
		} else {
			if fmt.Sprintf("%v", in) != fmt.Sprintf("%v", tc.Out) {
				fmt.Println(tc.In, in, tc.Out)
				t.Fatalf("case index %d", i)
			}
		}
	}
}
