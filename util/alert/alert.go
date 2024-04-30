package alert

import (
	"context"
)

// Alarm interface
type Alarm interface {
	SetProxy(string)
	SetHookURL(string)
	Send(ctx context.Context, content string) error
	Method() string
}
