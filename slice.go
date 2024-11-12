package bkit

import (
	"reflect"
	"strings"
)

var (
	Slice = SliceUtil{}
)

type SliceUtil struct{}

// UniqueSlice 切片去重，是否去掉空字符串，切片顺序不变
func (su SliceUtil) UniqueSlice(s []string, isDropEmpty bool) []string {
	if len(s) == 0 {
		return s
	}
	// 去重
	tmp := make([]string, 0, len(s))
	for _, v := range s {
		if isDropEmpty {
			if v == "" {
				continue
			}
		}
		if !su.InSlice(v, tmp) {
			tmp = append(tmp, v)
		}
	}
	return tmp

}

// InSlice 判断字符串是否在切片中
func (SliceUtil) InSlice(s string, slice []string) bool {
	isHit := false
	for _, v := range slice {
		if s == v {
			isHit = true
			break
		}
	}
	return isHit
}

// RemoveDuplicate 切片去重
func (SliceUtil) RemoveDuplicate(x interface{}, fn func(i, j int) bool) {
	v := reflect.ValueOf(x)
	if v.Kind() != reflect.Ptr {
		return
	}
	if v.Elem().Kind() != reflect.Slice {
		return
	}
	if v.Elem().Len() <= 1 {
		return
	}
	i := 0
	j := 1
	for {
		if i >= v.Elem().Len()-1 {
			break
		}

		isDuplicate := false
		for {
			if j >= v.Elem().Len() {
				break
			}
			if fn(i, j) {
				// 重复
				isDuplicate = true
				break
			}
			j++
		}
		if isDuplicate {
			//通过slice index 去掉重复元素
			//warning: 这里会产生申请内存扩容等操作，所以针对特别大的Slice，可能会有更高效的做法
			v.Elem().Set(reflect.AppendSlice(v.Elem().Slice(0, j), v.Elem().Slice(j+1, v.Elem().Len())))
			continue
		}
		i++
		j = i + 1
	}
}

type Set []string

// Has 是否存在
func (s Set) Has(v string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// Add 添加元素
func (s Set) Add(v string) Set {
	if !s.Has(v) {
		s = append(s, v)
	}
	return s
}

// Remove 删除元素
func (s Set) Remove(v string) Set {
	for i, vv := range s {
		if vv == v {
			s = append(s[:i], s[i+1:]...)
			break
		}
	}
	return s
}

// Union 并集
func (s Set) Union(s2 Set) Set {
	for _, v := range s2 {
		s = s.Add(v)
	}
	return s
}

// Intersect 交集
func (s Set) Intersect(s2 Set) Set {
	var s3 Set
	for _, v := range s2 {
		if s.Has(v) {
			s3 = s3.Add(v)
		}
	}
	return s3
}

// Difference 差集
func (s Set) Difference(s2 Set) Set {
	var s3 Set
	for _, v := range s {
		if !s2.Has(v) {
			s3 = s3.Add(v)
		}
	}
	return s3
}

func (s Set) TrimSpace(delEmpty bool) Set {
	for i := 0; i < len(s); i++ {
		v := strings.TrimSpace(s[i])
		if delEmpty && v == "" {
			s = append(s[:i], s[i+1:]...)
			i--
		} else {
			s[i] = v
		}
	}
	return s
}

type SetInt []int64

// Has 是否存在
func (s SetInt) Has(v int64) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}

// Add 添加元素
func (s SetInt) Add(v int64) SetInt {
	if !s.Has(v) {
		s = append(s, v)
	}
	return s
}

// Remove 删除元素
func (s SetInt) Remove(v int64) SetInt {
	for i, vv := range s {
		if vv == v {
			s = append(s[:i], s[i+1:]...)
			break
		}
	}
	return s
}

// Union 并集
func (s SetInt) Union(s2 SetInt) SetInt {
	for _, v := range s2 {
		s = s.Add(v)
	}
	return s
}

// Intersect 交集
func (s SetInt) Intersect(s2 SetInt) SetInt {
	var s3 SetInt
	for _, v := range s2 {
		if s.Has(v) {
			s3 = s3.Add(v)
		}
	}
	return s3
}

// Difference 差集
func (s SetInt) Difference(s2 SetInt) SetInt {
	var s3 SetInt
	for _, v := range s {
		if !s2.Has(v) {
			s3 = s3.Add(v)
		}
	}
	return s3
}
