package typ

import (
	"context"
	"github.com/bbdshow/bkit/errc"
	"github.com/bbdshow/bkit/logs"
	"github.com/bbdshow/qelog/api/types"
	"go.uber.org/zap"
	"time"
)

type ListResp struct {
	Count int64       `json:"count"`
	List  interface{} `json:"list"`
}

type LoginResp struct {
	Token    string `json:"token"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

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

func (b *BaseResp) WriteTraceID(ctx context.Context) *BaseResp {
	tid := logs.Qezap.TraceID(ctx)
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

type ReqData struct {
	TraceId  types.TraceID
	Route    string
	CallFunc string
	ClientIP string
	Body     string
	Error    string
}

func SetReqDataContext(ctx context.Context, data ReqData) context.Context {
	return context.WithValue(ctx, "bkit_ReqDataContext", data)
}

func GetReqDataContext(ctx context.Context) ReqData {
	v := ctx.Value("bkit_ReqDataContext")
	if v != nil {
		vv, ok := v.(ReqData)
		if ok {
			return vv
		}
	}
	return ReqData{}
}

func Resp(data interface{}, err error, ctx ...context.Context) *DataResp {
	reqData := ReqData{}
	if len(ctx) > 0 {
		c := ctx[0]
		reqData = GetReqDataContext(c)
		reqData.TraceId = logs.Qezap.TraceID(c)
	}
	code := errc.Success
	message := errc.GetMessage(code)
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
			reqData.Error = message
			logs.Qezap.Error("InternalException", zap.Any("reqData", reqData),
				logs.Qezap.ConditionOne(reqData.Route),
				logs.Qezap.ConditionTwo(reqData.CallFunc),
				logs.Qezap.ConditionThree(reqData.ClientIP),
				zap.String(types.EncoderTraceIDKey, reqData.TraceId.Hex()),
				zap.String("latency", time.Now().Sub(reqData.TraceId.Time()).String()))
			// hide system error
			message = errc.GetMessage(code)
		}
	}
	baseResp := NewBaseResp(code, message)
	baseResp.TraceID = reqData.TraceId.Hex()
	out := NewDataResp(baseResp, data)
	return out
}
