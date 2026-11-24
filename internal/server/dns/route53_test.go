package dns

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type fakeRoute53 struct {
	// zone holds the record sets in Route 53's listing order.
	zone    []types.ResourceRecordSet
	listErr error
	changes []types.Change
}

func (f *fakeRoute53) ChangeResourceRecordSets(_ context.Context, in *route53.ChangeResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
	f.changes = append(f.changes, in.ChangeBatch.Changes...)
	return &route53.ChangeResourceRecordSetsOutput{}, nil
}

// ListResourceRecordSets mimics Route 53: it returns record sets starting at
// the given name and type, which is the next one in the zone if there is no
// exact match.
func (f *fakeRoute53) ListResourceRecordSets(_ context.Context, in *route53.ListResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}

	for i, set := range f.zone {
		if sameDnsName(aws.ToString(set.Name), aws.ToString(in.StartRecordName)) && set.Type >= in.StartRecordType ||
			aws.ToString(set.Name) > aws.ToString(in.StartRecordName)+"." {
			return &route53.ListResourceRecordSetsOutput{ResourceRecordSets: f.zone[i : i+1]}, nil
		}
	}

	return &route53.ListResourceRecordSetsOutput{}, nil
}

func recordSet(name string, recordType types.RRType, value string) types.ResourceRecordSet {
	return types.ResourceRecordSet{
		Name:            aws.String(name),
		Type:            recordType,
		TTL:             aws.Int64(300),
		ResourceRecords: []types.ResourceRecord{{Value: aws.String(value)}},
	}
}

func TestRoute53Propagator_WithdrawnFamilies(t *testing.T) {
	const host = "home.example.com"
	const v6 = "2a00:1450:4001:830::200e"
	staleA := recordSet(host+".", types.RRTypeA, "203.0.113.7")

	tests := []struct {
		name        string
		zone        []types.ResourceRecordSet
		record      update.DnsRecord
		wantDeleted []types.ResourceRecordSet
		wantUpserts int
	}{
		{
			name:        "stale A record is deleted with its exact current state",
			zone:        []types.ResourceRecordSet{staleA, recordSet(host+".", types.RRTypeAaaa, v6)},
			record:      update.DnsRecord{Host: host, IpV6: v6, WithdrawnFamilies: []string{update.AddrFamilyIpv4}},
			wantDeleted: []types.ResourceRecordSet{staleA},
			wantUpserts: 1,
		},
		{
			name:        "no A record, nothing to delete",
			zone:        []types.ResourceRecordSet{recordSet(host+".", types.RRTypeAaaa, v6)},
			record:      update.DnsRecord{Host: host, IpV6: v6, WithdrawnFamilies: []string{update.AddrFamilyIpv4}},
			wantUpserts: 1,
		},
		{
			name:        "next record in the zone belongs to another host",
			zone:        []types.ResourceRecordSet{recordSet("other.example.com.", types.RRTypeA, "203.0.113.8")},
			record:      update.DnsRecord{Host: host, IpV6: v6, WithdrawnFamilies: []string{update.AddrFamilyIpv4}},
			wantUpserts: 1,
		},
		{
			name:        "withdraw only deletes the A record without upserting",
			zone:        []types.ResourceRecordSet{staleA},
			record:      update.DnsRecord{Host: host, WithdrawnFamilies: []string{update.AddrFamilyIpv4}},
			wantDeleted: []types.ResourceRecordSet{staleA},
			wantUpserts: 0,
		},
		{
			name:        "withdraw only, record already gone, no change sent",
			zone:        nil,
			record:      update.DnsRecord{Host: host, WithdrawnFamilies: []string{update.AddrFamilyIpv4}},
			wantUpserts: 0,
		},
		{
			name:        "nothing withdrawn, nothing looked up or deleted",
			zone:        []types.ResourceRecordSet{staleA},
			record:      update.DnsRecord{Host: host, IpV6: v6},
			wantUpserts: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &fakeRoute53{zone: tt.zone}
			p := &Route53Propagator{client: api, hostedZoneId: "zone", ttl: defaultRecordTtl}

			if err := p.PropagateChange(context.Background(), tt.record); err != nil {
				t.Fatalf("PropagateChange() error = %v", err)
			}

			assertChanges(t, api.changes, tt.wantDeleted, tt.wantUpserts)
		})
	}
}

func assertChanges(t *testing.T, changes []types.Change, wantDeleted []types.ResourceRecordSet, wantUpserts int) {
	t.Helper()

	var deleted []types.ResourceRecordSet
	var upserts int
	for _, c := range changes {
		switch c.Action {
		case types.ChangeActionDelete:
			deleted = append(deleted, *c.ResourceRecordSet)
		case types.ChangeActionUpsert:
			upserts++
		}
	}

	if upserts != wantUpserts {
		t.Errorf("upserts = %d, want %d", upserts, wantUpserts)
	}
	if len(deleted) != len(wantDeleted) {
		t.Fatalf("deleted %d record sets, want %d", len(deleted), len(wantDeleted))
	}
	for i := range deleted {
		if !sameRecordSet(deleted[i], wantDeleted[i]) {
			t.Errorf("deleted %+v, want %+v", deleted[i], wantDeleted[i])
		}
	}
}

func sameRecordSet(got, want types.ResourceRecordSet) bool {
	return aws.ToString(got.Name) == aws.ToString(want.Name) &&
		got.Type == want.Type &&
		aws.ToInt64(got.TTL) == aws.ToInt64(want.TTL) &&
		aws.ToString(got.ResourceRecords[0].Value) == aws.ToString(want.ResourceRecords[0].Value)
}

func TestRoute53Propagator_LookupFailureAbortsChange(t *testing.T) {
	api := &fakeRoute53{listErr: errors.New("access denied")}
	p := &Route53Propagator{client: api, hostedZoneId: "zone", ttl: defaultRecordTtl}

	record := update.DnsRecord{Host: "home.example.com", IpV6: "2a00:1450:4001:830::200e", WithdrawnFamilies: []string{update.AddrFamilyIpv4}}
	if err := p.PropagateChange(context.Background(), record); err == nil {
		t.Fatal("expected error when the record to withdraw cannot be looked up")
	}
	if len(api.changes) != 0 {
		t.Errorf("changes = %d, want none: a partial change would hide the failed withdrawal", len(api.changes))
	}
}
