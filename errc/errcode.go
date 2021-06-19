package errc

import (
	"fmt"
	"runtime"
)

// -1--999 为系统公共错误
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

var Messages = map[int]string{
	Failed:               "failed",
	Success:              "ok",
	InternalErr:          "internal error",
	AuthInternalErr:      "auth internal error",
	AuthRequired:         "auth Authorization header required",
	AuthExpired:          "auth expired",
	AuthInvalid:          "auth invalid",
	AuthSignatureInvalid: "auth signature invalid",

	ParamInvalid: "param validator invalid",

	NotFound: "not found",
}

var (
	ErrFailed      = NewError(Failed, Messages[Failed])
	ErrInternalErr = NewError(InternalErr, Messages[InternalErr])

	ErrAuthInternalErr      = NewError(AuthInternalErr, Messages[AuthInternalErr])
	ErrAuthRequired         = NewError(AuthRequired, Messages[AuthRequired])
	ErrAuthExpired          = NewError(AuthExpired, Messages[AuthExpired])
	ErrAuthInvalid          = NewError(AuthInvalid, Messages[AuthInvalid])
	ErrAuthSignatureInvalid = NewError(AuthSignatureInvalid, Messages[AuthSignatureInvalid])

	ErrParamInvalid = NewError(ParamInvalid, Messages[ParamInvalid])

	ErrNotFound = NewError(NotFound, Messages[NotFound])
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
