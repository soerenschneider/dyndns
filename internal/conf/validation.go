package conf

import (
	"net/url"
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

const (
	AddrFamilyIpv6 = "ip6"
	AddrFamilyIpv4 = "ip4"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func ValidateConfig[T any](c T) error {
	once.Do(func() {
		validate = validator.New()
		if err := validate.RegisterValidation("addrfamilies", validateAddrFamilies); err != nil {
			log.Fatal().Err(err).Msg("could not build custom validation 'addrfamilies'")
		}
		if err := validate.RegisterValidation("broker", validateBrokers); err != nil {
			log.Fatal().Err(err).Msg("could not build custom validation 'validateBrokers'")
		}

		validate.RegisterStructValidation(EmailConfigStructLevelValidation, EmailConfig{})
	})

	return validate.Struct(c)
}

func validateAddrFamilies(fl validator.FieldLevel) bool {
	// Get the field value and check if it's a slice
	field := fl.Field()
	if field.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < field.Len(); i++ {
		item := field.Index(i)

		// Convert to string and check its value
		str, ok := item.Interface().(string)
		if !ok {
			return false
		}

		if str != AddrFamilyIpv4 && str != AddrFamilyIpv6 {
			return false
		}
	}

	return true
}

func validateBrokers(fl validator.FieldLevel) bool {
	// Get the field value and check if it's a slice
	field := fl.Field()
	if field.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < field.Len(); i++ {
		item := field.Index(i)

		// Convert to string and check its value
		broker, ok := item.Interface().(string)
		if !ok || !IsValidUrl(broker) {
			return false
		}
	}

	return true
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
