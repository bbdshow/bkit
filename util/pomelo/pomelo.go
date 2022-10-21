package pomelo

import (
	"encoding/json"
	"errors"
	"github.com/bbdshow/bkit/util/pomelo/codec"
	"github.com/bbdshow/bkit/util/pomelo/message"
	"github.com/bbdshow/bkit/util/pomelo/packet"
	"log"
	"net"
	"sync"
	"time"
)

// Callback represents the callback type which will be called
// when the correspond events is occurred.
type Callback func(data []byte)

// Connector is a Pomelo client
type Connector struct {
	conn net.Conn // low-level connection

	decoder    codec.Decoder // decoder
	encoder    codec.Encoder // encoder
	msgEncoder message.Encoder

	mid               uint // message id
	muConn            sync.RWMutex
	connecting        bool        // connection status
	die               chan byte   // connector close channel
	chSend            chan []byte // send queue
	connectedCallback func()

	// some packet data
	handshakeData   []byte // handshake data
	handShakeRespDo HandShakeRespDo

	handshakeAckData []byte // handshake ack data
	heartbeatData    []byte // heartbeat data

	// events handler
	muEvents sync.RWMutex
	events   map[string]Callback

	// response handler
	muResponses sync.RWMutex
	responses   map[uint]Callback
}

func NewConnector() *Connector {
	c := &Connector{
		conn:       nil,
		decoder:    codec.NewPomeloDecoder(),
		encoder:    codec.NewPomeloEncoder(),
		msgEncoder: message.NewMessagesEncoder(),
		mid:        1,
		connecting: false,
		die:        make(chan byte, 1),
		chSend:     make(chan []byte, 128),
		events:     make(map[string]Callback),
		responses:  make(map[uint]Callback),
	}
	return c
}

// SetHandshakeData 设置握手协议数据
func (c *Connector) SetHandshakeData(handshake interface{}) error {
	data, err := json.Marshal(handshake)
	if err != nil {
		return err
	}
	// 握手数据
	c.handshakeData, err = c.encoder.Encode(packet.Handshake, data)
	if err != nil {
		return err
	}

	return nil
}

// SetHandshakeAckData 设置握手ACK数据
func (c *Connector) SetHandshakeAckData(handshakeAck interface{}) error {
	var err error
	if handshakeAck == nil {
		c.handshakeAckData, err = c.encoder.Encode(packet.HandshakeAck, nil)
		if err != nil {
			return err
		}
		return nil
	}

	data, err := json.Marshal(handshakeAck)
	if err != nil {
		return err
	}

	c.handshakeAckData, err = c.encoder.Encode(packet.HandshakeAck, data)
	if err != nil {
		return err
	}

	return nil
}

// SetHeartBeatData 设置心跳包数据
func (c *Connector) SetHeartBeatData(heartbeat interface{}) error {
	var err error
	if heartbeat == nil {
		c.heartbeatData, err = c.encoder.Encode(packet.Heartbeat, nil)
		if err != nil {
			return err
		}
		return nil
	}
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	c.heartbeatData, err = c.encoder.Encode(packet.Heartbeat, data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) SetHandshakeRespDo(do HandShakeRespDo) {
	c.handShakeRespDo = do
}

// Connected 连接之后 cb
func (c *Connector) Connected(cb func()) {
	c.connectedCallback = cb
}

// Run 启动连接
func (c *Connector) Run(addr string) error {
	if c.handshakeData == nil {
		return errors.New("handshake not defined")
	}

	if c.handshakeAckData == nil {
		err := c.SetHandshakeAckData(nil)
		if err != nil {
			return err
		}
	}

	if c.heartbeatData == nil {
		err := c.SetHeartBeatData(nil)
		if err != nil {
			return err
		}
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	c.conn = conn
	c.connecting = true

	go c.write()

	// 握手
	c.send(c.handshakeData)

	// 读取数据
	err = c.read()

	return err
}

// Request send a request to server and register a callback for the response
func (c *Connector) Request(route string, data []byte, callback Callback) error {
	msg := &message.Message{
		Type:  message.Request,
		Route: route,
		ID:    c.mid,
		Data:  data,
	}

	c.setResponseHandler(c.mid, callback)
	if err := c.sendMessage(msg); err != nil {
		c.setResponseHandler(c.mid, nil)
		return err
	}

	return nil
}

// Notify send a notification to server
func (c *Connector) Notify(route string, data []byte) error {
	msg := &message.Message{
		Type:  message.Notify,
		Route: route,
		Data:  data,
	}
	return c.sendMessage(msg)
}

// On add the callback for the event
func (c *Connector) On(event string, callback Callback) {
	c.muEvents.Lock()
	defer c.muEvents.Unlock()

	c.events[event] = callback
}

// Close  the connection, and shutdown the benchmark
func (c *Connector) Close() {
	if !c.connecting {
		return
	}
	log.Println("关闭连接")
	c.conn.Close()
	c.die <- 1
	c.connecting = false
}

// IsClosed check the connection is closed
func (c *Connector) IsClosed() bool {
	return !c.connecting
}

func (c *Connector) eventHandler(event string) (Callback, bool) {
	c.muEvents.RLock()
	defer c.muEvents.RUnlock()

	cb, ok := c.events[event]
	return cb, ok
}

func (c *Connector) responseHandler(mid uint) (Callback, bool) {
	c.muResponses.RLock()
	defer c.muResponses.RUnlock()

	cb, ok := c.responses[mid]
	return cb, ok
}

func (c *Connector) setResponseHandler(mid uint, cb Callback) {
	c.muResponses.Lock()
	defer c.muResponses.Unlock()

	if cb == nil {
		delete(c.responses, mid)
	} else {
		c.responses[mid] = cb
	}
}

func (c *Connector) sendMessage(msg *message.Message) error {
	data, err := c.msgEncoder.Encode(msg)
	if err != nil {
		return err
	}

	payload, err := c.encoder.Encode(packet.Data, data)
	if err != nil {
		return err
	}

	c.mid++
	c.send(payload)

	return nil
}

func (c *Connector) write() {
	for {
		select {
		case data := <-c.chSend:
			if _, err := c.conn.Write(data); err != nil {
				log.Println("写消息失败", err.Error(), string(data))
				// c.Close()
			}
		case <-c.die:
			return
		}
	}
}

func (c *Connector) send(data []byte) {
	if len(data) > 0 {
		c.chSend <- data
	}
}

func (c *Connector) read() error {
	for {
		//time.Sleep(time.Second)
		if c.IsClosed() {
			return errors.New("连接已断开")
		}

		b, err := message.GetNextMessage(c.conn)
		if err != nil {
			log.Println("读取数据失败", err.Error())
			c.Close()
			return err
			// continue
		}

		packets, err := c.decoder.Decode(b)
		if err != nil {
			log.Println("解数据失败", err.Error())
			// c.Close()
			// return
			continue
		}

		for i := range packets {
			p := packets[i]
			c.processPacket(p)
		}
	}
}

type HandShakeRespDo interface {
	Unmarshal(data []byte) error
	Success() bool
	HeartbeatSec() int64
}

func (c *Connector) processPacket(p *packet.Packet) {
	//log.Printf("packet: %s \n", p.String())
	switch p.Type {
	case packet.Handshake:
		if c.handShakeRespDo == nil {
			log.Fatal("handShakeRespDo", string(p.Data))
			c.Close()
			return
		}
		if err := c.handShakeRespDo.Unmarshal(p.Data); err != nil {
			log.Fatal("handShakeRespDo.Unmarshal", string(p.Data))
			c.Close()
			return
		}
		if !c.handShakeRespDo.Success() {
			log.Fatal("handShakeRespDo.Failed", string(p.Data))
			c.Close()
			return
		}
		// send heartbeat
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(c.handShakeRespDo.HeartbeatSec()))
			for range ticker.C {
				if c.IsClosed() {
					return
				}
				c.send(c.heartbeatData)
			}
		}()
		// ack
		c.send(c.handshakeAckData)
		if c.connectedCallback != nil {
			c.connectedCallback()
		}
	case packet.Data:
		msg, err := message.Decode(p.Data)
		if err != nil {
			return
		}
		c.processMessage(msg)

	case packet.Kick:
		log.Println("服务器主动断开连接")
		c.Close()
	}
}

func (c *Connector) processMessage(msg *message.Message) {
	switch msg.Type {
	case message.Push:
		//fmt.Println("events", c.events)
		cb, ok := c.eventHandler(msg.Route)
		if !ok {
			log.Println("event handler not found", msg.Route)
			return
		}
		cb(msg.Data)
	case message.Response:
		cb, ok := c.responseHandler(msg.ID)
		if !ok {
			log.Println("response handler not found", msg.ID)
			return
		}

		cb(msg.Data)
		c.setResponseHandler(msg.ID, nil)
	}
}
