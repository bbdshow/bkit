package bkit

import (
	"bytes"
	"net/url"

	"github.com/gin-gonic/gin"
)

type BaseResp struct {
	Err
	RequestID string `json:"request_id"`
}

type DataResp struct {
	*BaseResp
	Data interface{} `json:"data"`
}

type CloudResp struct {
	Response interface{} `json:"Response"`
}

type CloudBaseResp struct {
	Code      ErrCode `json:"Code"`
	Msg       string  `json:"Msg"`
	RequestID string  `json:"RequestId"`
}

type CloudDataResp struct {
	*CloudBaseResp
	Data interface{} `json:"Data"`
}

// RespErrInternalLog 返回错误内部日志结构
type RespErrInternalLog struct {
	Method   string
	Path     string
	IP       string
	Form     url.Values
	PostForm url.Values
	Func     string
	Error    string
}

// DumpRespWriter response writer
type DumpRespWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *DumpRespWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type ListResp struct {
	TotalCount int64       `json:"total_count"`
	Data       interface{} `json:"data"`
}
