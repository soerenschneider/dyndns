package server

import (
	"testing"
	"time"

	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/events"
	"github.com/soerenschneider/dyndns/internal/verification"
	"github.com/soerenschneider/dyndns/server/dns"
)

type SimpleVerifier struct {
	verificationResult bool
}

func (s SimpleVerifier) Verify(signature string, ip common.DnsRecord) bool {
	return s.verificationResult
}

func TestServer_verifyMessage(t *testing.T) {
	type fields struct {
		knownHosts map[string][]verification.VerificationKey
		listener   events.EventListener
		requests   chan common.UpdateRecordRequest
		propagator dns.Propagator
		cache      map[string]common.DnsRecord
	}
	type args struct {
		env common.UpdateRecordRequest
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
				listener:   nil,
				requests:   nil,
				propagator: nil,
				cache:      map[string]common.DnsRecord{},
			},
			args: args{
				env: common.UpdateRecordRequest{
					PublicIp: common.DnsRecord{
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
				listener:   nil,
				requests:   nil,
				propagator: nil,
				cache:      map[string]common.DnsRecord{},
			},
			args: args{
				env: common.UpdateRecordRequest{
					PublicIp: common.DnsRecord{
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
				listener:   nil,
				requests:   nil,
				propagator: nil,
				cache:      map[string]common.DnsRecord{},
			},
			args: args{
				env: common.UpdateRecordRequest{
					PublicIp: common.DnsRecord{
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
			server := &Server{
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
