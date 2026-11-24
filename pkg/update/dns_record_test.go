package update

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
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
		{
			name: "invalid, cgnat",
			fields: fields{
				IpV4: "100.64.12.34",
			},
			want: false,
		},
		{
			name: "valid ipv6",
			fields: fields{
				IpV6: "2a00:1450:4001:830::200e",
			},
			want: true,
		},
		{
			name: "valid ipv4 and ipv6",
			fields: fields{
				IpV4: "8.8.8.8",
				IpV6: "2a00:1450:4001:830::200e",
			},
			want: true,
		},
		{
			name: "valid ipv4, invalid ipv6 must not be ignored",
			fields: fields{
				IpV4: "8.8.8.8",
				IpV6: "fe80::1",
			},
			want: false,
		},
		{
			name: "valid ipv4, garbage ipv6",
			fields: fields{
				IpV4: "8.8.8.8",
				IpV6: "invalid IP",
			},
			want: false,
		},
		{
			name: "cgnat ipv4, valid ipv6",
			fields: fields{
				IpV4: "100.64.12.34",
				IpV6: "2a00:1450:4001:830::200e",
			},
			want: false,
		},
		{
			name: "garbage ipv4, valid ipv6",
			fields: fields{
				IpV4: "this is garbage",
				IpV6: "2a00:1450:4001:830::200e",
			},
			want: false,
		},
		{
			name: "ipv6 in ipv4 field",
			fields: fields{
				IpV4: "2a00:1450:4001:830::200e",
			},
			want: false,
		},
		{
			name: "ipv4 in ipv6 field",
			fields: fields{
				IpV6: "8.8.8.8",
			},
			want: false,
		},
		{
			name: "v4-mapped ipv6",
			fields: fields{
				IpV6: "::ffff:8.8.8.8",
			},
			want: false,
		},
		{
			name:   "ipv4 link-local",
			fields: fields{IpV4: "169.254.1.1"},
			want:   false,
		},
		{
			name:   "ipv4 this-network",
			fields: fields{IpV4: "0.1.2.3"},
			want:   false,
		},
		{
			name:   "ipv4 unspecified",
			fields: fields{IpV4: "0.0.0.0"},
			want:   false,
		},
		{
			name:   "ipv4 broadcast",
			fields: fields{IpV4: "255.255.255.255"},
			want:   false,
		},
		{
			name:   "ipv4 multicast",
			fields: fields{IpV4: "224.0.0.1"},
			want:   false,
		},
		{
			name:   "ipv4 documentation",
			fields: fields{IpV4: "203.0.113.5"},
			want:   false,
		},
		{
			name:   "ipv4 benchmarking",
			fields: fields{IpV4: "198.18.0.1"},
			want:   false,
		},
		{
			name:   "ipv4 reserved",
			fields: fields{IpV4: "240.0.0.1"},
			want:   false,
		},
		{
			name:   "ipv4 just outside cgnat",
			fields: fields{IpV4: "100.128.0.1"},
			want:   true,
		},
		{
			name:   "ipv6 loopback",
			fields: fields{IpV6: "::1"},
			want:   false,
		},
		{
			name:   "ipv6 unspecified",
			fields: fields{IpV6: "::"},
			want:   false,
		},
		{
			name:   "ipv6 link-local",
			fields: fields{IpV6: "fe80::1"},
			want:   false,
		},
		{
			name:   "ipv6 link-local with zone",
			fields: fields{IpV6: "fe80::1%eth0"},
			want:   false,
		},
		{
			name:   "ipv6 ula",
			fields: fields{IpV6: "fd12:3456:789a::1"},
			want:   false,
		},
		{
			name:   "ipv6 multicast",
			fields: fields{IpV6: "ff02::1"},
			want:   false,
		},
		{
			name:   "ipv6 documentation",
			fields: fields{IpV6: "2001:db8::1"},
			want:   false,
		},
		{
			name:   "ipv6 documentation 3fff::/20",
			fields: fields{IpV6: "3fff::1"},
			want:   false,
		},
		{
			name:   "ipv6 teredo",
			fields: fields{IpV6: "2001:0:4136:e378::1"},
			want:   false,
		},
		{
			name:   "ipv6 6to4",
			fields: fields{IpV6: "2002:c000:0204::1"},
			want:   false,
		},
		{
			name:   "ipv6 nat64",
			fields: fields{IpV6: "64:ff9b::808:808"},
			want:   false,
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

func TestDnsRecord_DropNonPublic(t *testing.T) {
	tests := []struct {
		name          string
		ipV4, ipV6    string
		wantV4        string
		wantV6        string
		wantDroppedV4 string
		wantDroppedV6 string
		wantValid     bool
	}{
		{
			name:      "both public, nothing dropped",
			ipV4:      "8.8.8.8",
			ipV6:      "2a00:1450:4001:830::200e",
			wantV4:    "8.8.8.8",
			wantV6:    "2a00:1450:4001:830::200e",
			wantValid: true,
		},
		{
			name:          "cgnat ipv4 dropped, ipv6 kept",
			ipV4:          "100.64.12.34",
			ipV6:          "2a00:1450:4001:830::200e",
			wantV6:        "2a00:1450:4001:830::200e",
			wantDroppedV4: "100.64.12.34",
			wantValid:     true,
		},
		{
			name:          "link-local ipv6 dropped, ipv4 kept",
			ipV4:          "8.8.8.8",
			ipV6:          "fe80::1",
			wantV4:        "8.8.8.8",
			wantDroppedV6: "fe80::1",
			wantValid:     true,
		},
		{
			name:          "only cgnat ipv4, record left empty",
			ipV4:          "100.64.12.34",
			wantDroppedV4: "100.64.12.34",
			wantValid:     false,
		},
		{
			name:          "both dropped",
			ipV4:          "192.168.1.1",
			ipV6:          "fd12:3456:789a::1",
			wantDroppedV4: "192.168.1.1",
			wantDroppedV6: "fd12:3456:789a::1",
			wantValid:     false,
		},
		{
			name:      "empty record",
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DnsRecord{IpV4: tt.ipV4, IpV6: tt.ipV6}
			droppedV4, droppedV6 := r.DropNonPublic()
			if droppedV4 != tt.wantDroppedV4 || droppedV6 != tt.wantDroppedV6 {
				t.Errorf("DropNonPublic() dropped = (%q, %q), want (%q, %q)", droppedV4, droppedV6, tt.wantDroppedV4, tt.wantDroppedV6)
			}
			if r.IpV4 != tt.wantV4 || r.IpV6 != tt.wantV6 {
				t.Errorf("record after DropNonPublic() = (%q, %q), want (%q, %q)", r.IpV4, r.IpV6, tt.wantV4, tt.wantV6)
			}
			if got := r.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() after DropNonPublic() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestIsCGNAT(t *testing.T) {
	for ip, want := range map[string]bool{
		"100.64.0.0":      true,
		"100.127.255.255": true,
		"100.63.255.255":  false,
		"100.128.0.0":     false,
		"8.8.8.8":         false,
		"garbage":         false,
		"":                false,
	} {
		if got := IsCGNAT(ip); got != want {
			t.Errorf("IsCGNAT(%q) = %v, want %v", ip, got, want)
		}
	}
}

func TestDnsRecord_Validate(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		record  DnsRecord
		wantErr bool
	}{
		{
			name:   "valid ipv4",
			record: DnsRecord{Host: "my.host.tld", IpV4: "8.8.8.8", Timestamp: now},
		},
		{
			name:   "valid ipv6",
			record: DnsRecord{Host: "my.host.tld", IpV6: "2a00:1450:4001:830::200e", Timestamp: now},
		},
		{
			name:   "valid ipv4 and ipv6",
			record: DnsRecord{Host: "my.host.tld", IpV4: "8.8.8.8", IpV6: "2a00:1450:4001:830::200e", Timestamp: now},
		},
		{
			name:    "missing host",
			record:  DnsRecord{IpV4: "8.8.8.8", Timestamp: now},
			wantErr: true,
		},
		{
			name:    "no addresses",
			record:  DnsRecord{Host: "my.host.tld", Timestamp: now},
			wantErr: true,
		},
		{
			name:    "missing timestamp",
			record:  DnsRecord{Host: "my.host.tld", IpV4: "8.8.8.8"},
			wantErr: true,
		},
		{
			name:    "cgnat ipv4",
			record:  DnsRecord{Host: "my.host.tld", IpV4: "100.64.12.34", Timestamp: now},
			wantErr: true,
		},
		{
			name:    "private ipv4",
			record:  DnsRecord{Host: "my.host.tld", IpV4: "10.0.0.1", Timestamp: now},
			wantErr: true,
		},
		{
			name:    "garbage ipv4",
			record:  DnsRecord{Host: "my.host.tld", IpV4: "invalid IP", Timestamp: now},
			wantErr: true,
		},
		{
			name:    "public ipv4, link-local ipv6",
			record:  DnsRecord{Host: "my.host.tld", IpV4: "8.8.8.8", IpV6: "fe80::1", Timestamp: now},
			wantErr: true,
		},
		{
			name:    "ula ipv6",
			record:  DnsRecord{Host: "my.host.tld", IpV6: "fd12:3456:789a::1", Timestamp: now},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.record.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDnsRecord_HashWithdrawnFamilies(t *testing.T) {
	ts := time.Unix(1700000000, 0)
	plain := DnsRecord{Host: "my.host.tld", IpV6: "2a00:1450:4001:830::200e", Timestamp: ts}

	// Must stay byte-identical to the pre-withdrawal format so existing
	// clients and servers keep verifying each other.
	want := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d%s%s%s", ts.Unix(), plain.Host, plain.IpV4, plain.IpV6))))
	if got := plain.Hash(); got != want {
		t.Errorf("Hash() without withdrawn families = %s, want legacy %s", got, want)
	}

	withdrawn := plain
	withdrawn.WithdrawnFamilies = []string{AddrFamilyIpv4}
	if withdrawn.Hash() == plain.Hash() {
		t.Error("withdrawn families must be covered by the hash")
	}

	both := plain
	both.WithdrawnFamilies = []string{AddrFamilyIpv6, AddrFamilyIpv4}
	reordered := plain
	reordered.WithdrawnFamilies = []string{AddrFamilyIpv4, AddrFamilyIpv6}
	if both.Hash() != reordered.Hash() {
		t.Error("hash must not depend on the order of withdrawn families")
	}
}

func TestDnsRecord_Withdraw(t *testing.T) {
	r := &DnsRecord{}
	r.Withdraw(AddrFamilyIpv4)
	r.Withdraw(AddrFamilyIpv4)
	if len(r.WithdrawnFamilies) != 1 || !r.IsWithdrawn(AddrFamilyIpv4) || r.IsWithdrawn(AddrFamilyIpv6) {
		t.Errorf("Withdraw() = %v, want [ip4] once", r.WithdrawnFamilies)
	}
}

func TestDnsRecord_EqualsWithdrawnFamilies(t *testing.T) {
	a := &DnsRecord{Host: "h", IpV6: "2a00:1450:4001:830::200e"}
	b := &DnsRecord{Host: "h", IpV6: "2a00:1450:4001:830::200e", WithdrawnFamilies: []string{AddrFamilyIpv4}}
	if a.Equals(b) {
		t.Error("records differing in withdrawn families must not be equal")
	}

	c := &DnsRecord{Host: "h", WithdrawnFamilies: []string{AddrFamilyIpv6, AddrFamilyIpv4}}
	d := &DnsRecord{Host: "h", WithdrawnFamilies: []string{AddrFamilyIpv4, AddrFamilyIpv6}}
	if !c.Equals(d) {
		t.Error("order of withdrawn families must not matter")
	}
}

func TestDnsRecord_ValidateWithdrawnFamilies(t *testing.T) {
	now := time.Now()
	v6 := "2a00:1450:4001:830::200e"
	tests := []struct {
		name    string
		record  DnsRecord
		wantErr bool
	}{
		{"withdraw ipv4, keep ipv6", DnsRecord{Host: "h", IpV6: v6, WithdrawnFamilies: []string{AddrFamilyIpv4}, Timestamp: now}, false},
		{"unknown family", DnsRecord{Host: "h", IpV6: v6, WithdrawnFamilies: []string{"ip5"}, Timestamp: now}, true},
		{"ipv6 set and withdrawn", DnsRecord{Host: "h", IpV6: v6, WithdrawnFamilies: []string{AddrFamilyIpv6}, Timestamp: now}, true},
		{"ipv4 set and withdrawn", DnsRecord{Host: "h", IpV4: "8.8.8.8", WithdrawnFamilies: []string{AddrFamilyIpv4}, Timestamp: now}, true},
		{"withdraw only", DnsRecord{Host: "h", WithdrawnFamilies: []string{AddrFamilyIpv4}, Timestamp: now}, false},
		{"withdraw both families", DnsRecord{Host: "h", WithdrawnFamilies: []string{AddrFamilyIpv4, AddrFamilyIpv6}, Timestamp: now}, false},
		{"neither addresses nor withdrawn", DnsRecord{Host: "h", Timestamp: now}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.record.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDnsRecord_IsValidWithdrawOnly(t *testing.T) {
	tests := []struct {
		name   string
		record DnsRecord
		want   bool
	}{
		{"withdraw only", DnsRecord{WithdrawnFamilies: []string{AddrFamilyIpv4}}, true},
		{"withdrawn ipv4, public ipv6", DnsRecord{IpV6: "2a00:1450:4001:830::200e", WithdrawnFamilies: []string{AddrFamilyIpv4}}, true},
		{"withdrawn ipv4, non-public ipv6", DnsRecord{IpV6: "fe80::1", WithdrawnFamilies: []string{AddrFamilyIpv4}}, false},
		{"nothing at all", DnsRecord{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDnsRecord_String(t *testing.T) {
	tests := []struct {
		record DnsRecord
		want   string
	}{
		{DnsRecord{Host: "h", IpV4: "8.8.8.8"}, "h: 8.8.8.8 (v4)"},
		{DnsRecord{Host: "h", IpV4: "8.8.8.8", IpV6: "2a00::1"}, "h: 8.8.8.8 (v4), 2a00::1 (v6)"},
		{DnsRecord{Host: "h", IpV6: "2a00::1", WithdrawnFamilies: []string{AddrFamilyIpv4}}, "h: 2a00::1 (v6), ip4 withdrawn"},
		{DnsRecord{Host: "h", WithdrawnFamilies: []string{AddrFamilyIpv6, AddrFamilyIpv4}}, "h: ip4 withdrawn, ip6 withdrawn"},
	}
	for _, tt := range tests {
		if got := tt.record.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}
