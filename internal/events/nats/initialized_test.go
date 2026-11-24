package nats

import (
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

// connectedJetStream stands in for an established connection; the
// constructors only check it for nil.
type connectedJetStream struct {
	jetstream.JetStream
}

func TestConstructors_ConnectedJetStreamIsInitialized(t *testing.T) {
	config := &hybrid.NatsConfig{}
	js := connectedJetStream{}

	client, err := NewNatsDyndnsClient(config, js)
	if err != nil {
		t.Fatal(err)
	}
	if !client.isInitialized.Load() {
		t.Error("NatsDyndnsClient: not initialized despite connected jetstream")
	}

	cloudevents, err := NewNatsCloudevents(config, js)
	if err != nil {
		t.Fatal(err)
	}
	if !cloudevents.isInitialized.Load() {
		t.Error("CloudeventsClient: not initialized despite connected jetstream")
	}

	server, err := NewNatsDyndnsServer(config, js, make(chan update.UpdateRecordRequest))
	if err != nil {
		t.Fatal(err)
	}
	if !server.isInitialized.Load() {
		t.Error("NatsDyndnsServer: not initialized despite connected jetstream")
	}
	if server.reqChan == nil {
		t.Error("NatsDyndnsServer: requests channel not stored")
	}
}
