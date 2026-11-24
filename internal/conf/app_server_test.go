package conf

import (
	"os"
	"reflect"
	"testing"

	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
)

func TestReadConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Conf
		wantErr bool
	}{
		{
			name: "happy path - yaml",
			args: args{"../../contrib/server.yaml"},
			want: wantServerConf(),
		},
		{
			name: "happy path - json",
			args: args{"../../contrib/server.json"},
			want: wantServerConf(),
		},
		{
			name:    "non-existing-file",
			args:    args{"some-file"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "wrong format",
			args:    args{"../../contrib/textfile.txt"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty yaml file content",
			args: args{"../../contrib/empty.yaml"},
			want: GetDefaultConfig(),
		},
		{
			name: "empty path",
			args: args{""},
			want: GetDefaultConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func wantServerConf() *Conf {
	return &Conf{
		Mode:            "server",
		MetricsListener: ":6666",
		Client:          getDefaultClientConfig(),
		Server: &ServerConf{
			KnownHosts: map[string][]string{
				"host": []string{"key1", "key2"},
			},
			SqsConfig:    hybrid.DefaultSqsConfig(),
			HostedZoneId: "hosted-zone-id-x",
			MqttConfig: hybrid.MqttConfig{
				Brokers:  []string{"tcp://mqtt.eclipseprojects.io:1883"},
				ClientId: "my-client-id",
			},
		},
		EmailConfig: hybrid.EmailConfig{
			From:         "from",
			To:           []string{"to-1"},
			SmtpHost:     "smtp-host",
			SmtpPort:     465,
			SmtpUsername: "username",
			SmtpPassword: "password",
		},
	}
}

func TestServerConf_ParseEnvVariables_KnownHosts(t *testing.T) {
	envKey := "DYNDNS_KNOWN_HOSTS"
	os.Setenv(envKey, "{\"key1\": [\"value1\", \"value2\"], \"key2\": [\"value3\", \"value4\"]}")
	// unset after running test
	defer os.Setenv(envKey, "")

	empty := GetDefaultConfig()
	err := ParseEnvVariables(empty)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string][]string{
		"key1": []string{"value1", "value2"},
		"key2": []string{"value3", "value4"},
	}

	if !reflect.DeepEqual(empty.Server.KnownHosts, expected) {
		t.Fatalf("expected %v, got %v", expected, empty.Server.KnownHosts)
	}
}
