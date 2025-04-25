package nats

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	rand2 "math/rand"
	"sync/atomic"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/soeren.cloud-events/pkg/dyndns"
)

type CloudeventsClient struct {
	config       *conf.NatsConfig
	instanceName string

	js            jetstream.JetStream
	isInitialized atomic.Bool
}

func NewNatsCloudevents(config *conf.NatsConfig, js jetstream.JetStream) (*CloudeventsClient, error) {
	if config == nil {
		return nil, errors.New("nil config supplied")
	}

	ret := &CloudeventsClient{
		config:        config,
		js:            js,
		isInitialized: atomic.Bool{},
	}

	ret.isInitialized.Store(js == nil)

	if js == nil {
		go func() {
			for {
				js, err := Connect(*config)
				if err != nil {
					rand := rand2.Intn(15) //nolint G404
					time.Sleep(time.Duration(rand) * time.Second)
				} else {
					ret.js = js
					ret.isInitialized.Store(true)
					return
				}
			}
		}()
	}

	return ret, nil
}

func (n *CloudeventsClient) Close(ctx context.Context) error {
	if !n.isInitialized.Load() {
		return nil
	}

	return Close(ctx, n.js)
}

func (n *CloudeventsClient) NotifyUpdatedIpDetected(ip *common.DnsRecord) error {
	if !n.isInitialized.Load() {
		return ErrNotInitialized
	}

	data := &dyndns.NewIpDetectedEvent{
		IPv4:     ip.IpV4,
		IPv6:     ip.IpV6,
		Hostname: ip.Host,
	}
	event := dyndns.NewNewIpDetectedEvent(n.instanceName, data)
	return n.Accept(context.Background(), event)
}

func (n *CloudeventsClient) NotifyUpdatedIpApplied(ip *common.DnsRecord) error {
	if !n.isInitialized.Load() {
		return ErrNotInitialized
	}

	data := &dyndns.NewIpAppliedEvent{
		IPv4:     ip.IpV4,
		IPv6:     ip.IpV6,
		Hostname: ip.Host,
	}
	event := dyndns.NewNewIpAppliedEvent(n.instanceName, data)
	return n.Accept(context.Background(), event)
}

func (n *CloudeventsClient) Accept(ctx context.Context, event cloudevents.Event) error {
	if !n.isInitialized.Load() {
		return ErrNotInitialized
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ack, err := n.js.PublishMsg(ctx, &nats.Msg{
		Data:    data,
		Subject: n.config.EventsSubject,
	})
	if err != nil {
		return err
	}

	slog.Debug("Published msg", "sequence number", ack.Sequence, "stream", ack.Stream)
	return nil
}
