package vault

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"reflect"
	"testing"
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
