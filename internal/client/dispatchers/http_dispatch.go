package dispatchers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
)

type HttpDispatch struct {
	client *http.Client
	url    string
}

func NewHttpDispatcher(url string) (*HttpDispatch, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.HTTPClient.Timeout = 5 * time.Second

	return &HttpDispatch{
		client: client.HTTPClient,
		url:    url,
	}, nil
}

func (h *HttpDispatch) Notify(msg *common.UpdateRecordRequest) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	response, err := h.client.Post(h.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != 200 {
		log.Error().Str("component", "http_dispatch").Int("status", response.StatusCode).Msg("bad request")
		return fmt.Errorf("http dispatcher received status code %d", response.StatusCode)
	}
	log.Debug().Str("component", "http_dispatch").Int("status", response.StatusCode).Msg("http dispatcher received reply")
	return nil
}
