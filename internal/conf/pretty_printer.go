package conf

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

var SensitiveFields = []string{"KeyPair", "SmtpPassword"}

func PrintFields(data any, ignoredKeys ...string) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Dereference the pointer
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		if isEmptyOrNil(value) {
			continue
		}

		if slices.Contains(ignoredKeys, field.Name) {
			log.Info().Msgf("%s=%s", field.Name, "*** (redacted)")
		} else {
			log.Info().Msgf("%s=%s", field.Name, fieldValueToString(field.Name, value))
		}
	}
}

// TODO: replace with a generic slice function in go > 1.20
func sliceContains(slice []string, val string) bool {
	val = strings.ToLower(val)
	for _, entry := range slice {
		if strings.ToLower(entry) == val {
			return true
		}
	}
	return false
}

func fieldValueToString(nam string, value reflect.Value) string {
	if value.CanInterface() {
		if value.Kind() == reflect.Ptr {
			// Handle the case where value is a pointer
			if value.IsNil() {
				return "<nil>"
			}
			value = value.Elem()
		}

		if stringer, ok := value.Interface().(fmt.Stringer); ok {
			return stringer.String()
		}

		// Check if the address of the struct implements fmt.Stringer
		if value.Kind() == reflect.Struct {
			ptrValue := value.Addr()
			if stringer, ok := ptrValue.Interface().(fmt.Stringer); ok {
				return stringer.String()
			}
		}
	}
	return fmt.Sprintf("%v", value.Interface())
}

func isEmptyOrNil(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return value.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	default:
		return false
	}
}
