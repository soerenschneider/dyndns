package dispatchers

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
)

type SqsDispatch struct {
	client   *sqs.Client
	queueUrl string
}

func NewSqsDispatcher(sqsConf hybrid.SqsConfig, provider aws.CredentialsProvider) (*SqsDispatch, error) {
	awsConf := aws.Config{
		Region: sqsConf.Region,
	}
	if provider != nil {
		log.Info().Str("component", "sqs").Msg("Building AWS client using given credentials provider")
		awsConf.Credentials = aws.NewCredentialsCache(provider)
	}

	ret := &SqsDispatch{
		queueUrl: sqsConf.SqsQueue,
		client:   sqs.NewFromConfig(awsConf),
	}
	return ret, nil
}

func (h *SqsDispatch) UpdateRecord(ctx context.Context, msg update.UpdateRecordRequest) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	metrics.SqsApiCalls.WithLabelValues("send_message").Inc()
	result, err := h.client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody:  aws.String(string(data)),
		QueueUrl:     aws.String(h.queueUrl),
		DelaySeconds: 0,
	})

	if err == nil {
		log.Info().Str("component", "sqs").Str("message_id", *result.MessageId).Msg("Successfully dispatched message")
	}

	return err
}
