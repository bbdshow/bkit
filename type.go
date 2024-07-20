package bkit

import "strings"

type StringSplit string

func (ss StringSplit) String() string {
	return string(ss)
}

func (ss StringSplit) Unmarshal(sep string) []string {
	strSlice := make([]string, 0)
	for _, s := range strings.Split(string(ss), sep) {
		s = strings.TrimSpace(s)
		if s != "" {
			strSlice = append(strSlice, s)
		}
	}
	return strSlice
}

func (ss StringSplit) Marshal(val []string, sep string) StringSplit {
	strSlice := make([]string, 0)
	for _, s := range val {
		s = strings.TrimSpace(s)
		if s != "" {
			strSlice = append(strSlice, s)
		}
	}
	return StringSplit(strings.Join(strSlice, sep))
}

func (ss StringSplit) Has(v string, sep string) bool {
	strSlice := ss.Unmarshal(sep)
	for _, s := range strSlice {
		if s == v {
			return true
		}
	}
	return false
}

func (ss StringSplit) Merge(v StringSplit, sep string) StringSplit {
	strSlice := ss.Unmarshal(sep)
	for _, s := range v.Unmarshal(sep) {
		hit := false
		for _, s1 := range strSlice {
			if s1 == s {
				hit = true
				break
			}
		}
		if !hit {
			strSlice = append(strSlice, s)
		}
	}
	return StringSplit(strings.Join(strSlice, sep))
}
