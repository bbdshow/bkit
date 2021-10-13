package icopier

import "github.com/jinzhu/copier"

// Option sets copy options
type Option struct {
	// setting this value to true will ignore copying zero values of all the fields, including bools, as well as a
	// struct having all it's fields set to their zero values respectively (see IsZero() in reflect/value.go)
	IgnoreEmpty bool
	DeepCopy    bool
}

var (
	ErrInvalidCopyDestination = copier.ErrInvalidCopyDestination
	ErrInvalidCopyFrom        = copier.ErrInvalidCopyFrom
	ErrMapKeyNotMatch         = copier.ErrMapKeyNotMatch
	ErrNotSupported           = copier.ErrNotSupported
)

func Copy(to interface{}, from interface{}) error {
	return copier.Copy(to, from)
}

func CopyWithOption(to, from interface{}, opt Option) error {
	return copier.CopyWithOption(to, from, copier.Option{
		IgnoreEmpty: opt.IgnoreEmpty,
		DeepCopy:    opt.DeepCopy,
	})
}
