package errs

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func ToStatus(err error) error {
	if err == nil {
		return nil
	}
	switch KindOf(err) {
	case Invalid:
		return status.Error(codes.InvalidArgument, "Invalid request")
	case NotFound:
		return status.Error(codes.NotFound, "Not found")
	case AlreadyExists:
		return status.Error(codes.AlreadyExists, "Already exists")
	case Unauthenticated:
		return status.Error(codes.Unauthenticated, "Unauthenticated")
	case PermissionDenied:
		return status.Error(codes.PermissionDenied, "Permission denied")
	case Conflict:
		return status.Error(codes.Aborted, "Conflict")
	case Timeout:
		return status.Error(codes.DeadlineExceeded, "Timeout")
	case Unavailable:
		return status.Error(codes.Unavailable, "Service unavailable")
	default:
		return status.Error(codes.Internal, "Internal error")
	}
}
