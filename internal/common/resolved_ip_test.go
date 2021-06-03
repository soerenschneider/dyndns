package common

import (
	"testing"
)

func TestResolvedIp_IsValidValidV4(t *testing.T) {
	ipv4 := "8.8.8.8"
	ipv6 := "invalid"
	resolved := &ResolvedIp{IpV4: ipv4, IpV6: ipv6}
	resolved.IsValid()
	if resolved.IpV4 != ipv4 {
		t.Fatalf("Expected %s", ipv4)
	}
	if resolved.IpV6 != "" {
		t.Fatal("Expected empty Ipv6")
	}
}

func TestResolvedIp_IsValidValidV6(t *testing.T) {
	ipv4 := "invalid"
	ipv6 := "::1"
	resolved := &ResolvedIp{IpV4: ipv4, IpV6: ipv6}
	resolved.IsValid()
	if resolved.IpV6 != ipv6 {
		t.Fatalf("Expected %s", ipv6)
	}
	if resolved.IpV4 != "" {
		t.Fatal("Expected empty Ipv4")
	}
}

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
			name: "one valid ipv6, ipv4 empty",
			fields: fields{
				IpV6: "::1",
			},
			want: true,
		},
		{
			name: "both valid",
			fields: fields{
				IpV4: "127.0.0.1",
				IpV6: "::1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := &ResolvedIp{
				IpV4: tt.fields.IpV4,
				IpV6: tt.fields.IpV6,
			}
			if got := resolved.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
