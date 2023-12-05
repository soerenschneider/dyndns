package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"go.uber.org/multierr"
)

type HttpServer struct {
	address  string
	requests chan common.UpdateRecordRequest

	// optional
	certFile string
	keyFile  string
}

type WebhookOpts func(*HttpServer) error

func New(address string, requestsChan chan common.UpdateRecordRequest, opts ...WebhookOpts) (*HttpServer, error) {
	if len(address) == 0 {
		return nil, errors.New("empty address provided")
	}

	w := &HttpServer{
		address: address,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(w); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return w, errs
}

func (s *HttpServer) IsTLSConfigured() bool {
	return len(s.certFile) > 0 && len(s.keyFile) > 0
}

func (s *HttpServer) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	defer r.Body.Close()

	payload := common.UpdateRecordRequest{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	s.requests <- payload
	w.WriteHeader(http.StatusOK)
}

func (s *HttpServer) Listen(ctx context.Context, events chan bool, wg *sync.WaitGroup) error {
	wg.Add(1)

	mux := http.NewServeMux()
	mux.HandleFunc("/update", s.handle)

	server := http.Server{
		Addr:              s.address,
		Handler:           mux,
		ReadTimeout:       3 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	errChan := make(chan error)
	go func() {
		if s.IsTLSConfigured() {
			if err := server.ListenAndServeTLS(s.certFile, s.keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- fmt.Errorf("can not start webhook server: %w", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- fmt.Errorf("can not start webhook server: %w", err)
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("Stopping webhook server")
		err := server.Shutdown(ctx)
		wg.Done()
		return err
	case err := <-errChan:
		return err
	}
}
