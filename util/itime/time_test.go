package itime

import (
	"context"
	"testing"
	"time"
)

func TestCtxAfterSecDeadline(t *testing.T) {
	ctxTimeout, _ := context.WithTimeout(context.Background(), 5*time.Second)
	v := CtxAfterSecDeadline(ctxTimeout, 25)
	t.Log(v.String())

	v1 := CtxAfterSecDeadline(context.Background(), 25)
	t.Log(v1.String())
}
