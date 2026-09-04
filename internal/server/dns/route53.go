package dns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

const defaultRecordTtl = 60

type Route53Propagator struct {
	client       *route53.Client
	hostedZoneId string
	ttl          int64
}

func NewRoute53Propagator(hostedZoneId string, provider aws.CredentialsProvider) (*Route53Propagator, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	if provider != nil {
		log.Info().
			Str("component", "route53").
			Msg("Building AWS client using given credentials provider")

		awsCfg.Credentials = aws.NewCredentialsCache(provider)
	} else {
		log.Info().
			Str("component", "route53").
			Msg("Building AWS client using default credentials provider")
	}

	svc := route53.NewFromConfig(awsCfg)

	return &Route53Propagator{
		client:       svc,
		hostedZoneId: hostedZoneId,
		ttl:          defaultRecordTtl,
	}, nil
}

func (dns *Route53Propagator) PropagateChange(ctx context.Context, resolvedIp update.DnsRecord) error {
	changes := getChanges(resolvedIp, dns.ttl)
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

	_, err := dns.client.ChangeResourceRecordSets(ctx, in)
	if err != nil {
		return fmt.Errorf("updating resource failed '%s': %v", resolvedIp.Host, err)
	}

	return nil
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
