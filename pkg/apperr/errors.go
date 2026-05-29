package apperr

import (
	"fmt"
	"strings"
)

type Field struct {
	Name  string
	Value any
}

type Frame struct {
	Op     string
	Fields []Field
}

type Error struct {
	Frame Frame
	Err   error
	Msg   string
}

func (e *Error) Error() string {
	return Format(e)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Wrap(op string, msg string, err error, fields ...Field) error {
	if err == nil {
		return nil
	}

	return &Error{
		Frame: Frame{
			Op:     op,
			Fields: fields,
		},
		Err: err,
		Msg: msg,
	}
}

func Format(err error) string {
	var ops []string
	var fields []Field
	var msgs []string
	var cause error

	for err != nil {
		appErr, ok := err.(*Error)
		if !ok {
			cause = err
			break
		}

		if appErr.Frame.Op != "" {
			ops = append(ops, appErr.Frame.Op)
		}
		fields = append(fields, appErr.Frame.Fields...)
		if appErr.Msg != "" {
			msgs = append(msgs, appErr.Msg)
		}

		err = appErr.Err
	}

	var parts []string
	if len(ops) > 0 {
		parts = append(parts, fmt.Sprintf("[%s]", strings.Join(ops, " -> ")))
	}
	if len(fields) > 0 {
		parts = append(parts, fmt.Sprintf("{%s}", formatFields(fields)))
	}
	if len(msgs) > 0 {
		parts = append(parts, strings.Join(msgs, ": "))
	}
	if cause != nil {
		parts = append(parts, cause.Error())
	}

	return strings.Join(parts, ": ")
}

func Fields(err error) []Field {
	var fields []Field

	for err != nil {
		appErr, ok := err.(*Error)
		if !ok {
			break
		}

		fields = append(fields, appErr.Frame.Fields...)
		err = appErr.Err
	}

	return fields
}

func Ops(err error) []string {
	var ops []string

	for err != nil {
		appErr, ok := err.(*Error)
		if !ok {
			break
		}

		if appErr.Frame.Op != "" {
			ops = append(ops, appErr.Frame.Op)
		}
		err = appErr.Err
	}

	return ops
}

func formatFields(fields []Field) string {
	formatted := make([]string, 0, len(fields))
	for _, field := range fields {
		formatted = append(formatted, fmt.Sprintf("%s=%v", field.Name, field.Value))
	}
	return strings.Join(formatted, ", ")
}
