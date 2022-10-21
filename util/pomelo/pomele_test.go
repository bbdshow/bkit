package pomelo

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestNewConnector(t *testing.T) {
	c := NewConnector()
	assert.NotNil(t, c)
}

type MockHandshakeRespDo struct {
	resp HandshakeResp
}

func (m *MockHandshakeRespDo) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, &m.resp)
	return err
}

func (m *MockHandshakeRespDo) Success() bool {
	return m.resp.Code == 200
}

func (m *MockHandshakeRespDo) HeartbeatSec() int64 {
	return m.resp.Sys.Heartbeat
}

type HandshakeResp struct {
	Code int `json:"code"`
	Sys  struct {
		Heartbeat  int64 `json:"heartbeat"`
		ServerTime int64 `json:"servertime"`
	}
}

func run() *Connector {
	c := NewConnector()
	if err := c.SetHandshakeData(nil); err != nil {
		log.Fatal(err)
	}
	c.SetHandshakeRespDo(&MockHandshakeRespDo{})
	c.Connected(func() {
		log.Println("已连接")
	})
	go func() {
		if err := c.Run("0.0.0.0:48611"); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(2 * time.Second)
	return c
}

func TestConnector_Run(t *testing.T) {
	c := run()
	type in struct {
		Username string `json:"username"`
	}

	c.On("onWelcome", func(data []byte) {
		fmt.Println(string(data))
	})

	b, _ := json.Marshal(in{Username: "pomelo client"})
	if err := c.Request("auth.GateLogin", b, func(data []byte) {
		fmt.Println("GeteLoginResp ", string(data))
	}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Minute)

}
