package bkit

import "github.com/jinzhu/copier"

// Option sets copy options
type CopyOption struct {
	// setting this value to true will ignore copying zero values of all the fields, including bools, as well as a
	// struct having all it's fields set to their zero values respectively (see IsZero() in reflect/value.go)
	IgnoreEmpty bool
	DeepCopy    bool
}

var (
	ErrCopyInvalidDestination = copier.ErrInvalidCopyDestination
	ErrCopyInvalidFrom        = copier.ErrInvalidCopyFrom
	ErrCopyMapKeyNotMatch     = copier.ErrMapKeyNotMatch
	ErrCopyNotSupported       = copier.ErrNotSupported
)

// Copy copies data from src to dst. It uses github.com/jinzhu/copier under the hood.
func Copy(to interface{}, from interface{}) error {
	return copier.Copy(to, from)
}

func CopyWithOption(to, from interface{}, opt CopyOption) error {
	return copier.CopyWithOption(to, from, copier.Option{
		IgnoreEmpty: opt.IgnoreEmpty,
		DeepCopy:    opt.DeepCopy,
	})
}
