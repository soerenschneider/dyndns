package conf

import (
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
