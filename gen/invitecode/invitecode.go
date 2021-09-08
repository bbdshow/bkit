package invitecode

import (
	"strings"
)

const (
	BASE    = "E8S2DZX9WYLTN6BGQF7P5IK3MJUAR4HV"
	DECIMAL = 32
	PAD     = "C"
	LEN     = 6
)

// id to code
func Encode(uid uint64) string {
	id := uid
	mod := uint64(0)
	res := ""
	for id != 0 {
		mod = id % DECIMAL
		id = id / DECIMAL
		res += string(BASE[mod])
	}
	resLen := len(res)
	if resLen < LEN {
		res += PAD
		for i := 0; i < LEN-resLen-1; i++ {
			res += string(BASE[(int(uid)+i)%DECIMAL])
		}
	}
	return res
}

// code to id
func Decode(code string) uint64 {
	res := uint64(0)
	lenCode := len(code)
	baseArr := []byte(BASE)       // string decimal to byte array
	baseRev := make(map[byte]int) // decimal data key to map
	for k, v := range baseArr {
		baseRev[v] = k
	}
	// find cover char addr
	isPad := strings.Index(code, PAD)
	if isPad != -1 {
		lenCode = isPad
	}
	r := 0
	for i := 0; i < lenCode; i++ {
		// if cover char , continue
		if string(code[i]) == PAD {
			continue
		}
		index, ok := baseRev[code[i]]
		if !ok {
			return 0
		}
		b := uint64(1)
		for j := 0; j < r; j++ {
			b *= DECIMAL
		}
		res += uint64(index) * b
		r++
	}
	return res
}
