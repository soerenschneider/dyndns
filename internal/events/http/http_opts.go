package http

import (
	"errors"
)

func WithTLS(certFile, keyFile string) func(s *HttpServer) error {
	return func(s *HttpServer) error {
		if len(certFile) == 0 {
			return errors.New("empty certfile")
		}

		if len(keyFile) == 0 {
			return errors.New("empty keyfile")
		}

		s.certFile = certFile
		s.keyFile = keyFile
		return nil
	}
}
