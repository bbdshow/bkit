package bkit

import (
	"context"
	"testing"
	"time"
)

func TestCtxAfterSecDeadline(t *testing.T) {
	ctxTimeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
	v := Time.CtxAfterSecDeadline(ctxTimeout, 25)
	t.Log(v.String())

	v1 := Time.CtxAfterSecDeadline(context.Background(), 25)
	t.Log(v1.String())
}
