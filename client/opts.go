package client

import (
	"errors"
	"time"
)

func WithInterval(interval time.Duration) func(c *Client) error {
	return func(c *Client) error {
		if interval < 10*time.Second {
			return errors.New("interval must not be < 10s")
		}

		if interval > 5*time.Minute {
			return errors.New("interval must not be > 5m")
		}

		c.resolveInterval = interval
		return nil
	}
}
