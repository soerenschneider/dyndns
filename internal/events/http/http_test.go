package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
