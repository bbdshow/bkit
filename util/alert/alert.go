package alert

import (
	"context"
)

// Alarm
type Alarm interface {
	SetHookURL(string)
	Send(ctx context.Context, content string) error
	Method() string
}
