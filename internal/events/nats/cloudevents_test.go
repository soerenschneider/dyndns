package nats

import (
	"context"
	"encoding/json"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"github.com/soerenschneider/soeren.cloud-events/pkg"
)

// publishingJetStream records the messages published to it.
type publishingJetStream struct {
	jetstream.JetStream
	published []*nats.Msg
}

func (p *publishingJetStream) PublishMsg(_ context.Context, msg *nats.Msg, _ ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	p.published = append(p.published, msg)
	return &jetstream.PubAck{}, nil
}

func TestNewNatsCloudevents_EmptySource(t *testing.T) {
	if _, err := NewNatsCloudevents(context.Background(), &hybrid.NatsConfig{}, &publishingJetStream{}, ""); err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestCloudeventsClient_PublishesValidEvents(t *testing.T) {
	js := &publishingJetStream{}
	client, err := NewNatsCloudevents(context.Background(), &hybrid.NatsConfig{EventsSubject: "dyndns.events"}, js, "dyndns/test")
	if err != nil {
		t.Fatal(err)
	}

	record := &update.DnsRecord{Host: "home.example.com", IpV4: "8.8.8.8"}
	if err := client.NotifyUpdatedIpDetected(record); err != nil {
		t.Fatal(err)
	}
	if err := client.NotifyUpdatedIpApplied(record); err != nil {
		t.Fatal(err)
	}

	wantTypes := []string{string(pkg.TypeDyndnsNewIpDetected), string(pkg.TypeDyndnsNewIpApplied)}
	if len(js.published) != len(wantTypes) {
		t.Fatalf("published %d messages, want %d", len(js.published), len(wantTypes))
	}

	for i, msg := range js.published {
		assertCloudevent(t, msg, "dyndns.events", "dyndns/test", wantTypes[i])
	}
}

func assertCloudevent(t *testing.T, msg *nats.Msg, wantSubject, wantSource, wantType string) {
	t.Helper()

	if msg.Subject != wantSubject {
		t.Errorf("subject = %q, want %q", msg.Subject, wantSubject)
	}

	event := cloudevents.NewEvent()
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		t.Fatalf("message is not a cloudevent: %v", err)
	}
	if err := event.Validate(); err != nil {
		t.Errorf("invalid cloudevent: %v", err)
	}
	if event.Source() != wantSource || event.Type() != wantType {
		t.Errorf("event source/type = %q/%q, want %q/%q", event.Source(), event.Type(), wantSource, wantType)
	}
}
