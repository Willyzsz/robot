package domain

import (
	"errors"
	"fmt"
)

var (
	ErrEmpty = errors.New("empty")
	ErrNotEnough = errors.New("not enough")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound = errors.New("not found")
	ErrInvalid = errors.New("invalid")
	ErrInvalidReference = errors.New("invalid reference")
)

type RobotError struct {
	Op	string
	InternalOp string
	Field string
	Value any
	Err error
	Msg string
}

func (e *RobotError) Error() string {
	//[Create->Insert] ID: 1050 invalid. ID must be a number 
	return	fmt.Sprintf("[%s->%s] %s: %v - %s. %s", e.Op, e.InternalOp, e.Field, e.Value, e.Err, e.Msg) 
}

func (e *RobotError) Unwrap() error {
	return e.Err
}

func NewRobotErr(op, intOp, field string, value any, err error, msg string) *RobotError {
	if err == nil {
		return nil
	}

	return &RobotError{
		Op: op,
		InternalOp: intOp,
		Field: field,
		Value: value,
		Err: err,
		Msg: msg,
	}
}