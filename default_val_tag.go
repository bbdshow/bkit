package bkit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	errInvalidType = func(t string) error {
		return fmt.Errorf("default value invalid %s type", t)
	}
)

type DefaultValueTag struct {
	valueTag string
	valueSep string
}

// NewDefaultValueTag 根据标签生成默认值: 当存在默认值标签时，设置默认值
// tag 默认值标签 (default)
// sep 默认值分隔符 (,)
// 仅支持的格式: 当递归到最后一层时，支持的类型有: int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, string, time.Duration
// 当最后一层为 指针时， 只有当指针为空时，才会设置默认值
func NewDefaultValueTag() *DefaultValueTag {
	dvt := &DefaultValueTag{
		valueTag: "default",
		valueSep: ",",
	}
	return dvt
}

// SetTag 设置默认值标签
func (dvt *DefaultValueTag) SetTag(tag string) {
	dvt.valueTag = tag
}

// SetSep 设置默认值分隔符
func (dvt *DefaultValueTag) SetSep(sep string) {
	dvt.valueSep = sep
}

// SetDefaultVal 设置结构体默认值
// v 必须是指针类型
func (dvt *DefaultValueTag) SetDefaultVal(v interface{}) error {
	if dvt.valueTag == "" {
		dvt.valueTag = "default"
		dvt.valueSep = ","
	}
	val := reflect.ValueOf(v).Elem()
	return dvt.parse(dvt.valueTag, val)
}
func (dvt *DefaultValueTag) parse(tag string, val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.Struct {
			if err := dvt.parse(tag, field); err != nil {
				return err
			}
			continue
		}

		defVal, ok := val.Type().Field(i).Tag.Lookup(tag)
		if !ok {
			// 没有设置默认值，要判断是否为引用类型
			// 判断是否为 slice
			if field.Kind() == reflect.Slice {
				// 判断是否为结构体, 如果是结构体，递归处理
				if field.Len() > 0 && field.Index(0).Kind() == reflect.Struct {
					for ii := 0; ii < field.Len(); ii++ {
						if err := dvt.parse(tag, field.Index(ii)); err != nil {
							return err
						}
					}
				}
				// 判断是否为指针
				if field.Len() > 0 && field.Index(0).Kind() == reflect.Ptr {
					for ii := 0; ii < field.Len(); ii++ {
						if !field.Index(ii).IsNil() {
							if err := dvt.parse(tag, field.Index(ii).Elem()); err != nil {
								return err
							}
						}
					}
				}
			}
			// 判断是否为指针, 且必须是空指针，且是结构体
			if field.Kind() == reflect.Ptr && field.IsNil() {
				if IsStructPtr(field.Type().String()) {
					field.Set(reflect.New(field.Type().Elem()))
					if err := dvt.parse(tag, field.Elem()); err != nil {
						return err
					}
				}
			}
			// 处理之后，继续下一个字段
			continue
		}

		typ := field.Type().String()
		switch typ {
		case "int8", "int16", "int", "int32", "int64":
			v, err := strconv.ParseInt(defVal, 10, 64)
			if err != nil {
				return errInvalidType(typ)
			}
			if field.CanSet() {
				field.SetInt(v)
			}
		case "uint8", "uint16", "uint", "uint32", "uint64":
			v, err := strconv.ParseUint(defVal, 10, 64)
			if err != nil {
				return errInvalidType(typ)
			}
			if field.CanSet() {
				field.SetUint(v)
			}
		case "float32", "float64":
			v, err := strconv.ParseFloat(defVal, 64)
			if err != nil {
				return errInvalidType(typ)
			}
			if field.CanSet() {
				field.SetFloat(v)
			}
		case "bool":
			v, err := strconv.ParseBool(defVal)
			if err != nil {
				return errInvalidType(typ)
			}
			if field.CanSet() {
				field.SetBool(v)
			}
		case "string":
			if field.CanSet() {
				field.SetString(defVal)
			}
		case "time.Duration":
			v, err := time.ParseDuration(defVal)
			if err != nil {
				return errInvalidType(typ)
			}
			if field.CanSet() {
				field.SetInt(v.Nanoseconds())
			}
		case "[]int8", "[]int16", "[]int", "[]int32", "[]int64":
			if !field.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, dvt.valueSep)
			setVal := []int64{}
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
				rv = reflect.ValueOf(SliceInt[int8](setVal))
			case "[]int16":
				rv = reflect.ValueOf(SliceInt[int16](setVal))
			case "[]int":
				rv = reflect.ValueOf(SliceInt[int](setVal))
			case "[]int32":
				rv = reflect.ValueOf(SliceInt[int32](setVal))
			default:
				rv = reflect.ValueOf(setVal)
			}
			field.Set(rv)
		case "[]uint8", "[]uint16", "[]uint", "[]uint32", "[]uint64":
			if !field.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, dvt.valueSep)
			setVal := []uint64{}
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
				rv = reflect.ValueOf(SliceUint[uint8](setVal))
			case "[]uint16":
				rv = reflect.ValueOf(SliceUint[uint16](setVal))
			case "[]uint":
				rv = reflect.ValueOf(SliceUint[uint](setVal))
			case "[]uint32":
				rv = reflect.ValueOf(SliceUint[uint32](setVal))
			default:
				rv = reflect.ValueOf(setVal)
			}
			field.Set(rv)
		case "[]float32", "[]float64":
			if !field.CanSet() {
				continue
			}

			sliceVal := strings.Split(defVal, dvt.valueSep)
			setVal := []float64{}
			for _, v := range sliceVal {
				if v == "" {
					continue
				}
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return errInvalidType(typ)
				}
				setVal = append(setVal, f)
			}
			switch typ {
			case "[]float32":
				field.Set(reflect.ValueOf(SliceFloat[float32](setVal)))
			default:
				field.Set(reflect.ValueOf(setVal))
			}
		case "[]bool":
			if !field.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, dvt.valueSep)
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
			field.Set(reflect.ValueOf(setVal))
		case "[]string":
			if !field.CanSet() {
				continue
			}
			sliceVal := strings.Split(defVal, dvt.valueSep)
			setVal := make([]string, 0)
			for _, v := range sliceVal {
				if v == "" {
					continue
				}
				setVal = append(setVal, v)
			}

			field.Set(reflect.ValueOf(setVal))
		default:
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					if !IsStructPtr(typ) {
						field.Set(reflect.ValueOf(&defVal))
					}
				}
			}
		}
	}
	return nil
}
