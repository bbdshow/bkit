package codec

import (
	"github.com/bbdshow/bkit/util/pomelo/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func toBytes(packetType packet.Type, length, data []byte) []byte {
	if data == nil {
		return nil
	}

	bytes := make([]byte, 0)
	bytes = append(bytes, byte(packetType))
	bytes = append(bytes, length...)
	bytes = append(bytes, data...)
	return bytes
}

var tooBigData = make([]byte, 1<<25)

var encodeTables = map[string]struct {
	packetType packet.Type
	length     []byte
	data       []byte
	err        error
}{
	"test_encode_handshake":    {packet.Handshake, []byte{0x00, 0x00, 0x02}, []byte{0x01, 0x00}, nil},
	"test_invalid_packet_type": {0xff, nil, nil, packet.ErrWrongPacketType},
	"test_too_big_packet":      {packet.Data, nil, tooBigData, ErrPacketSizeExceed},
}

func TestEncode(t *testing.T) {
	t.Parallel()

	for name, table := range encodeTables {
		t.Run(name, func(t *testing.T) {
			pe := NewPomeloEncoder()

			encoded, err := pe.Encode(table.packetType, table.data)
			if table.err != nil {
				assert.Equal(t, table.err, err)
			} else {
				expectedEncoded := toBytes(table.packetType, table.length, table.data)
				assert.Equal(t, expectedEncoded, encoded)
			}
		})
	}
}
