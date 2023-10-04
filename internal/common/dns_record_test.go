package common

import (
	"testing"
)

func TestResolvedIp_IsValid(t *testing.T) {
	type fields struct {
		IpV4 string
		IpV6 string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "no valid ips",
			fields: fields{},
			want:   false,
		},
		{
			name: "only one ipv4 which is garbage",
			fields: fields{
				IpV4: "this is garbage",
			},
			want: false,
		},
		{
			name: "only one ipv6 which is garbage",
			fields: fields{
				IpV6: "this is garbage",
			},
			want: false,
		},
		{
			name: "both ips are garbage",
			fields: fields{
				IpV4: "this is garbage",
				IpV6: "this as well",
			},
			want: false,
		},
		{
			name: "one valid ipv4, ipv6 empty",
			fields: fields{
				IpV4: "1.1.1.1",
			},
			want: true,
		},
		{
			name: "invalid because private",
			fields: fields{
				IpV4: "192.168.1.1",
			},
			want: false,
		},
		{
			name: "invalid because loopback",
			fields: fields{
				IpV4: "127.0.0.1",
			},
			want: false,
		},
		{
			name: "valid",
			fields: fields{
				IpV4: "8.8.8.8",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := &DnsRecord{
				IpV4: tt.fields.IpV4,
				IpV6: tt.fields.IpV6,
			}
			if got := resolved.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
