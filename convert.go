package bkit

func VToPrt[T any](v T) (prt *T) {
	prt = &v
	return
}

func PrtToV[T any](prt *T) (v T) {
	if prt != nil {
		v = *prt
	}
	return
}

func PtrSliceString(v []*string) []string {
	if v == nil {
		return nil
	}
	out := make([]string, 0, len(v))
	for _, item := range v {
		out = append(out, PrtToV(item))
	}
	return out
}

func SliceStringPtr(v []string) []*string {
	s := make([]*string, len(v))
	for i := 0; i < len(v); i++ {
		s[i] = &v[i]
	}
	return s
}

// ConvertUtil 类型转换
type ConvertUtil struct{}
