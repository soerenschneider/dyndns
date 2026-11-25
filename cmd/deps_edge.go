package main

import (
	"context"

	client2 "github.com/soerenschneider/dyndns/v2/internal/client"
	"github.com/soerenschneider/dyndns/v2/internal/conf"
)

func newEdge(ctx context.Context, config *conf.Conf) (*appEdgeMode, error) {
	serverApp, err := newServer(ctx, config)
	if err != nil {
		return nil, err
	}

	dispatchers := map[string]client2.RecordUpdater{
		"edge": serverApp.server,
	}

	clientApp, err := newClient(ctx, config, dispatchers)
	if err != nil {
		return nil, err
	}

	return &appEdgeMode{
		serverApp: serverApp,
		clientApp: clientApp,
	}, nil
}
