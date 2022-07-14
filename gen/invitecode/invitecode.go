package invitecode

import (
	"errors"
	"strings"
)

var (
	base    = "E8S2DZX9WYLTN6BGQF7P5IK3MJUAR4HV"
	decimal = 32
	pad     = "C"
	length  = 6
)

func SetBase(b string) {
	b = strings.ToUpper(strings.TrimSpace(b))
	if len(b) <= 0 {
		return
	}
	decimal = len(base)
	base = b
}

func SetPad(p string) error {
	p = strings.ToUpper(p)
	if strings.Contains(base, p) {
		return errors.New("pad should not exists in base")
	}
	pad = p
	return nil
}

func SetLength(n int) {
	length = n
}

// Encode uid to code
func Encode(uid uint64) string {
	id := uid
	mod := uint64(0)
	res := ""
	for id != 0 {
		mod = id % uint64(decimal)
		id = id / uint64(decimal)
		res += string(base[mod])
	}
	resLen := len(res)
	if resLen < length {
		res += pad
		for i := 0; i < length-resLen-1; i++ {
			res += string(base[(int(uid)+i)%decimal])
		}
	}
	return res
}

// Decode code to uid
func Decode(code string) uint64 {
	res := uint64(0)
	lenCode := len(code)
	baseArr := []byte(base)       // string decimal to byte array
	baseRev := make(map[byte]int) // decimal data key to map
	for k, v := range baseArr {
		baseRev[v] = k
	}
	// find cover char addr
	isPad := strings.Index(code, pad)
	if isPad != -1 {
		lenCode = isPad
	}
	r := 0
	for i := 0; i < lenCode; i++ {
		// if cover char , continue
		if string(code[i]) == pad {
			continue
		}
		index, ok := baseRev[code[i]]
		if !ok {
			return 0
		}
		b := uint64(1)
		for j := 0; j < r; j++ {
			b *= uint64(decimal)
		}
		res += uint64(index) * b
		r++
	}
	return res
}
