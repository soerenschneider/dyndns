package conf

import (
	"reflect"
	"testing"
)

func TestReadClientConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *ClientConf
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{"../contrib/client.json"},
			want: &ClientConf{
				Host:            "my.host.tld",
				KeyPairPath:     "/tmp/keypair.json",
				Urls:            defaultHttpResolverUrls,
				MetricsListener: ":9191",
				MqttConfig: MqttConfig{
					Brokers:  []string{"tcp://mqtt.eclipseprojects.io:1883", "ssl://mqtt.eclipseprojects.io:8883"},
					ClientId: "my-client-id",
				},
			},
			wantErr: false,
		},
		{
			name:    "non-existing-file",
			args:    args{"some-file"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "wrong format",
			args:    args{"../contrib/textfile.txt"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty json file content",
			args:    args{"../contrib/empty.json"},
			want:    getDefaultClientConfig(),
			wantErr: false,
		},
		{
			name:    "empty path",
			args:    args{""},
			want:    &ClientConf{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadClientConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadClientConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadClientConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
