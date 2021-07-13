package ginutil

import (
	"bytes"
	"github.com/bbdshow/bkit/errc"
	"github.com/bbdshow/bkit/logs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"
)

type BaseResp struct {
	errc.Error
	TraceID string `json:"traceId"`
}

func NewBaseResp(code int, msg string) *BaseResp {
	return &BaseResp{
		Error:   errc.NewError(code, msg),
		TraceID: "",
	}
}

func (b *BaseResp) WriteTraceID(c *gin.Context) *BaseResp {
	tid := logs.Qezap.TraceID(c.Request.Context())
	if !tid.IsZero() {
		b.TraceID = tid.Hex()
	}
	return b
}

type DataResp struct {
	*BaseResp
	Data interface{} `json:"data"`
}

func NewDataResp(baseResp *BaseResp, data interface{}) *DataResp {
	return &DataResp{
		BaseResp: baseResp,
		Data:     data,
	}
}

type ResponseErr struct {
	Method   string
	Path     string
	IP       string
	Form     url.Values
	PostForm url.Values
	Func     string
	Error    string
}

func Resp(c *gin.Context, httpCode int, data interface{}, err error) {
	code := errc.Success
	message := errc.Messages[code]
	if err != nil {
		if e, ok := err.(errc.Error); ok {
			code = e.Code
			message = e.Message
		} else {
			code = errc.Failed
			message = err.Error()
		}
		switch code {
		case errc.InternalErr:
			// 拦截响应中间件已经打日志
			respErr := &ResponseErr{
				Method:   c.Request.Method,
				Path:     c.Request.URL.RequestURI(),
				IP:       ClientIP(c),
				Form:     c.Request.Form,
				PostForm: c.Request.PostForm,
				Func:     c.HandlerName(),
				Error:    message,
			}
			traceId := logs.Qezap.TraceID(c.Request.Context())
			logs.Qezap.Error("内部错误", zap.Any("respErr", respErr),
				logs.Qezap.ConditionOne(respErr.Path),
				logs.Qezap.ConditionTwo(respErr.Method),
				logs.Qezap.ConditionThree(respErr.IP),
				logs.Qezap.FieldTraceID(c.Request.Context()),
				zap.String("latency", time.Now().Sub(traceId.Time()).String()))
			// 屏蔽掉系统错误
			message = errc.Messages[code]
		}
	}

	out := NewDataResp(NewBaseResp(code, message).WriteTraceID(c), data)
	c.JSON(httpCode, out)
}

// RespSuccess
func RespSuccess(c *gin.Context, status ...int) {
	Resp(c, httpStatus(status...), nil, nil)
}

func RespErr(c *gin.Context, err error, status ...int) {
	Resp(c, httpStatus(status...), nil, err)
}

// RespData
func RespData(c *gin.Context, data interface{}, status ...int) {
	Resp(c, httpStatus(status...), data, nil)
}

func httpStatus(status ...int) int {
	if len(status) == 1 && status[0] > 0 {
		return status[0]
	}
	return http.StatusOK
}

// respWriter 用于 dump response body 使用
type respWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *respWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
