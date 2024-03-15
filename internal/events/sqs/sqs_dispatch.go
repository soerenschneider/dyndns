package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/conf"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"go.uber.org/multierr"
)

type SqsListener struct {
	client   *sqs.SQS
	queueUrl string
	requests chan common.UpdateRecordRequest
}

type SqsOpts func(consumer *SqsListener) error

func NewSqsConsumer(sqsConf conf.SqsConfig, provider credentials.Provider, reqChan chan common.UpdateRecordRequest, opts ...SqsOpts) (*SqsListener, error) {
	if reqChan == nil {
		return nil, errors.New("empty chan provided")
	}

	ret := &SqsListener{
		queueUrl: sqsConf.SqsQueue,
		requests: reqChan,
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
		log.Info().Msg("Building AWS client using given credentials provider")
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
		log.Error().Err(err).Msg("Fetching SQS messages failed")
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Received signal, stopping SQS listener")
			wg.Done()
			return nil
		case <-ticker.C:
			if err := h.fetchMessages(ctx); err != nil {
				log.Error().Err(err).Msg("Fetching SQS messages failed")
			}
		}
	}
}

func (h *SqsListener) fetchMessages(ctx context.Context) error {
	log.Debug().Msg("Trying to receive SQS messages")
	result, err := h.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(h.queueUrl),
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(20),
	})

	if err != nil {
		return err
	}

	for _, message := range result.Messages {
		if err := h.handleMessage(message); err != nil {
			log.Error().Err(err).Msgf("handling received message failed: %v", err)
		}
	}

	return nil
}

func (h *SqsListener) handleMessage(message *sqs.Message) error {
	if message == nil || message.Body == nil {
		log.Warn().Msg("Received empty SQS message")
		return nil
	}

	err := h.dispatch([]byte(*message.Body))
	if err != nil {
		return err
	}

	_, err = h.client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(h.queueUrl),
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		return fmt.Errorf("could not delete message %q from queue: %w", *message.MessageId, err)
	}

	return nil
}

func (h *SqsListener) dispatch(msg []byte) error {
	var env common.UpdateRecordRequest
	err := json.Unmarshal(msg, &env)
	if err != nil {
		metrics.MessageParsingFailed.Inc()
		log.Warn().Msgf("Can't parse message: %v", err)
		return err
	}

	h.requests <- env
	return nil
}
