package main

import (
	client2 "github.com/soerenschneider/dyndns/v2/internal/client"
	"github.com/soerenschneider/dyndns/v2/internal/conf"
)

func newEdge(config *conf.Conf) (*appEdgeMode, error) {
	serverApp, err := newServer(config)
	if err != nil {
		return nil, err
	}

	dispatchers := map[string]client2.RecordUpdater{
		"edge": serverApp.server,
	}

	clientApp, err := newClient(config, dispatchers)
	if err != nil {
		return nil, err
	}

	return &appEdgeMode{
		serverApp: serverApp,
		clientApp: clientApp,
	}, nil
}
