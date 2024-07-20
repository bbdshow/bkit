package codec

import (
	"bytes"
	"github.com/bbdshow/bkit/util/pomelo/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

var forwardTables = map[string]struct {
	buf []byte
	err error
}{
	"test_handshake_type":     {[]byte{packet.Handshake, 0x00, 0x00, 0x00}, nil},
	"test_handshake_ack_type": {[]byte{packet.HandshakeAck, 0x00, 0x00, 0x00}, nil},
	"test_heartbeat_type":     {[]byte{packet.Heartbeat, 0x00, 0x00, 0x00}, nil},
	"test_data_type":          {[]byte{packet.Data, 0x00, 0x00, 0x00}, nil},
	"test_kick_type":          {[]byte{packet.Kick, 0x00, 0x00, 0x00}, nil},

	"test_wrong_packet_type": {[]byte{0x06, 0x00, 0x00, 0x00}, packet.ErrWrongPacketType},
}

var (
	handshakeHeaderPacket = []byte{packet.Handshake, 0x00, 0x00, 0x01, 0x01}
	invalidHeader         = []byte{0xff, 0x00, 0x00, 0x01}
)

var decodeTables = map[string]struct {
	data   []byte
	packet []*packet.Packet
	err    error
}{
	"test_not_enough_bytes": {[]byte{0x01}, nil, nil},
	"test_error_on_forward": {invalidHeader, nil, packet.ErrWrongPacketType},
	"test_forward":          {handshakeHeaderPacket, []*packet.Packet{{packet.Handshake, 1, []byte{0x01}}}, nil},
	"test_forward_many":     {append(handshakeHeaderPacket, handshakeHeaderPacket...), []*packet.Packet{{packet.Handshake, 1, []byte{0x01}}, {packet.Handshake, 1, []byte{0x01}}}, nil},
}

func TestNewPomeloDecoder(t *testing.T) {
	t.Parallel()
	pd := NewPomeloDecoder()
	assert.NotNil(t, pd)
}

func TestForward(t *testing.T) {
	t.Parallel()
	for name, table := range forwardTables {
		t.Run(name, func(t *testing.T) {
			pd := NewPomeloDecoder()

			sz, typ, err := pd.forward(bytes.NewBuffer(table.buf))
			if table.err == nil {
				assert.Equal(t, packet.Type(table.buf[0]), typ)
				assert.Equal(t, 0, sz)
			}
			assert.Equal(t, table.err, err)
		})
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()

	for name, table := range decodeTables {
		t.Run(name, func(t *testing.T) {
			pd := NewPomeloDecoder()

			packet, err := pd.Decode(table.data)
			assert.Equal(t, table.err, err)
			assert.ElementsMatch(t, table.packet, packet)
		})
	}
}
