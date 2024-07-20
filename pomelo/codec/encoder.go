package codec

import "github.com/bbdshow/bkit/util/pomelo/packet"

type Encoder interface {
	Encode(typ packet.Type, data []byte) ([]byte, error)
}

// PomeloEncoder struct
type PomeloEncoder struct {
}

// NewPomeloEncoder ctor
func NewPomeloEncoder() *PomeloEncoder {
	return &PomeloEncoder{}
}

// Encode create a packet.Packet from  the raw bytes slice and then encode to network bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func (pe *PomeloEncoder) Encode(typ packet.Type, data []byte) ([]byte, error) {
	if typ < packet.Handshake || typ > packet.Kick {
		return nil, packet.ErrWrongPacketType
	}

	if len(data) > MaxPacketSize {
		return nil, ErrPacketSizeExceed
	}

	p := &packet.Packet{Type: typ, Length: len(data)}
	buf := make([]byte, p.Length+HeadLength)
	buf[0] = byte(p.Type)

	copy(buf[1:HeadLength], IntToBytes(p.Length))
	copy(buf[HeadLength:], data)

	return buf, nil
}
