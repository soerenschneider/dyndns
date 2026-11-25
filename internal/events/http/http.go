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
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

// shutdownTimeout bounds how long Listen waits for running requests; replaced
// in tests.
var shutdownTimeout = 5 * time.Second

type HttpServer struct {
	address  string
	requests chan update.UpdateRecordRequest
}

func New(address string, requestsChan chan update.UpdateRecordRequest) (*HttpServer, error) {
	if len(address) == 0 {
		return nil, errors.New("empty address provided")
	}

	if requestsChan == nil {
		return nil, errors.New("nil requests channel provided")
	}

	return &HttpServer{
		address:  address,
		requests: requestsChan,
	}, nil
}

func (s *HttpServer) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Limit request body to 1 MB (1 << 20 bytes)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	payload := update.UpdateRecordRequest{}
	if err := json.Unmarshal(data, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.requests <- payload
	w.WriteHeader(http.StatusOK)
}

func (s *HttpServer) Listen(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

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
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("can not start webhook server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("Stopping webhook server")
		// ctx is already done, so it cannot bound the shutdown.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Warn().Err(err).Msg("Webhook server did not shut down cleanly")
		}
		return nil
	case err := <-errChan:
		return err
	}
}
