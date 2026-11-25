package dns

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

const defaultRecordTtl = 60

type route53Api interface {
	ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
	ListResourceRecordSets(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
}

type Route53Propagator struct {
	client       route53Api
	hostedZoneId string
	ttl          int64
}

func NewRoute53Propagator(ctx context.Context, hostedZoneId string, provider aws.CredentialsProvider) (*Route53Propagator, error) {
	if provider != nil {
		log.Info().Str("component", "route53").Msg("Building AWS client using given credentials provider")
	} else {
		log.Info().Str("component", "route53").Msg("Building AWS client using default credentials provider")
	}

	// Route 53 is global; its endpoint is in us-east-1
	awsCfg, err := hybrid.LoadAwsConfig(ctx, "us-east-1", provider)
	if err != nil {
		return nil, err
	}

	return &Route53Propagator{
		client:       route53.NewFromConfig(awsCfg),
		hostedZoneId: hostedZoneId,
		ttl:          defaultRecordTtl,
	}, nil
}

func (dns *Route53Propagator) PropagateChange(ctx context.Context, resolvedIp update.DnsRecord) error {
	changes := getChanges(resolvedIp, dns.ttl)

	deletions, err := dns.getDeletions(ctx, resolvedIp)
	if err != nil {
		return err
	}
	changes = append(changes, deletions...)

	// Withdraw-only record whose records are already gone.
	if len(changes) == 0 && len(resolvedIp.WithdrawnFamilies) > 0 {
		return nil
	}

	if len(changes) == 0 {
		return errors.New("empty list of changes")
	}

	batch := &types.ChangeBatch{
		Changes: changes,
		Comment: aws.String(fmt.Sprintf(
			"Dyndns Change from %s",
			time.Now().Format("2006-01-02T15:04:05Z07:00"),
		)),
	}

	in := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  batch,
		HostedZoneId: &dns.hostedZoneId,
	}

	_, err = dns.client.ChangeResourceRecordSets(ctx, in)
	if err != nil {
		return fmt.Errorf("updating resource failed '%s': %v", resolvedIp.Host, err)
	}

	return nil
}

// getDeletions builds DELETE changes for the records of withdrawn address
// families. Route 53 only deletes a record set given its exact current state,
// so it is looked up first; a record that does not exist needs no change.
func (dns *Route53Propagator) getDeletions(ctx context.Context, resolved update.DnsRecord) ([]types.Change, error) {
	var deletions []types.Change
	for _, family := range resolved.WithdrawnFamilies {
		recordType := types.RRTypeA
		if family == update.AddrFamilyIpv6 {
			recordType = types.RRTypeAaaa
		}

		existing, err := dns.getRecordSet(ctx, resolved.Host, recordType)
		if err != nil {
			return nil, fmt.Errorf("looking up %s record of '%s' to withdraw it: %w", recordType, resolved.Host, err)
		}

		if existing == nil {
			continue
		}

		log.Info().
			Str("component", "route53").
			Str("host", resolved.Host).
			Str("type", string(recordType)).
			Msg("deleting record of withdrawn address family")

		deletions = append(deletions, types.Change{
			Action:            types.ChangeActionDelete,
			ResourceRecordSet: existing,
		})
	}

	return deletions, nil
}

// getRecordSet returns the record set of host and recordType, or nil if there is none.
func (dns *Route53Propagator) getRecordSet(ctx context.Context, host string, recordType types.RRType) (*types.ResourceRecordSet, error) {
	out, err := dns.client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId:    &dns.hostedZoneId,
		StartRecordName: aws.String(host),
		StartRecordType: recordType,
		MaxItems:        aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}

	// The listing starts at host but returns the next record set in the zone
	// if host has none of that type.
	for _, set := range out.ResourceRecordSets {
		if set.Type == recordType && sameDnsName(aws.ToString(set.Name), host) {
			return &set, nil
		}
	}

	return nil, nil
}

func sameDnsName(a, b string) bool {
	return strings.EqualFold(strings.TrimSuffix(a, "."), strings.TrimSuffix(b, "."))
}

func buildChange(host, value string, recordType types.RRType, ttl int64) (types.Change, error) {
	return types.Change{
		Action: types.ChangeActionUpsert,
		ResourceRecordSet: &types.ResourceRecordSet{
			Name: aws.String(host),
			ResourceRecords: []types.ResourceRecord{
				{
					Value: aws.String(value),
				},
			},
			TTL:  aws.Int64(ttl),
			Type: types.RRType(recordType),
		},
	}, nil
}

func getChanges(resolved update.DnsRecord, ttl int64) []types.Change {
	var records []types.Change

	if resolved.HasIpV4() {
		change, err := buildChange(resolved.Host, resolved.IpV4, types.RRTypeA, ttl)
		if err != nil {
			log.Warn().
				Str("component", "route53").
				Err(err).
				Msg("couldn't build change for ipv4")
		} else {
			records = append(records, change)
		}
	}

	if resolved.HasIpV6() {
		change, err := buildChange(resolved.Host, resolved.IpV6, types.RRTypeAaaa, ttl)
		if err != nil {
			log.Warn().
				Str("component", "route53").
				Err(err).
				Msg("couldn't build change for ipv6")
		} else {
			records = append(records, change)
		}
	}

	return records
}
