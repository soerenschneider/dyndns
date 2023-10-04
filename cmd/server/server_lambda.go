//go:build aws

package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
	conf2 "github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal/common"
	server2 "github.com/soerenschneider/dyndns/server"
	"github.com/soerenschneider/dyndns/server/dns"
)

const defaultRegion = "us-east-1"

var propagator dns.Propagator
var server *server2.Server

func init() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	conf := conf2.GetDefaultServerConfig()
	if err := conf2.ParseEnvVariables(conf); err != nil {
		log.Fatal().Err(err).Msg("could not parse config")
	}

	var err error
	propagator, err = dns.NewRoute53Propagator(conf.HostedZoneId, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("could not build propagator")
	}

	c := make(chan common.UpdateRecordRequest)
	server, err = server2.NewServer(*conf, propagator, c, nil)
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

	if err := server.HandlePropagateRequest(payload); err != nil {
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
