package conf

import (
	"os"
	"reflect"
	"testing"
)

func TestParseClientConfEnv(t *testing.T) {
	envKey := "DYNDNS_HTTP_DISPATCHER_CONF"
	os.Setenv(envKey, "[{\"url\":\"https://one\"}, {\"url\":\"https://two\"}]")
	// unset after running test
	defer os.Setenv(envKey, "")

	empty := GetDefaultConfig()
	err := ParseEnvVariables(empty)
	if err != nil {
		t.Fatal(err)
	}

	expected := []HttpDispatcherConfig{
		{
			Url: "https://one",
		},
		{
			Url: "https://two",
		},
	}

	if !reflect.DeepEqual(empty.Client.HttpDispatcherConf, expected) {
		t.Fatalf("expected %v, got %v", expected, empty.Client.HttpDispatcherConf)
	}
}
