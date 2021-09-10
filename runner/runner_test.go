package runner

import (
	"context"
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	ctx := context.Background()
	go func() {
		if err := RunServer(new(EmptyServer), WithContext(ctx)); err != nil {
			t.Fatal(err)
		}
	}()

	if err := RunServer(new(EmptyServer), WithContext(ctx)); err != nil {
		t.Fatal(err)
	}
	fmt.Println("exit")
}
