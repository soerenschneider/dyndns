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
				MqttConfig: MqttConfig{
					Brokers:  []string{"broker-1", "broker-2"},
					ClientId: "my-client-id",
				},
				VaultConfig: VaultConfig{
					RoleName:      "my-role-name",
					VaultAddr:     "https://vault:8200",
					AppRoleId:     "my-approle-id",
					AppRoleSecret: "my-approle-secret",
					VaultToken:    "the-holy-token",
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
