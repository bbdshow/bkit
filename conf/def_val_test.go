package conf

import (
	"fmt"
	"testing"
	"time"
)

type def struct {
	MyInt         int8              `defval:"-1"`
	MyUint        uint              `defval:"1"`
	MyString      string            `defval:"hello"`
	MyBool        bool              `defval:"true"`
	MyFloat       float32           `defval:"66.6"`
	MySliceString []string          `defval:"xx,xx,xx"`
	MySliceFloat  []float64         `defval:"66.6,77.7"`
	MySliceInt    []int8            `defval:"-1,0,9"`
	MySliceUint   []uint16          `defval:"0,2,4"`
	MyMap         map[string]string `defval:"a=1,2"`
	MyDuration    time.Duration     `defval:"30s"` // 30s
	MyStruct      MyStruct
	MyStruct2
	Ptr     *Ptr
	Conns   []Conn
	NullVal string `defval:"30s" null:""`
}

type MyStruct struct {
	Key   string `defval:"structKey"`
	Value MyStruct2
	S     []string `defval:"xx,xx,xx"`
}

type MyStruct2 struct {
	Value int32    `defval:"8"`
	S2    []string `defval:"xx,xx,xx"`
}

type Conn struct {
	Password string `defval:"Password" null:""`
}

type Ptr struct {
	Value string `defval:"Point" null:""`
}

func TestParseDefaultVal(t *testing.T) {
	def := def{}
	if err := ParseDefaultVal(&def); err != nil {
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
	fmt.Printf("%#v \n", def)
}

func TestInitialNullVal(t *testing.T) {
	def := def{
		Ptr: &Ptr{
			Value: "Point",
		},
	}
	if err := ParseDefaultVal(&def); err != nil {
		t.Fatal(err)
	}
	def.Conns = []Conn{{Password: "Password"}}
	if def.NullVal != "30s" {
		t.Fatal(def.NullVal)
	}
	fmt.Println("before", def, def.Ptr)
	if err := InitialNullVal(&def); err != nil {
		t.Fatal(err)
	}
	fmt.Println("after", def, def.Ptr)
	if def.NullVal != "" {
		t.Fatal("should null string", def.NullVal)
	}

}
