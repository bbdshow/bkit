package packet

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPacket(t *testing.T) {
	p := New()
	assert.NotNil(t, p)
}

func TestString(t *testing.T) {
	tables := []struct {
		tp     Type
		data   []byte
		strOut string
	}{
		{Handshake, []byte{0x01}, fmt.Sprintf("Type: %d, Length: %d, Data: %s", Handshake, 1, string([]byte{0x01}))},
		{Data, []byte{0x01, 0x02, 0x03}, fmt.Sprintf("Type: %d, Length: %d, Data: %s", Data, 3, string([]byte{0x01, 0x02, 0x03}))},
		{Kick, []byte{0x05, 0x02, 0x03, 0x04}, fmt.Sprintf("Type: %d, Length: %d, Data: %s", Kick, 4, string([]byte{0x05, 0x02, 0x03, 0x04}))},
	}

	for _, table := range tables {
		t.Run(string(table.data), func(t *testing.T) {
			p := &Packet{}
			p.Data = table.data
			p.Type = table.tp
			p.Length = len(table.data)
			assert.Equal(t, table.strOut, p.String())
		})
	}
}
