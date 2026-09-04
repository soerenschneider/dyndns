package server

import (
	"testing"
	"time"

	"github.com/soerenschneider/dyndns/v2/internal/verification"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type SimpleVerifier struct {
	verificationResult bool
}

func (s SimpleVerifier) Verify(signature string, ip update.DnsRecord) bool {
	return s.verificationResult
}

func TestServer_verifyMessage(t *testing.T) {
	type fields struct {
		knownHosts map[string][]verification.VerificationKey
		requests   chan update.UpdateRecordRequest
		propagator propagator
		cache      map[string]update.DnsRecord
	}
	type args struct {
		env update.UpdateRecordRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				knownHosts: map[string][]verification.VerificationKey{
					"my-host.tld": []verification.VerificationKey{
						&SimpleVerifier{false},
						&SimpleVerifier{true},
					},
					"other-host.tld": []verification.VerificationKey{
						&SimpleVerifier{false},
						&SimpleVerifier{true},
					},
				},
				requests:   nil,
				propagator: nil,
				cache:      map[string]update.DnsRecord{},
			},
			args: args{
				env: update.UpdateRecordRequest{
					PublicIp: update.DnsRecord{
						IpV4:      "8.8.4.4",
						Host:      "my-host.tld",
						Timestamp: time.Now(),
					},
					Signature: "dummy-value",
				},
			},
			wantErr: false,
		},

		{
			name: "validation not successful",
			fields: fields{
				knownHosts: map[string][]verification.VerificationKey{
					"my-host.tld": []verification.VerificationKey{
						&SimpleVerifier{false},
					},
					"other-host.tld": []verification.VerificationKey{
						&SimpleVerifier{false},
					},
				},
				requests:   nil,
				propagator: nil,
				cache:      map[string]update.DnsRecord{},
			},
			args: args{
				env: update.UpdateRecordRequest{
					PublicIp: update.DnsRecord{
						IpV4:      "8.8.4.4",
						Host:      "my-host.tld",
						Timestamp: time.Now(),
					},
					Signature: "dummy-value",
				},
			},
			wantErr: true,
		},

		{
			name: "ho host",
			fields: fields{
				knownHosts: map[string][]verification.VerificationKey{
					"my-host.tld": []verification.VerificationKey{
						&SimpleVerifier{false},
					},
					"other-host.tld": []verification.VerificationKey{
						&SimpleVerifier{false},
					},
				},
				requests:   nil,
				propagator: nil,
				cache:      map[string]update.DnsRecord{},
			},
			args: args{
				env: update.UpdateRecordRequest{
					PublicIp: update.DnsRecord{
						IpV4:      "8.8.4.4",
						Host:      "not-found.tld",
						Timestamp: time.Now(),
					},
					Signature: "dummy-value",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &DyndnsServer{
				knownHosts: tt.fields.knownHosts,
				requests:   tt.fields.requests,
				propagator: tt.fields.propagator,
				cache:      tt.fields.cache,
			}
			if err := server.verifyMessage(tt.args.env); (err != nil) != tt.wantErr {
				t.Errorf("verifyMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
