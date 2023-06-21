package conf

import (
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
			name: "happy path",
			args: args{"../contrib/server.json"},
			want: &ServerConf{
				KnownHosts: map[string][]string{
					"host": []string{"key1", "key2"},
				},
				HostedZoneId:    "hosted-zone-id-x",
				MetricsListener: ":666",
				MqttConfig: &MqttConfig{
					Brokers:  []string{"broker-1", "broker-2"},
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
				VaultConfig: &VaultConfig{
					RoleName:      "my-role-name",
					VaultAddr:     "https://vault:8200",
					AppRoleId:     "my-approle-id",
					AppRoleSecret: "my-approle-secret",
				},
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
