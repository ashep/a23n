package api

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
)

type ErrInvalidArg struct {
	Msg string
}

func (e ErrInvalidArg) Error() string {
	return e.Msg
}

func (e ErrInvalidArg) Is(err error) bool {
	_, ok := err.(ErrInvalidArg)
	return ok
}
