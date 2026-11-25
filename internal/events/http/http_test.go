package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

func TestNew_NilChannel(t *testing.T) {
	if _, err := New(":8080", nil); err == nil {
		t.Fatal("expected error for nil requests channel")
	}
}

func TestHandle(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		body        string
		wantStatus  int
		wantRequest bool
	}{
		{"valid request is forwarded", http.MethodPost, `{"public_ip":{"host":"my.host.tld"}}`, http.StatusOK, true},
		{"wrong method is rejected", http.MethodGet, `{"public_ip":{"host":"my.host.tld"}}`, http.StatusMethodNotAllowed, false},
		{"invalid json is rejected", http.MethodPost, `not json`, http.StatusBadRequest, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := make(chan update.UpdateRecordRequest, 1)
			s, err := New(":8080", requests)
			if err != nil {
				t.Fatal(err)
			}

			rec := httptest.NewRecorder()
			s.handle(rec, httptest.NewRequest(tt.method, "/update", strings.NewReader(tt.body)))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if gotRequest := len(requests) == 1; gotRequest != tt.wantRequest {
				t.Errorf("request forwarded = %v, want %v", gotRequest, tt.wantRequest)
			}
		})
	}
}

// Listen used to pass the already cancelled ctx to Shutdown, which then
// returned its error while a request was still running, and the server exited
// with a fatal error.
func TestListen_ShutsDownCleanlyWithRunningRequest(t *testing.T) {
	orig := shutdownTimeout
	shutdownTimeout = 100 * time.Millisecond
	t.Cleanup(func() { shutdownTimeout = orig })

	addr := freeAddr(t)
	// Nobody reads the requests, so the handler blocks and the request keeps running.
	s, err := New(addr, make(chan update.UpdateRecordRequest))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Listen(ctx, &sync.WaitGroup{})
	}()

	go func() {
		body := strings.NewReader(`{"public_ip":{"host":"my.host.tld"}}`)
		for i := 0; i < 50; i++ {
			resp, err := http.Post("http://"+addr+"/update", "application/json", body)
			if err == nil {
				_ = resp.Body.Close()
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()
	time.Sleep(300 * time.Millisecond)
	cancel()

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Listen() error = %v, want nil on shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Listen() did not return after ctx was cancelled")
	}
}

func freeAddr(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}
