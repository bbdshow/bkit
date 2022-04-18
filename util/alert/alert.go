package alert

import (
	"context"
)

// Alarm interface
type Alarm interface {
	SetHookURL(string)
	Send(ctx context.Context, content string) error
	Method() string
}
