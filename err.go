package bkit

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

func ErrMulti(err ...error) error {
	var errs []string
	for _, v := range err {
		if v != nil {
			errs = append(errs, v.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, " "))
	}
	return nil
}

func ErrWithStack(err error) error {
	if err == nil {
		return nil
	}
	frames := runtime.CallersFrames(callers())
	stack := ""
	frame, more := frames.Next()
	if more {
		stack = fmt.Sprintf("\n%s:%d %s", frame.File, frame.Line, frame.Function)
	}
	return fmt.Errorf("%s\nstack:%s", err.Error(), stack)
}

// -1--999 system common error code
type ErrCode int

const (
	ErrCodeOK ErrCode = 0
	// 通用错误码
	ErrCodeFailed   ErrCode = 1
	ErrCodeInternal ErrCode = 2

	// 权限相关-内部定义
	ErrCodeAuthInvalid ErrCode = 401

	// 参数相关-内部定义
	ErrCodeParamInvalid ErrCode = 20

	// 资源相关-内部定义
	ErrCodeNotFound ErrCode = 404
)

var ErrMsgLang = "CN"

func SetErrMsgLang(lang string) {
	ErrMsgLang = lang
}

var ErrMsgCN = map[ErrCode]string{
	ErrCodeFailed:       "失败",
	ErrCodeOK:           "成功",
	ErrCodeInternal:     "内部错误",
	ErrCodeAuthInvalid:  "权限无效",
	ErrCodeParamInvalid: "参数验证无效",
	ErrCodeNotFound:     "资源不存在",
}

var ErrMsg = map[ErrCode]string{
	ErrCodeFailed:       "failed",
	ErrCodeOK:           "ok",
	ErrCodeInternal:     "internal error",
	ErrCodeAuthInvalid:  "auth invalid",
	ErrCodeParamInvalid: "param invalid",
	ErrCodeNotFound:     "not found",
}

func GetErrMsg(code ErrCode) string {
	switch ErrMsgLang {
	case "CN":
		return ErrMsgCN[code]
	default:
		return ErrMsg[code]
	}
}

var (
	ErrFailed       = NewErr(ErrCodeFailed, GetErrMsg(ErrCodeFailed))
	ErrInternal     = NewErr(ErrCodeInternal, GetErrMsg(ErrCodeInternal))
	ErrAuthInvalid  = NewErr(ErrCodeAuthInvalid, GetErrMsg(ErrCodeAuthInvalid))
	ErrParamInvalid = NewErr(ErrCodeParamInvalid, GetErrMsg(ErrCodeParamInvalid))
	ErrNotFound     = NewErr(ErrCodeNotFound, GetErrMsg(ErrCodeNotFound))
)

type Err struct {
	Code    ErrCode  `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`
	Stack   []string `json:"-" xml:"-"`
}

func NewErr(code ErrCode, msg string) Err {
	return Err{Code: code, Message: msg, Stack: make([]string, 0, 1)}
}

func (e Err) Error() string {
	return fmt.Sprintf("code:%d message:%s stack:%v", e.Code, e.Message, e.Stack)
}

func (e Err) ErrMulti(err error) Err {
	if err != nil {
		if e1, ok := err.(Err); ok {
			e.Message += " " + e1.Message
			return e.addStack(e1.Stack...)
		} else {
			e.Message += " " + err.Error()
		}
	}
	return e
}

func (e Err) MsgMulti(format string, v ...interface{}) Err {
	msg := fmt.Sprintf(format, v...)
	if msg != "" {
		e.Message += " " + msg
	}
	return e
}

func (e Err) CodeMsg() (ErrCode, string) {
	return e.Code, e.Message
}

func (e Err) addStack(v ...string) Err {
	e.Stack = append(e.Stack, v...)
	return e
}

func NewErrMsg(msg string) Err {
	return NewErr(ErrCodeFailed, msg)
}

func callers() []uintptr {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}

// ErrToInternal if not Err, to internal error
func ErrToInternal(err error) error {
	if err == nil {
		return nil
	}
	e, ok := err.(Err)
	if ok {
		return e
	}
	// to internal
	return NewErr(ErrCodeInternal, err.Error())
}
