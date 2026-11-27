package conf

import (
	"strings"
	"testing"

	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
)

func TestIsValidNatsUrl(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid url",
			args: args{
				input: "nats://nats.nats",
			},
			want: true,
		},
		{
			name: "wrong protocol",
			args: args{
				input: "https://nats.nats",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidNatsUrl(tt.args.input); got != tt.want {
				t.Errorf("IsValidNatsUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidNatsSubject(t *testing.T) {
	type args struct {
		subject string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "happy path",
			args: args{
				subject: "this.is.a.test",
			},
			want: true,
		},
		{
			name: "single token",
			args: args{
				subject: "this",
			},
			want: true,
		},
		{
			name: "empty",
			args: args{
				subject: "",
			},
			want: false,
		},
		{
			name: "wildcard >",
			args: args{
				subject: "test.>",
			},
			want: false,
		},
		{
			name: "wildcard *",
			args: args{
				subject: "test.*",
			},
			want: false,
		},
		{
			name: "only dots",
			args: args{
				subject: "..",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidNatsSubject(tt.args.subject); got != tt.want {
				t.Errorf("IsValidNatsSubject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailConfig_Validate(t *testing.T) {
	type fields struct {
		From             string
		FromFile         string
		To               []string
		ToFile           string
		SmtpHost         string
		SmtpPort         int
		SmtpUsername     string
		SmtpUsernameFile string
		SmtpPassword     string
		SmtpPasswordFile string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "valid config, nothing defined",
			fields:  fields{},
			wantErr: false,
		},
		{
			name: "valid config, no files",
			fields: fields{
				From:         "dyndns@yourdomain.tld",
				To:           []string{"you@yourdomain.tld"},
				SmtpHost:     "localhost",
				SmtpPort:     25,
				SmtpUsername: "email",
				SmtpPassword: "secret",
			},
			wantErr: false,
		},
		{
			name: "valid config, use files",
			fields: fields{
				FromFile:         "from.txt",
				ToFile:           "to.txt",
				SmtpHost:         "localhost",
				SmtpPort:         25,
				SmtpUsernameFile: "user.txt",
				SmtpPasswordFile: "secret.txt",
			},
			wantErr: false,
		},
		{
			name: "valid config, mixed usage of files and actual values",
			fields: fields{
				FromFile:         "from.txt",
				To:               []string{"mail@domain.tld"},
				SmtpHost:         "localhost",
				SmtpPort:         25,
				SmtpUsername:     "user",
				SmtpPasswordFile: "secret.txt",
			},
			wantErr: false,
		},
		{
			name: "invalid config, missing from",
			fields: fields{
				ToFile:           "to.txt",
				SmtpHost:         "localhost",
				SmtpPort:         25,
				SmtpUsernameFile: "user.txt",
				SmtpPasswordFile: "secret.txt",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := hybrid.EmailConfig{
				From:             tt.fields.From,
				FromFile:         tt.fields.FromFile,
				To:               tt.fields.To,
				ToFile:           tt.fields.ToFile,
				SmtpHost:         tt.fields.SmtpHost,
				SmtpPort:         tt.fields.SmtpPort,
				SmtpUsername:     tt.fields.SmtpUsername,
				SmtpUsernameFile: tt.fields.SmtpUsernameFile,
				SmtpPassword:     tt.fields.SmtpPassword,
				SmtpPasswordFile: tt.fields.SmtpPasswordFile,
			}

			if err := ValidateConfig[hybrid.EmailConfig](conf); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_HostRequiredForMode(t *testing.T) {
	tests := []struct {
		mode    string
		host    string
		wantErr bool
	}{
		{mode: "server", host: "", wantErr: false},
		{mode: "client", host: "", wantErr: true},
		{mode: "edge", host: "", wantErr: true},
		{mode: "edge", host: "my.host.tld", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.mode+"/"+tt.host, func(t *testing.T) {
			config := GetDefaultConfig()
			config.Mode = tt.mode
			config.Client.Host = tt.host

			err := ValidateConfig(config)
			if gotErr := err != nil && strings.Contains(err.Error(), "'Host'"); gotErr != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, want host error %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_ClientKeypair(t *testing.T) {
	tests := []struct {
		name        string
		mode        string
		keypair     string
		keypairPath string
		wantErr     bool
	}{
		{name: "client without keypair", mode: "client", wantErr: true},
		{name: "client with keypair", mode: "client", keypair: "keypair"},
		{name: "client with keypair path", mode: "client", keypairPath: "/etc/dyndns/keypair.json"},
		{name: "client with both", mode: "client", keypair: "keypair", keypairPath: "/etc/dyndns/keypair.json"},
		{name: "server without keypair", mode: "server"},
		{name: "edge without keypair", mode: "edge"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.Mode = tt.mode
			config.Client.KeyPair = tt.keypair
			config.Client.KeyPairPath = tt.keypairPath

			err := ValidateConfig(config)
			if gotErr := err != nil && strings.Contains(err.Error(), "KeyPair"); gotErr != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, want keypair error %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_AwsCredentials(t *testing.T) {
	keyPath := "/run/secrets/aws"
	tests := []struct {
		name    string
		aws     hybrid.AwsCredentialsConfig
		wantErr bool
	}{
		{name: "none", aws: hybrid.AwsCredentialsConfig{}},
		{name: "key and secret", aws: hybrid.AwsCredentialsConfig{AccessKeyId: "id", SecretAccessKey: "secret"}},
		{name: "key and secret file", aws: hybrid.AwsCredentialsConfig{AccessKeyId: "id", SecretAccessKeyFile: keyPath}},
		{name: "key without secret", aws: hybrid.AwsCredentialsConfig{AccessKeyId: "id"}, wantErr: true},
		{name: "secret without key", aws: hybrid.AwsCredentialsConfig{SecretAccessKey: "secret"}, wantErr: true},
		{name: "secret and secret file", aws: hybrid.AwsCredentialsConfig{AccessKeyId: "id", SecretAccessKey: "secret", SecretAccessKeyFile: keyPath}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetDefaultConfig()
			config.Mode = "server"
			config.Server.Aws = tt.aws

			err := ValidateConfig(config)
			if gotErr := err != nil && strings.Contains(err.Error(), "Aws."); gotErr != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, want aws error %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_SectionsRequired(t *testing.T) {
	for _, mode := range []string{"client", "server", "edge"} {
		t.Run(mode, func(t *testing.T) {
			withoutClient := GetDefaultConfig()
			withoutClient.Mode = mode
			withoutClient.Client = nil
			if err := ValidateConfig(withoutClient); err == nil || !strings.Contains(err.Error(), "'Client'") {
				t.Errorf("ValidateConfig() without client section error = %v, want a Client error", err)
			}

			withoutServer := GetDefaultConfig()
			withoutServer.Mode = mode
			withoutServer.Server = nil
			if err := ValidateConfig(withoutServer); err == nil || !strings.Contains(err.Error(), "'Server'") {
				t.Errorf("ValidateConfig() without server section error = %v, want a Server error", err)
			}
		})
	}
}
