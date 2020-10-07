package glloq

import "errors"

var (
	ErrUnsupportedDSN = errors.New("glloq: DSN unknown or unsupported")
	ErrDSNNotSet      = errors.New("glloq: GLLOQ_DSN environment is not set")
	ErrTimeout        = errors.New("glloq: timeout")
)
