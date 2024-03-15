package client

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
)

type SqsDispatch struct {
	client   *sqs.SQS
	queueUrl string
}

func NewSqsDispatcher(queueUrl string, provider credentials.Provider) (*SqsDispatch, error) {
	awsConf := &aws.Config{}
	if provider != nil {
		log.Info().Msg("Building AWS client using given credentials provider")
		awsConf.Credentials = credentials.NewCredentials(provider)
	}
	awsSession := session.Must(session.NewSession(awsConf))

	ret := &SqsDispatch{
		queueUrl: queueUrl,
	}
	ret.client = sqs.New(awsSession)
	return ret, nil
}

func (h *SqsDispatch) Notify(msg *common.UpdateRecordRequest) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// TODO: change interface signature
	ctx := context.Background()
	result, err := h.client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		MessageBody:  aws.String(string(data)),
		QueueUrl:     aws.String(h.queueUrl),
		DelaySeconds: aws.Int64(0),
	})

	if err == nil {
		log.Info().Msgf("Successfully dispatched SQS message %s", *result.MessageId)
	}

	return err
}
