package conf

import (
	"os"
	"reflect"
	"testing"
)

func TestReadServerConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *ServerConf
		wantErr bool
	}{
		{
			name: "happy path - yaml",
			args: args{"../contrib/server.yaml"},
			want: &ServerConf{
				KnownHosts: map[string][]string{
					"host": []string{"key1", "key2"},
				},
				HostedZoneId:    "hosted-zone-id-x",
				MetricsListener: ":6666",
				MqttConfig: &MqttConfig{
					Brokers:  []string{"tcp://mqtt.eclipseprojects.io:1883"},
					ClientId: "my-client-id",
				},
				EmailConfig: &EmailConfig{
					From:         "from",
					To:           []string{"to-1"},
					SmtpHost:     "smtp-host",
					SmtpPort:     465,
					SmtpUsername: "username",
					SmtpPassword: "password",
				},
				VaultConfig: GetDefaultVaultConfig(),
			},
		},
		{
			name: "happy path - json",
			args: args{"../contrib/server.json"},
			want: &ServerConf{
				KnownHosts: map[string][]string{
					"host": []string{"key1", "key2"},
				},
				HostedZoneId:    "hosted-zone-id-x",
				MetricsListener: ":6666",
				MqttConfig: &MqttConfig{
					Brokers:  []string{"tcp://mqtt.eclipseprojects.io:1883"},
					ClientId: "my-client-id",
				},
				EmailConfig: &EmailConfig{
					From:         "from",
					To:           []string{"to-1"},
					SmtpHost:     "smtp-host",
					SmtpPort:     465,
					SmtpUsername: "username",
					SmtpPassword: "password",
				},
				VaultConfig: GetDefaultVaultConfig(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadServerConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadServerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadServerConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerConf_ParseEnvVariables_KnownHosts(t *testing.T) {
	envKey := "DYNDNS_KNOWN_HOSTS"
	os.Setenv(envKey, "{\"key1\": [\"value1\", \"value2\"], \"key2\": [\"value3\", \"value4\"]}")
	// unset after running test
	defer os.Setenv(envKey, "")

	empty := &ServerConf{}
	err := ParseEnvVariables(empty)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string][]string{
		"key1": []string{"value1", "value2"},
		"key2": []string{"value3", "value4"},
	}

	if !reflect.DeepEqual(empty.KnownHosts, expected) {
		t.Fatalf("expected %v, got %v", expected, empty.KnownHosts)
	}
}

func TestServerConf_ParseEnvVariables_AuthStrategy(t *testing.T) {
	envKey := "DYNDNS_VAULT_AUTH_STRATEGY"
	os.Setenv(envKey, "approle")
	// unset after running test
	defer os.Setenv(envKey, "")

	empty := &ServerConf{
		VaultConfig: &VaultConfig{},
	}
	err := ParseEnvVariables(empty)
	if err != nil {
		t.Fatal(err)
	}

	expected := VaultAuthStrategyApprole
	if !reflect.DeepEqual(empty.AuthStrategy, expected) {
		t.Fatalf("expected %v, got %v", expected, empty.AuthStrategy)
	}
}

func TestServerConf_ParseEnvVariables_AuthStrategy_Invalid(t *testing.T) {
	envKey := "DYNDNS_VAULT_AUTH_STRATEGY"
	os.Setenv(envKey, "unknown")
	// unset after running test
	defer os.Setenv(envKey, "")

	empty := &ServerConf{
		VaultConfig: &VaultConfig{},
	}
	err := ParseEnvVariables(empty)
	if err != nil {
		t.Fatal(err)
	}

	for _, not := range []VaultAuthStrategy{VaultAuthStrategyToken, VaultAuthStrategyKubernetes, VaultAuthStrategyApprole} {
		if reflect.DeepEqual(empty.AuthStrategy, not) {
			t.Fatalf("must not be %v, got %v", not, empty.AuthStrategy)
		}
	}
}

func TestServerConf_GetKnownHostsHash_HappyPath(t *testing.T) {
	var h1, h2 map[string][]string
	h1 = map[string][]string{
		"host1": []string{"abc"},
	}
	hash1, err := GetKnownHostsHash(h1)
	if err != nil {
		t.Fatal()
	}

	h2 = map[string][]string{
		"host1": []string{"abc"},
	}
	hash2, err := GetKnownHostsHash(h2)
	if err != nil {
		t.Fatal()
	}

	if hash1 != hash2 {
		t.Fatal()
	}
}

func TestServerConf_GetKnownHostsHash_MultipleHosts(t *testing.T) {
	host1 := "host1"
	host2 := "host2"

	var h1, h2 map[string][]string
	h1 = map[string][]string{
		host1: []string{"abc", "1234"},
		host2: []string{"zzz"},
	}
	hash1, err := GetKnownHostsHash(h1)
	if err != nil {
		t.Fatal()
	}

	h2 = map[string][]string{
		host2: []string{"zzz"},
		host1: []string{"abc", "1234"},
	}
	hash2, err := GetKnownHostsHash(h2)
	if err != nil {
		t.Fatal()
	}

	if hash1 != hash2 {
		t.Fatal()
	}
}

func TestServerConf_GetKnownHostsHash_ListWrongOrder(t *testing.T) {
	host1 := "host1"
	host2 := "host2"

	var h1, h2 map[string][]string
	h1 = map[string][]string{
		host1: []string{"abc", "1234"},
		host2: []string{"zzz"},
	}
	hash1, err := GetKnownHostsHash(h1)
	if err != nil {
		t.Fatal()
	}

	h2 = map[string][]string{
		host2: []string{"zzz"},
		host1: []string{"1234", "abc"},
	}
	hash2, err := GetKnownHostsHash(h2)
	if err != nil {
		t.Fatal()
	}

	if hash1 == hash2 {
		t.Fatal()
	}
}
