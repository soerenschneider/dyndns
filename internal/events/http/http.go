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

type HttpServer struct {
	address  string
	requests chan update.UpdateRecordRequest
}

func New(address string, requestsChan chan update.UpdateRecordRequest) (*HttpServer, error) {
	if len(address) == 0 {
		return nil, errors.New("empty address provided")
	}

	return &HttpServer{
		address: address,
	}, nil
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

	payload := update.UpdateRecordRequest{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	s.requests <- payload
	w.WriteHeader(http.StatusOK)
}

func (s *HttpServer) Listen(ctx context.Context, wg *sync.WaitGroup) error {
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
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("can not start webhook server: %w", err)
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
