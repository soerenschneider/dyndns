// //go:build aws
package server

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//
//	"github.com/aws/aws-lambda-go/events"
//	"github.com/aws/aws-lambda-go/lambda"
//	"github.com/rs/zerolog/log"
//	server2 "github.com/soerenschneider/dyndns/v2/internal/conf"
//	"github.com/soerenschneider/dyndns/v2/internal/server"
//	"github.com/soerenschneider/dyndns/v2/internal/server/dns"
//	"github.com/soerenschneider/dyndns/v2/pkg/update"
//)
//
//var propagator dns.Propagator
//var dyndnsServer *server.DyndnsServer
//
//func init() {
//	config := server2.GetDefaultServerConfig()
//	if err := server2.ParseEnvVariables(config); err != nil {
//		log.Fatal().Err(err).Msg("could not parse config")
//	}
//
//	var err error
//	propagator, err = dns.NewRoute53Propagator(config.HostedZoneId, nil)
//	if err != nil {
//		log.Fatal().Err(err).Msg("could not build propagator")
//	}
//
//	c := make(chan update.UpdateRecordRequest)
//	dyndnsServer, err = server.NewServer(*config, propagator, c, nil)
//	if err != nil {
//		log.Fatal().Err(err).Msg("could not build server")
//	}
//}
//
//func handleSQSEvent(_ context.Context, event events.SQSEvent) error {
//	for _, message := range event.Records {
//		payload := update.UpdateRecordRequest{}
//		if err := json.Unmarshal([]byte(message.Body), &payload); err != nil {
//			return err
//		}
//
//		if err := dyndnsServer.HandlePropagateRequest(payload); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func handleAPIGatewayRequest(_ context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
//	payload := update.UpdateRecordRequest{}
//	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
//		return events.APIGatewayProxyResponse{
//			Body:       "could not parse json",
//			StatusCode: 400,
//		}, err
//	}
//
//	if err := dyndnsServer.HandlePropagateRequest(payload); err != nil {
//		return events.APIGatewayProxyResponse{
//			StatusCode: 500,
//		}, err
//	}
//
//	return events.APIGatewayProxyResponse{
//		StatusCode: 200,
//	}, nil
//}
//
//func handler(ctx context.Context, rawEvent json.RawMessage) (interface{}, error) {
//	// Try to parse as APIGatewayProxyRequest
//	var apiGatewayRequest events.APIGatewayProxyRequest
//	if err := json.Unmarshal(rawEvent, &apiGatewayRequest); err == nil && apiGatewayRequest.HTTPMethod != "" {
//		return handleAPIGatewayRequest(ctx, apiGatewayRequest)
//	}
//
//	// Try to parse as SQSEvent
//	var sqsEvent events.SQSEvent
//	if err := json.Unmarshal(rawEvent, &sqsEvent); err == nil && len(sqsEvent.Records) > 0 {
//		return nil, handleSQSEvent(ctx, sqsEvent)
//	}
//
//	return nil, fmt.Errorf("unsupported event type")
//}
//
//func main() {
//	lambda.Start(handler)
//}
