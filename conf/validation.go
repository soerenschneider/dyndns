package conf

import (
	"github.com/go-playground/validator/v10"
	"net/url"
	"sync"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func ValidateConfig[T any](c T) error {
	once.Do(func() {
		validate = validator.New()
	})

	return validate.Struct(c)
}

func IsValidUrl(input string) bool {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}

	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" || u.Host == "" || u.Port() == "" {
		return false
	}

	return true
}
