package errs

import (
	"errors"
	"fmt"
)

type Kind string

const (
	Invalid          Kind = "invalid"
	NotFound         Kind = "not_found"
	AlreadyExists    Kind = "already_exists"
	Unauthenticated  Kind = "unauthenticated"
	PermissionDenied Kind = "permission_denied"
	Conflict         Kind = "conflict"
	Unavailable      Kind = "unavailable"
	Timeout          Kind = "timeout"
	Internal         Kind = "internal"
)

type E struct {
	Op   string
	Kind Kind
	Err  error
}

func (e *E) Error() string {
	switch {
	case e == nil:
		return "<nil>"
	case e.Op != "" && e.Err != nil:
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	case e.Op != "":
		return e.Op
	case e.Err != nil:
		return e.Err.Error()
	default:
		return string(e.Kind)
	}
}

func (e *E) Unwrap() error { return e.Err }

func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	var ex *E
	if errors.As(err, &ex) {
		return &E{Op: op, Kind: ex.Kind, Err: err}
	}
	return &E{Op: op, Err: err}
}

func WithKind(op string, k Kind, err error) error {
	return &E{Op: op, Kind: k, Err: err}
}

func KindOf(err error) Kind {
	var ex *E
	if errors.As(err, &ex) {
		if ex.Kind != "" {
			return ex.Kind
		}
	}
	return Internal
}
