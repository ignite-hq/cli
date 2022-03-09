package cosmoserror

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternal       = errors.New("internal error")
	ErrInvalidRequest = errors.New("invalid request")
)

func Unwrap(err error) error {
	s, ok := status.FromError(err)
	if ok {
		switch s.Code() {
		case codes.InvalidArgument:
			return ErrInvalidRequest
		case codes.Internal:
			return ErrInternal
		}
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		return unwrapped
	}
	return err
}
