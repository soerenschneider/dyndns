package conf

import (
	"net/url"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/conf/hybrid"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func ValidateConfig[T any](c T) error {
	once.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		if err := validate.RegisterValidation("required_for_mode", requiredForMode); err != nil {
			log.Fatal().Err(err).Msg("could not build custom validation 'required_for_mode'")
		}

		if err := validate.RegisterValidation("addrfamilies", validateAddrFamilies); err != nil {
			log.Fatal().Err(err).Msg("could not build custom validation 'addrfamilies'")
		}
		if err := validate.RegisterValidation("broker", validateBrokers); err != nil {
			log.Fatal().Err(err).Msg("could not build custom validation 'validateBrokers'")
		}
		if err := validate.RegisterValidation("nats_url", validateNatsUrl); err != nil {
			log.Fatal().Err(err).Msg("could not build custom validation 'nats_url'")
		}
		if err := validate.RegisterValidation("nats_subject", validateNatsSubject); err != nil {
			log.Fatal().Err(err).Msg("could not build custom validation 'nats_subject'")
		}

		validate.RegisterStructValidation(validateEmailConfig, hybrid.EmailConfig{})
	})

	return validate.Struct(c)
}

const modeField = "Mode"

// requiredForMode implements `required_for_mode=<mode>[ <mode>...]`.
// The field must be non-empty when Conf.Mode matches one of the listed modes.
func requiredForMode(fl validator.FieldLevel) bool {
	top := fl.Top()
	for top.Kind() == reflect.Pointer || top.Kind() == reflect.Interface {
		if top.IsNil() {
			return true
		}
		top = top.Elem()
	}
	if top.Kind() != reflect.Struct {
		return true
	}

	f := top.FieldByName(modeField)
	if !f.IsValid() || f.Kind() != reflect.String {
		// Not validating from Conf — no mode to check against, so don't fail.
		return true
	}
	mode := f.String()

	var applies bool
	for _, m := range strings.Fields(fl.Param()) {
		if m == mode {
			applies = true
			break
		}
	}
	if !applies {
		return true
	}
	return hasValue(fl.Field())
}

func hasValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Slice, reflect.Map:
		return !field.IsNil() && field.Len() > 0
	case reflect.Pointer, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && !field.IsZero()
	}
}

//nolint:cyclop
func validateEmailConfig(sl validator.StructLevel) {
	config := sl.Current().Interface().(hybrid.EmailConfig)

	if config.SmtpPort == 0 && len(config.SmtpHost)+len(config.SmtpUsername)+len(config.SmtpUsernameFile)+len(config.SmtpPasswordFile)+len(config.SmtpPassword)+len(config.From)+len(config.FromFile)+len(config.To)+len(config.ToFile) == 0 {
		return
	}

	if config.SmtpUsername == "" && config.SmtpUsernameFile == "" {
		sl.ReportError(config.SmtpUsername, "SmtpUsername", "SmtpUsername", "usernameOrFileRequired", "")
		sl.ReportError(config.SmtpUsernameFile, "SmtpUsernameFile", "SmtpUsernameFile", "usernameOrFileRequired", "")
	}
	if config.SmtpPassword == "" && config.SmtpPasswordFile == "" {
		sl.ReportError(config.SmtpPassword, "SmtpPassword", "SmtpPassword", "passwordOrFileRequired", "")
		sl.ReportError(config.SmtpPasswordFile, "SmtpPasswordFile", "SmtpPasswordFile", "passwordOrFileRequired", "")
	}

	if config.From == "" && config.FromFile == "" {
		sl.ReportError(config.From, "From", "From", "requiredWithoutFromFile", "")
		sl.ReportError(config.FromFile, "FromFile", "FromFile", "requiredWithoutFrom", "")
	}

	if len(config.To) == 0 && config.ToFile == "" {
		sl.ReportError(config.To, "To", "To", "requiredWithoutToFile", "")
		sl.ReportError(config.ToFile, "ToFile", "ToFile", "requiredWithoutTo", "")
	}
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

		if str != internal.AddrFamilyIpv4 && str != internal.AddrFamilyIpv6 {
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
		if !ok || !IsValidMqttUrl(broker) {
			return false
		}
	}

	return true
}

func IsValidMqttUrl(input string) bool {
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

func validateNatsUrl(fl validator.FieldLevel) bool {
	// Get the field value and check if it's a slice
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}

	// Convert to string and check its value
	url, ok := field.Interface().(string)
	if !ok || !IsValidNatsUrl(url) {
		return false
	}

	return true
}

func IsValidNatsUrl(input string) bool {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}

	u, err := url.Parse(input)
	if err != nil || u.Scheme != "nats" || u.Host == "" {
		return false
	}

	return true
}

func validateNatsSubject(fl validator.FieldLevel) bool {
	// Get the field value and check if it's a slice
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}

	// Convert to string and check its value
	url, ok := field.Interface().(string)
	if !ok || !IsValidNatsSubject(url) {
		return false
	}

	return true
}

func IsValidNatsSubject(subject string) bool {
	if subject == "" {
		return false
	}

	if strings.Contains(subject, ">") || strings.Contains(subject, "*") {
		return false
	}

	tokens := strings.Split(subject, ".")
	for _, token := range tokens {
		if token == "" {
			return false
		}

		if strings.ContainsAny(token, " \t\r\n") {
			return false
		}

		for _, r := range token {
			if unicode.IsControl(r) {
				return false
			}
		}
	}

	return true
}
