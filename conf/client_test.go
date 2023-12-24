package conf

import (
	"os"
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
				AddrFamilies:    []string{AddrFamilyIpv4},
				KeyPairPath:     "/tmp/keypair.json",
				PreferredUrls:   defaultHttpResolverUrls,
				MetricsListener: "0.0.0.0:9191",
				MqttConfig: MqttConfig{
					Brokers:  []string{"ssl://mqtt.eclipseprojects.io:8883"},
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
			want:    getDefaultClientConfig(),
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

func TestParseClientConfEnv(t *testing.T) {
	envKey := "DYNDNS_HTTP_DISPATCHER_CONF"
	os.Setenv(envKey, "[{\"url\":\"https://one\"}, {\"url\":\"https://two\"}]")
	// unset after running test
	defer os.Setenv(envKey, "")

	empty := &ClientConf{}
	err := ParseClientConfEnv(empty)
	if err != nil {
		t.Fatal(err)
	}

	expected := []HttpDispatcherConfig{
		{
			Url: "https://one",
		},
		{
			Url: "https://two",
		},
	}

	if !reflect.DeepEqual(empty.HttpDispatcherConf, expected) {
		t.Fatalf("expected %v, got %v", expected, empty.HttpDispatcherConf)
	}
}
