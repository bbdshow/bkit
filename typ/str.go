package typ

import (
	"strconv"
	"strings"
)

type StringSplit string

func (ss StringSplit) Unmarshal() []string {
	strs := make([]string, 0)
	for _, s := range strings.Split(string(ss), ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			strs = append(strs, s)
		}
	}
	return strs
}

func (ss StringSplit) Marshal(val []string) StringSplit {
	strs := make([]string, 0)
	for _, s := range val {
		s = strings.TrimSpace(s)
		if s != "" {
			strs = append(strs, s)
		}
	}
	return StringSplit(strings.Join(strs, ","))
}

type IntSplit string

func (is IntSplit) Unmarshal() ([]int, error) {
	ints := make([]int, 0)
	for _, s := range strings.Split(string(is), ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		ints = append(ints, i)
	}
	return ints, nil
}

func TrimStringToInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func TrimStringToFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 10)
}
