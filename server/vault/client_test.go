package vault

import (
	"github.com/soerenschneider/dyndns/conf"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestConvertCredentials(t *testing.T) {
	tests := []struct {
		name string
		args Credentials
		want credentials.Value
	}{
		{
			args: Credentials{
				AccessKey:     "access key",
				SecretKey:     "secret key",
				SecurityToken: "security token",
			},
			want: credentials.Value{
				AccessKeyID:     "access key",
				SecretAccessKey: "secret key",
				SessionToken:    "security token",
				ProviderName:    "vault",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertCredentials(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVaultCredentialProvider_parseAwsCredentialsReply(t *testing.T) {
	type fields struct {
		vaultToken string
		client     *http.Client
		expiry     time.Time
		conf       *conf.VaultConfig
	}
	tests := []struct {
		name    string
		fields  fields
		body    []byte
		want    *VaultCredentialResponse
		wantErr bool
	}{
		{
			name:    "empty",
			wantErr: true,
			want:    nil,
		},
		{
			wantErr: false,
			body: []byte(`{
  "auth": {
    "renewable": true,
    "lease_duration": 1200,
    "metadata": null,
    "token_policies": ["default"],
    "accessor": "fd6c9a00-d2dc-3b11-0be5-af7ae0e1d374",
    "client_token": "5b1a0318-679c-9c45-e5c6-d1b9a9035d49"
  },
  "warnings": null,
  "wrap_info": null,
  "data": null,
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}`),
			want: &VaultCredentialResponse{
				Data: Credentials{},
				Auth: AuthReply{
					Renewable:     true,
					LeaseDuration: 1200,
					Metadata:      nil,
					TokenPolicies: []string{"default"},
					Accessor:      "fd6c9a00-d2dc-3b11-0be5-af7ae0e1d374",
					ClientToken:   "5b1a0318-679c-9c45-e5c6-d1b9a9035d49",
				},
				Renewable:     false,
				LeaseDuration: 0,
				Warnings:      nil,
				RequestId:     "",
				LeaseId:       "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &VaultCredentialProvider{
				vaultToken: tt.fields.vaultToken,
				client:     tt.fields.client,
				expiry:     tt.fields.expiry,
				conf:       tt.fields.conf,
			}
			got, err := m.parseAwsCredentialsReply(tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAwsCredentialsReply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAwsCredentialsReply() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVaultCredentialProvider_parseLookupReply(t *testing.T) {
	expiry, _ := time.Parse(time.RFC3339, "2018-05-19T11:35:54.466476215-04:00")
	issue, _ := time.Parse(time.RFC3339, "2018-04-17T11:35:54.466476078-04:00")

	type fields struct {
		vaultToken string
		client     *http.Client
		expiry     time.Time
		conf       *conf.VaultConfig
	}
	tests := []struct {
		name    string
		fields  fields
		body    []byte
		want    *TokenInfo
		wantErr bool
	}{
		{
			name:    "empty",
			wantErr: true,
			want:    nil,
		},
		{
			body: []byte(`{
  "data": {
    "accessor": "8609694a-cdbc-db9b-d345-e782dbb562ed",
    "creation_time": 1523979354,
    "creation_ttl": 2764800,
    "display_name": "ldap2-tesla",
    "entity_id": "7d2e3179-f69b-450c-7179-ac8ee8bd8ca9",
    "expire_time": "2018-05-19T11:35:54.466476215-04:00",
    "explicit_max_ttl": 0,
    "id": "cf64a70f-3a12-3f6c-791d-6cef6d390eed",
    "identity_policies": ["dev-group-policy"],
    "issue_time": "2018-04-17T11:35:54.466476078-04:00",
    "meta": {
      "username": "tesla"
    },
    "num_uses": 0,
    "orphan": true,
    "path": "auth/ldap2/login/tesla",
    "policies": ["default", "testgroup2-policy"],
    "renewable": true,
    "ttl": 2764790
  }
}`),
			wantErr: false,
			want: &TokenInfo{
				IssueTime:   issue,
				ExpireTime:  expiry,
				DisplayName: "ldap2-tesla",
				Policies:    []string{"default", "testgroup2-policy"},
				Metadata:    map[string]string{"username": "tesla"},
				NumUses:     0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &VaultCredentialProvider{
				vaultToken: tt.fields.vaultToken,
				client:     tt.fields.client,
				expiry:     tt.fields.expiry,
				conf:       tt.fields.conf,
			}
			got, err := m.parseLookupReply(tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLookupReply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLookupReply() got = %v, want %v", got, tt.want)
			}
		})
	}
}
