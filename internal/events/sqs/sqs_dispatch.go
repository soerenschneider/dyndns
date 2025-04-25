package client

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/conf"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"go.uber.org/multierr"
)

const defaultWaitTimeSeconds = 20

type SqsListener struct {
	client   *sqs.SQS
	queueUrl string
	requests chan common.UpdateRecordRequest

	waitTimeSeconds int64
}

type SqsOpts func(consumer *SqsListener) error

func NewSqsConsumer(sqsConf conf.SqsConfig, provider credentials.Provider, reqChan chan common.UpdateRecordRequest, opts ...SqsOpts) (*SqsListener, error) {
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

	awsConf := &aws.Config{
		Region: aws.String(sqsConf.Region),
	}

	if provider != nil {
		log.Info().Str("component", "sqs").Msg("Building AWS client using given credentials provider")
		awsConf.Credentials = credentials.NewCredentials(provider)
	}
	awsSession := session.Must(session.NewSession(awsConf))

	ret.client = sqs.New(awsSession)
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
	result, err := h.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(h.queueUrl),
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(h.waitTimeSeconds),
	})
	if err != nil {
		return err
	}

	var errs error
	for _, message := range result.Messages {
		if err := h.handleMessage(message); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}

func (h *SqsListener) handleMessage(message *sqs.Message) error {
	defer func() {
		// the client is not going to stop ip update requests as long the ip has not been updated, so we have the luxury
		// to not care about edge cases too much and delete the message after receiving it.
		log.Debug().Str("component", "sqs").Str("message_id", *message.MessageId).Msg("Deleting message from queue")
		_, err := h.client.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(h.queueUrl),
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			log.Error().Err(err).Str("component", "sqs").Str("message_id", *message.MessageId).Msg("Could not delete message from queue")
		}
		metrics.SqsApiCalls.WithLabelValues("delete_message").Inc()
	}()

	if message == nil || message.Body == nil {
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
	var env common.UpdateRecordRequest
	err := json.Unmarshal(msg, &env)
	if err != nil {
		metrics.MessageParsingFailed.Inc()
		log.Warn().Str("component", "sqs").Err(err).Msg("Message parsing failed")
		return err
	}

	h.requests <- env
	return nil
}
