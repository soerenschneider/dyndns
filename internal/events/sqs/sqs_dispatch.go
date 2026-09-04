package client

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"go.uber.org/multierr"
)

const defaultWaitTimeSeconds = 20

type SqsListener struct {
	client   *sqs.Client
	queueUrl string
	requests chan update.UpdateRecordRequest

	waitTimeSeconds int64
}

type SqsOpts func(consumer *SqsListener) error

func NewSqsConsumer(sqsConf hybrid.SqsConfig, provider aws.CredentialsProvider, reqChan chan update.UpdateRecordRequest, opts ...SqsOpts) (*SqsListener, error) {
	if reqChan == nil {
		return nil, errors.New("empty chan provided")
	}

	ret := &SqsListener{
		queueUrl:        sqsConf.SqsQueue,
		requests:        reqChan,
		waitTimeSeconds: defaultWaitTimeSeconds,
	}

	var errs error
	for _, opt := range opts {
		if err := opt(ret); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	if errs != nil {
		return nil, errs
	}

	awsConf := aws.Config{
		Region: sqsConf.Region,
	}

	if provider != nil {
		log.Info().Str("component", "sqs").Msg("Building AWS client using given credentials provider")
		awsConf.Credentials = aws.NewCredentialsCache(provider)
	}

	ret.client = sqs.NewFromConfig(awsConf)
	return ret, nil
}

func (h *SqsListener) Listen(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	ticker := time.NewTicker(1 * time.Minute)
	if err := h.fetchMessages(ctx); err != nil {
		log.Error().Err(err).Str("component", "sqs").Msg("Fetching messages failed")
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Str("component", "sqs").Msg("Received signal, stopping listener")
			wg.Done()
			return nil
		case <-ticker.C:
			if err := h.fetchMessages(ctx); err != nil {
				log.Error().Err(err).Str("component", "sqs").Msg("Fetching messages failed")
			}
		}
	}
}

func (h *SqsListener) fetchMessages(ctx context.Context) error {
	log.Debug().Str("component", "sqs").Msg("Trying to receive messages")
	metrics.SqsApiCalls.WithLabelValues("receive_message").Inc()
	result, err := h.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(h.queueUrl),
		MaxNumberOfMessages: 10,
		VisibilityTimeout:   30,
		WaitTimeSeconds:     int32(h.waitTimeSeconds), //nolint G115
	})
	if err != nil {
		return err
	}

	var errs error
	for _, message := range result.Messages {
		if err := h.handleMessage(ctx, message); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}

func (h *SqsListener) handleMessage(ctx context.Context, message types.Message) error {
	defer func() {
		// the client is not going to stop ip update requests as long the ip has not been updated, so we have the luxury
		// to not care about edge cases too much and delete the message after receiving it.
		log.Debug().Str("component", "sqs").Str("message_id", *message.MessageId).Msg("Deleting message from queue")
		_, err := h.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(h.queueUrl),
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			log.Error().Err(err).Str("component", "sqs").Str("message_id", *message.MessageId).Msg("Could not delete message from queue")
		}
		metrics.SqsApiCalls.WithLabelValues("delete_message").Inc()
	}()

	if message.Body == nil {
		log.Warn().Str("component", "sqs").Msg("Received empty message")
		return nil
	}

	err := h.dispatch([]byte(*message.Body))
	if err != nil {
		return err
	}

	return nil
}

func (h *SqsListener) dispatch(msg []byte) error {
	var env update.UpdateRecordRequest
	err := json.Unmarshal(msg, &env)
	if err != nil {
		metrics.MessageParsingFailed.Inc()
		log.Warn().Str("component", "sqs").Err(err).Msg("Message parsing failed")
		return err
	}

	h.requests <- env
	return nil
}
