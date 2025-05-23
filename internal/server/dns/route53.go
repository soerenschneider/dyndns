package dns

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
)

type Route53Propagator struct {
	client       *route53.Route53
	hostedZoneId string
	ttl          int64
}

func NewRoute53Propagator(hostedZoneId string, provider credentials.Provider) (*Route53Propagator, error) {
	var awsSession *session.Session
	if provider != nil {
		log.Info().Str("component", "route53").Msg("Building AWS client using given credentials provider")
		awsSession = session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewCredentials(provider),
		}))
	} else {
		log.Info().Str("component", "route53").Msg("Building AWS client session using default provider")
		awsSession = session.Must(session.NewSession())
	}

	svc := route53.New(awsSession)
	return &Route53Propagator{
		client:       svc,
		hostedZoneId: hostedZoneId,
		ttl:          defaultRecordTtl,
	}, nil
}

func (dns *Route53Propagator) PropagateChange(resolvedIp common.DnsRecord) error {
	changes := getChanges(resolvedIp, dns.ttl)
	if len(changes) == 0 {
		return errors.New("empty list of changes")
	}
	batch := &route53.ChangeBatch{
		Changes: changes,
		Comment: aws.String(fmt.Sprintf("Dyndns Change from %s", time.Now().Format("2006-01-02T15:04:05Z07:00"))),
	}

	in := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  batch,
		HostedZoneId: &dns.hostedZoneId,
	}

	_, err := dns.client.ChangeResourceRecordSets(in)
	if err != nil {
		return fmt.Errorf("updating resource failed '%s': %v", resolvedIp.Host, err)
	}

	return nil
}

func buildChange(host, value, recordType string, ttl int64) (*route53.Change, error) {
	validTypeSupplied := false
	for _, t := range route53.RRType_Values() {
		if t == recordType {
			validTypeSupplied = true
			break
		}
	}

	if !validTypeSupplied {
		return nil, fmt.Errorf("invalid record type supplied: %s", recordType)
	}

	return &route53.Change{
		Action: aws.String(route53.ChangeActionUpsert),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name: aws.String(host),
			ResourceRecords: []*route53.ResourceRecord{
				{Value: aws.String(value)},
			},
			TTL:  aws.Int64(ttl),
			Type: aws.String(recordType),
		},
	}, nil
}

func getChanges(resolved common.DnsRecord, ttl int64) []*route53.Change {
	var records []*route53.Change

	if resolved.HasIpV4() {
		change, err := buildChange(resolved.Host, resolved.IpV4, route53.RRTypeA, ttl)
		if err != nil {
			log.Warn().Str("component", "route53").Err(err).Msg("couldn't build change for ipv4")
		} else {
			records = append(records, change)
		}
	}

	if resolved.HasIpV6() {
		change, err := buildChange(resolved.Host, resolved.IpV6, route53.RRTypeAaaa, ttl)
		if err != nil {
			log.Warn().Str("component", "route53").Err(err).Msg("couldn't build change for ipv6")
		} else {
			records = append(records, change)
		}
	}

	return records
}
