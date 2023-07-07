package key_provider

import (
	"io"
)

type KeyProvider interface {
	Reader() (io.ReadCloser, error)
	Write([]byte) error
	CanWrite() bool
}
