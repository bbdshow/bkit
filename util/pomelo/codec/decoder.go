package codec

import (
	"bytes"
	"github.com/bbdshow/bkit/util/pomelo/packet"
)

type Decoder interface {
	Decode(data []byte) ([]*packet.Packet, error)
}

type PomeloDecoder struct {
}

func NewPomeloDecoder() *PomeloDecoder {
	return &PomeloDecoder{}
}

func (pd *PomeloDecoder) Decode(data []byte) ([]*packet.Packet, error) {
	buf := bytes.NewBuffer(nil)
	buf.Write(data)

	var (
		packets []*packet.Packet
		err     error
	)
	// check length
	if buf.Len() < HeadLength {
		return nil, nil
	}

	// first time
	size, typ, err := pd.forward(buf)
	if err != nil {
		return nil, err
	}

	for size <= buf.Len() {
		// 取size长度，完整包
		p := &packet.Packet{Type: typ, Length: size, Data: buf.Next(size)}
		packets = append(packets, p)

		// 查看剩余未读取的，是否还有其他包，if no more packets, break
		if buf.Len() < HeadLength {
			break
		}
		// 再去看下一个包， 循环直至读取到所有的包
		size, typ, err = pd.forward(buf)
		if err != nil {
			return nil, err
		}
	}

	return packets, nil
}

func (pd *PomeloDecoder) forward(buf *bytes.Buffer) (int, packet.Type, error) {
	header := buf.Next(HeadLength)
	return ParseHeader(header)
}
