package verification

import (
	"testing"
	"time"

	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

func TestEdge_Sign(t *testing.T) {
	host := "a.my.tld"
	e := edge{
		Host: host,
	}

	record := update.DnsRecord{
		IpV4:      "1.2.3.4",
		IpV6:      "::1",
		Host:      host,
		Timestamp: time.Now(),
	}

	signature := e.Sign(record)

	if verificationSuccessful := e.Verify(signature, record); !verificationSuccessful {
		t.Fail()
	}
}
