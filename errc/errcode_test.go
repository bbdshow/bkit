package errc

import (
	"fmt"
	"testing"
)

func TestWithStack(t *testing.T) {
	var err error
	if e := retError(); e != nil {
		err = WithStack(Message("wrap error").MultiErr(e))
	}
	err = WithStack(err)

	fmt.Println(err)
}

func retError() error {
	return WithStack(fmt.Errorf("1"))
}

func TestError_MultiMsg(t *testing.T) {
	msg := ErrNotFound.MultiMsg("arg ", 1, nil)
	fmt.Println(msg.Error())
}
