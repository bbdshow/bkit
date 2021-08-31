package defval

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	defValTag = "defval"
	nullTag   = "null"
)

// NullVal 清空字段当前值
func InitialNullVal(v interface{}) error {
	val := reflect.ValueOf(v).Elem()
	return initial(nullTag, val)
}

func initial(tag string, val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		filed := val.Field(i)
		if filed.Kind() == reflect.Slice {
			for ii := 0; ii < filed.Len(); ii++ {
				if filed.Index(ii).Kind() == reflect.Struct {
					if err := initial(tag, filed.Index(ii)); err != nil {
						return err
					}
				}
			}
		}

		if filed.Kind() == reflect.Ptr {
			if !filed.IsNil() {
				if err := initial(tag, filed.Elem()); err != nil {
					return err
				}
			}
		}

		if filed.Kind() == reflect.Struct {
			if err := initial(tag, filed); err != nil {
				return err
			}
		}

		_, ok := val.Type().Field(i).Tag.Lookup(tag)
		if !ok {
			continue
		}

		typ := filed.Type().String()
		switch typ {
		case "int8", "int16", "int", "int32", "int64":
			if filed.CanSet() {
				filed.SetInt(0)
			}
		case "uint8", "uint16", "uint", "uint32", "uint64":
			if filed.CanSet() {
				filed.SetUint(0)
			}
		case "float32", "float64":
			if filed.CanSet() {
				filed.SetFloat(0)
			}
		case "bool":
			if filed.CanSet() {
				filed.SetBool(false)
			}
		case "string":
			if filed.CanSet() {
				filed.SetString("")
			}
		case "time.Duration":
			if filed.CanSet() {
				filed.SetInt(0)
			}
		case "[]int8", "[]int16", "[]int", "[]int32", "[]int64":
			if !filed.CanSet() {
				continue
			}
			filed.Set(reflect.ValueOf(nil))
		case "[]uint8", "[]uint16", "[]uint", "[]uint32", "[]uint64":
			if !filed.CanSet() {
				continue
			}
			filed.Set(reflect.ValueOf(nil))
		case "[]float32", "[]float64":
			if !filed.CanSet() {
				continue
			}
			filed.Set(reflect.ValueOf(nil))
		case "[]bool":
			if !filed.CanSet() {
				continue
			}
			filed.Set(reflect.ValueOf(nil))
		case "[]string":
			if !filed.CanSet() {
				continue
			}
			filed.Set(reflect.ValueOf(nil))
		case "map[string]string":
			if !filed.CanSet() {
				continue
			}
			filed.Set(reflect.ValueOf(nil))
		}
	}
	return nil
}

// ParseDefaultVal 提取当前结构的 default tag 作为field的默认值, 只对 Struct 有效
func ParseDefaultVal(v interface{}) error {
	val := reflect.ValueOf(v).Elem()
	return parse(defValTag, val)
}

func parse(tag string, val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		filed := val.Field(i)

		if filed.Kind() == reflect.Struct {
			if err := parse(tag, filed); err != nil {
				return err
			}
		}

		defVal, ok := val.Type().Field(i).Tag.Lookup(tag)
		if !ok {
			continue
		}

		typ := filed.Type().String()
		switch typ {
		case "int8", "int16", "int", "int32", "int64":
			v, err := strconv.ParseInt(defVal, 10, 64)
			if err != nil {
				return errInvalidType(typ)
			}
			if filed.CanSet() {
				filed.SetInt(v)
			}
		case "uint8", "uint16", "uint", "uint32", "uint64":
			v, err := strconv.ParseUint(defVal, 10, 64)
			if err != nil {
				return errInvalidType(typ)
			}
			if filed.CanSet() {
				filed.SetUint(v)
			}
		case "float32", "float64":
			v, err := strconv.ParseFloat(defVal, 10)
			if err != nil {
				return errInvalidType(typ)
			}
			if filed.CanSet() {
				filed.SetFloat(v)
			}
		case "bool":
			v, err := strconv.ParseBool(defVal)
			if err != nil {
				return errInvalidType(typ)
			}
			if filed.CanSet() {
				filed.SetBool(v)
			}
		case "string":
			if filed.CanSet() {
				filed.SetString(defVal)
			}
		case "time.Duration":
			v, err := time.ParseDuration(defVal)
			if err != nil {
				return errInvalidType(typ)
			}
			if filed.CanSet() {
				filed.SetInt(v.Nanoseconds())
			}
		case "[]int8", "[]int16", "[]int", "[]int32", "[]int64":
			if !filed.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, ",")
			setVal := sliceInt{}
			for _, v := range sliceVal {
				if v == "" {
					continue
				}
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return errInvalidType(typ)
				}
				setVal = append(setVal, i)
			}
			var rv reflect.Value
			switch typ {
			case "[]int8":
				rv = reflect.ValueOf(setVal.Int8())
			case "[]int16":
				rv = reflect.ValueOf(setVal.Int16())
			case "[]int":
				rv = reflect.ValueOf(setVal.Int())
			case "[]int32":
				rv = reflect.ValueOf(setVal.Int32())
			default:
				rv = reflect.ValueOf(setVal)
			}
			filed.Set(rv)
		case "[]uint8", "[]uint16", "[]uint", "[]uint32", "[]uint64":
			if !filed.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, ",")
			setVal := sliceUint{}
			for _, v := range sliceVal {
				if v == "" {
					continue
				}
				i, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return errInvalidType(typ)
				}
				setVal = append(setVal, i)
			}
			var rv reflect.Value
			switch typ {
			case "[]uint8":
				rv = reflect.ValueOf(setVal.Uint8())
			case "[]uint16":
				rv = reflect.ValueOf(setVal.Uint16())
			case "[]uint":
				rv = reflect.ValueOf(setVal.Uint())
			case "[]uint32":
				rv = reflect.ValueOf(setVal.Uint32())
			default:
				rv = reflect.ValueOf(setVal)
			}
			filed.Set(rv)
		case "[]float32", "[]float64":
			if !filed.CanSet() {
				continue
			}

			sliceVal := strings.Split(defVal, ",")
			setVal := make(sliceFloat, 0)
			for _, v := range sliceVal {
				if v == "" {
					continue
				}
				f, err := strconv.ParseFloat(v, 10)
				if err != nil {
					return errInvalidType(typ)
				}
				setVal = append(setVal, f)
			}
			switch typ {
			case "[]float32":
				filed.Set(reflect.ValueOf(setVal.Float32()))
			default:
				filed.Set(reflect.ValueOf(setVal))
			}
		case "[]bool":
			if !filed.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, ",")
			setVal := make([]bool, 0)
			for _, v := range sliceVal {
				if v == "" {
					continue
				}
				b, err := strconv.ParseBool(v)
				if err != nil {
					return errInvalidType(typ)
				}
				setVal = append(setVal, b)
			}
			filed.Set(reflect.ValueOf(setVal))
		case "[]string":
			if !filed.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, ",")
			setVal := make([]string, 0)
			for _, v := range sliceVal {
				if v == "" {
					continue
				}
				setVal = append(setVal, v)
			}

			filed.Set(reflect.ValueOf(setVal))

		case "map[string]string":
			if !filed.CanSet() {
				continue
			}
			sliceStr := strings.Split(defVal, ",")
			setVal := make(map[string]string)
			for _, v := range sliceStr {
				vs := strings.Split(v, "=")
				if len(vs) != 2 {
					continue
				}
				setVal[vs[0]] = vs[1]
			}
			filed.Set(reflect.ValueOf(setVal))
		}
	}

	return nil
}

type sliceFloat []float64

func (s sliceFloat) Float32() []float32 {
	val := make([]float32, len(s))
	for i, v := range s {
		val[i] = float32(v)
	}
	return val
}

type sliceInt []int64

func (s sliceInt) Int8() []int8 {
	val := make([]int8, len(s))
	for i, v := range s {
		val[i] = int8(v)
	}
	return val
}
func (s sliceInt) Int16() []int16 {
	val := make([]int16, len(s))
	for i, v := range s {
		val[i] = int16(v)
	}
	return val
}
func (s sliceInt) Int() []int {
	val := make([]int, len(s))
	for i, v := range s {
		val[i] = int(v)
	}
	return val
}
func (s sliceInt) Int32() []int32 {
	val := make([]int32, len(s))
	for i, v := range s {
		val[i] = int32(v)
	}
	return val
}

type sliceUint []uint64

func (s sliceUint) Uint8() []uint8 {
	val := make([]uint8, len(s))
	for i, v := range s {
		val[i] = uint8(v)
	}
	return val
}
func (s sliceUint) Uint16() []uint16 {
	val := make([]uint16, len(s))
	for i, v := range s {
		val[i] = uint16(v)
	}
	return val
}
func (s sliceUint) Uint() []uint {
	val := make([]uint, len(s))
	for i, v := range s {
		val[i] = uint(v)
	}
	return val
}
func (s sliceUint) Uint32() []uint32 {
	val := make([]uint32, len(s))
	for i, v := range s {
		val[i] = uint32(v)
	}
	return val
}
func errInvalidType(typ string) error {
	return fmt.Errorf("default value invalid %s type", typ)
}
