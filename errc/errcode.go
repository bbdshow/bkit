package errc

import (
	"fmt"
	"runtime"
)

// -1--999 system common error code
const (
	Failed  = -1
	Success = 0

	InternalErr = 1

	AuthInternalErr      = 10
	AuthRequired         = 11
	AuthExpired          = 12
	AuthInvalid          = 13
	AuthSignatureInvalid = 14

	ParamRequired = 20
	ParamInvalid  = 21

	NotFound = 404
)

var Lang = "EN"

func SetLang(lang string) {
	Lang = lang
}

var MessagesCN = map[int]string{
	Failed:      "失败",
	Success:     "成功",
	InternalErr: "内部错误",

	AuthInternalErr:      "权限内部错误",
	AuthRequired:         "权限请求头必传",
	AuthExpired:          "权限过期",
	AuthInvalid:          "权限无效",
	AuthSignatureInvalid: "权限签名无效",

	ParamRequired: "参数必传",
	ParamInvalid:  "参数验证无效",
	NotFound:      "资源不存在",
}

var Messages = map[int]string{
	Failed:               "Failed",
	Success:              "Ok",
	InternalErr:          "Internal error",
	AuthInternalErr:      "Auth internal error",
	AuthRequired:         "Auth Authorization header required",
	AuthExpired:          "Auth expired",
	AuthInvalid:          "Auth invalid",
	AuthSignatureInvalid: "Auth signature invalid",

	ParamRequired: "Param required",
	ParamInvalid:  "Param validator invalid",

	NotFound: "Not found",
}

func GetMessage(code int) string {
	switch Lang {
	case "CN":
		return MessagesCN[code]
	default:
		return Messages[code]
	}
}

var (
	ErrFailed      = NewError(Failed, GetMessage(Failed))
	ErrInternalErr = NewError(InternalErr, GetMessage(InternalErr))

	ErrAuthInternalErr      = NewError(AuthInternalErr, GetMessage(AuthInternalErr))
	ErrAuthRequired         = NewError(AuthRequired, GetMessage(AuthRequired))
	ErrAuthExpired          = NewError(AuthExpired, GetMessage(AuthExpired))
	ErrAuthInvalid          = NewError(AuthInvalid, GetMessage(AuthInvalid))
	ErrAuthSignatureInvalid = NewError(AuthSignatureInvalid, GetMessage(AuthSignatureInvalid))

	ErrParamRequired = NewError(ParamRequired, GetMessage(ParamRequired))
	ErrParamInvalid  = NewError(ParamInvalid, GetMessage(ParamInvalid))

	ErrNotFound = NewError(NotFound, GetMessage(NotFound))
)

type Error struct {
	Code    int      `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`
	Stack   []string `json:"-" xml:"-"`
}

func NewError(code int, msg string) Error {
	return Error{Code: code, Message: msg, Stack: make([]string, 0, 1)}
}

func (e Error) Error() string {
	return fmt.Sprintf("code:%d message:%s statck:%v", e.Code, e.Message, e.Stack)
}

func (e Error) MultiErr(err error) Error {
	if err != nil {
		if e1, ok := err.(Error); ok {
			e.Message += " " + e1.Message
			return e.addStack(e1.Stack...)
		} else {
			e.Message += " " + err.Error()
		}
	}
	return e
}

func (e Error) MultiMsg(msg string) Error {
	if msg != "" {
		e.Message += " " + msg
	}
	return e
}

func (e Error) CodeMsg() (int, string) {
	return e.Code, e.Message
}

func (e Error) addStack(v ...string) Error {
	e.Stack = append(e.Stack, v...)
	return e
}

func Message(msg string) Error {
	return NewError(Failed, msg)
}

func WithStack(err error) error {
	if err == nil {
		return nil
	}
	frames := runtime.CallersFrames(callers())
	stack := ""
	frame, more := frames.Next()
	if more {
		stack = fmt.Sprintf("\n%s:%d %s", frame.File, frame.Line, frame.Function)
	}

	if e, ok := err.(Error); ok {
		return e.addStack(stack)
	}
	return NewError(Failed, err.Error()).addStack(stack)
}

func callers() []uintptr {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}

// ToInternalError if not Error, to internal error
func ToInternalError(err error) error {
	if err == nil {
		return nil
	}
	e, ok := err.(Error)
	if ok {
		return e
	}
	// to internal
	return NewError(InternalErr, err.Error())
}
