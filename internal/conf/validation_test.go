package conf

import "testing"

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
