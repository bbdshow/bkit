package bkit

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

var processFlag = func() [4]byte {
	var b [4]byte
	io.ReadFull(rand.Reader, b[:])
	return b
}()

var NilRequestID RequestID

type RequestID [12]byte

func NewRequestID() RequestID {
	var b [12]byte
	binary.BigEndian.PutUint64(b[0:8], uint64(time.Now().UnixNano()))
	copy(b[8:12], processFlag[:])
	return b
}

func (id RequestID) Time() time.Time {
	nSec := binary.BigEndian.Uint64(id[0:8])
	return time.Unix(0, int64(nSec))
}
func (id RequestID) String() string {
	return id.Hex()
}

func (id RequestID) Hex() string {
	return hex.EncodeToString(id[:])
}

func (id RequestID) IsZero() bool {
	return bytes.Equal(id[:], NilRequestID[:])
}

func (RequestID) FromHex(s string) (RequestID, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return NilRequestID, err
	}
	if len(b) != 12 {
		return NilRequestID, fmt.Errorf("invalid request id length %d", len(b))
	}
	var id RequestID
	copy(id[:], b[:])
	return id, nil
}
