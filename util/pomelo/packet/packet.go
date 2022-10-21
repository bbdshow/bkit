package packet

import "fmt"

type Type byte

type Packet struct {
	Type   Type
	Length int
	Data   []byte
}

func New() *Packet {
	return &Packet{}
}

func (p *Packet) String() string {
	return fmt.Sprintf("Type: %d, Length: %d, Data: %s", p.Type, p.Length, string(p.Data))
}
