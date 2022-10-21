package codec

import (
	"errors"
	"github.com/bbdshow/bkit/util/pomelo/packet"
)

// Codec constants.
const (
	HeadLength    = 4
	MaxPacketSize = 1 << 24 //16MB
)

// ErrPacketSizeExceed is the error used for encode/decode.
var ErrPacketSizeExceed = errors.New("codec: packet size exceed")

// ParseHeader parses a packet header and returns its dataLen and packetType or an error
func ParseHeader(header []byte) (int, packet.Type, error) {
	if len(header) != HeadLength {
		return 0, 0x00, packet.ErrInvalidHeader
	}
	typ := header[0]
	if typ < packet.Handshake || typ > packet.Kick {
		return 0, 0x00, packet.ErrWrongPacketType
	}

	size := BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, 0x00, ErrPacketSizeExceed
	}

	return size, packet.Type(typ), nil
}

// BytesToInt decode packet data length byte to int(Big end)
func BytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// IntToBytes encode packet data length to bytes(Big end)
func IntToBytes(n int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((n >> 16) & 0xFF)
	buf[1] = byte((n >> 8) & 0xFF)
	buf[2] = byte(n & 0xFF)
	return buf
}
