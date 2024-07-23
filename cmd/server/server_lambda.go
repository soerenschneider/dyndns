//go:build aws

package main

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/server"
	"github.com/soerenschneider/dyndns/internal/server/dns"
)

var propagator dns.Propagator
var dyndnsServer *server.DyndnsServer

func init() {
	config := conf.GetDefaultServerConfig()
	if err := conf.ParseEnvVariables(config); err != nil {
		log.Fatal().Err(err).Msg("could not parse config")
	}

	var err error
	propagator, err = dns.NewRoute53Propagator(config.HostedZoneId, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build propagator")
	}

	c := make(chan common.UpdateRecordRequest)
	dyndnsServer, err = server.NewServer(*config, propagator, c, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build server")
	}
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	payload := common.UpdateRecordRequest{}
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       "could not parse json",
			StatusCode: 400,
		}, err
	}

	if err := dyndnsServer.HandlePropagateRequest(payload); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
