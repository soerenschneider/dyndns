package hybrid

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAwsCredentialsConfig_CredentialsProvider(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secretFile, []byte("from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		conf       AwsCredentialsConfig
		wantNil    bool
		wantSecret string
		wantErr    bool
	}{
		{name: "nothing configured uses the default chain", wantNil: true},
		{name: "inline secret", conf: AwsCredentialsConfig{AccessKeyId: "id", SecretAccessKey: "inline"}, wantSecret: "inline"},
		{name: "secret from file, trailing newline trimmed", conf: AwsCredentialsConfig{AccessKeyId: "id", SecretAccessKeyFile: secretFile}, wantSecret: "from-file"},
		{name: "missing secret file", conf: AwsCredentialsConfig{AccessKeyId: "id", SecretAccessKeyFile: filepath.Join(t.TempDir(), "missing")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := tt.conf.CredentialsProvider()
			if (err != nil) != tt.wantErr {
				t.Fatalf("CredentialsProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if (provider == nil) != tt.wantNil {
				t.Fatalf("CredentialsProvider() = %v, want nil %v", provider, tt.wantNil)
			}
			if tt.wantNil {
				return
			}

			creds, err := provider.Retrieve(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if creds.AccessKeyID != tt.conf.AccessKeyId || creds.SecretAccessKey != tt.wantSecret {
				t.Errorf("Retrieve() = %q/%q, want %q/%q", creds.AccessKeyID, creds.SecretAccessKey, tt.conf.AccessKeyId, tt.wantSecret)
			}
		})
	}
}
