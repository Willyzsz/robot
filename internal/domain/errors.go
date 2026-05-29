package domain

import "errors"

var (
	ErrEmpty            = errors.New("empty")
	ErrNotEnough        = errors.New("not enough")
	ErrAlreadyExists    = errors.New("already exists")
	ErrNotFound         = errors.New("not found")
	ErrInvalid          = errors.New("invalid")
	ErrInvalidReference = errors.New("invalid reference")
)
