package key_provider

import (
	"errors"
	"io"
	"strings"
)

type EnvProvider struct {
	data string
}

func NewEnvProvider(data string) (*EnvProvider, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data provided")
	}
	return &EnvProvider{data: data}, nil
}

func (p *EnvProvider) Reader() (io.ReadCloser, error) {
	return &ReaderCloser{stringReader: strings.NewReader(p.data)}, nil
}

func (p *EnvProvider) CanWrite() bool {
	return false
}

func (p *EnvProvider) Write(_ []byte) error {
	return errors.New("unsupported")
}

type ReaderCloser struct {
	stringReader io.Reader
}

func (f *ReaderCloser) Read(n []byte) (int, error) {
	return f.stringReader.Read(n)
}

func (f *ReaderCloser) Close() error {
	return nil
}
