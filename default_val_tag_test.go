package bkit

import (
	"fmt"
	"testing"
	"time"
)

type def struct {
	MyInt         int8          `default:"-1"`
	MyUint        uint          `default:"1"`
	MyString      string        `default:"hello"`
	MyBool        bool          `default:"true"`
	MyFloat       float32       `default:"66.6"`
	MySliceString []string      `default:"xx,xx,xx"`
	MySliceFloat  []float64     `default:"66.6,77.7"`
	MySliceInt    []int8        `default:"-1,0,9"`
	MySliceUint   []uint16      `default:"0,2,4"`
	MyDuration    time.Duration `default:"30s"` // 30s
	MyStruct      MyStruct
	MyStruct2
	Ptr       *Ptr
	Conns     []Conn
	NullVal   string `default:"30s" null:""`
	MyStruct3 MyStruct2
	ConnsPtr  []*Conn
}

type MyStruct struct {
	Key   string `default:"structKey"`
	Value MyStruct2
	S     []string `default:"xx,xx,xx"`
}

type MyStruct2 struct {
	Value int32    `default:"8"`
	S2    []string `default:"xx,xx,xx"`
}

type Conn struct {
	User     string
	Password string  `default:"Password"`
	Ptr      *string `default:"Conn.Ptr"`
	Struct   *MyStruct
}

type Ptr struct {
	Value string `default:"Point"`
}

func TestDefaultValueTag_SetDefaultVal(t *testing.T) {
	ptr := "string"
	def := def{
		Conns: []Conn{
			{
				User: "root",
				Struct: &MyStruct{
					Key: "不能修改我",
				},
			},
		},
		Ptr: new(Ptr),
		ConnsPtr: []*Conn{
			{
				User: "root.Prt",
			},
			{
				User: "root.Prt",
				Ptr:  &ptr,
				Struct: &MyStruct{
					Key: "不能修改我222",
				},
			},
		},
	}
	dvt := NewDefaultValueTag()
	if err := dvt.SetDefaultVal(&def); err != nil {
		t.Fatal(err)
	}
	if !def.MyBool {
		t.Fatal("bool")
	}
	if def.MyDuration.String() != "30s" {
		fmt.Println(def.MyDuration.String())
		t.Fatal("duration")
	}
	if def.MyInt != -1 {
		t.Fatal("int")
	}
	if def.MyUint != 1 {
		t.Fatal("uint")
	}
	if def.MyFloat != 66.6 {
		t.Fatal("float")
	}
	if def.MyStruct.Value.Value != def.Value {
		t.Fatal("struct")
	}
	if def.Conns[0].Password != "Password" {
		t.Fatal("conn")
	}
	fmt.Printf("%#v \n", def)
}
