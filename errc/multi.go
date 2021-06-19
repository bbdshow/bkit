package errc

import (
	"errors"
	"strings"
)

func MultiError(err ...error) error {
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
