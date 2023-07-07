package key_provider

import (
	"fmt"
	"io"
	"os"
)

type FileProvider struct {
	path string
}

func NewFileProvider(path string) (*FileProvider, error) {
	return &FileProvider{path: path}, nil
}

func (p *FileProvider) Reader() (io.ReadCloser, error) {
	keyFile, err := os.Open(p.path)
	if err != nil {
		return nil, fmt.Errorf("could not read key from file: %w", err)
	}

	return keyFile, nil
}

func (p *FileProvider) CanWrite() bool {
	return true
}

func (p *FileProvider) Write(data []byte) error {
	if err := os.WriteFile(p.path, data, 0600); err != nil {
		return fmt.Errorf("can not key to path %s: %w", p.path, err)
	}

	return nil
}
