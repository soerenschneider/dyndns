package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soerenschneider/dyndns/v2/internal"
)

func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var out bytes.Buffer
	cmd := newRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	return out.String(), err
}

func TestRootCmd_RequiresSubcommand(t *testing.T) {
	if _, err := execute(t); err == nil {
		t.Fatal("expected an error without a subcommand")
	}
}

func TestModeFlags(t *testing.T) {
	tests := []struct {
		args        []string
		wantUnknown bool
	}{
		{args: []string{"server", "--once"}, wantUnknown: true},
		{args: []string{"server", "--force"}, wantUnknown: true},
		{args: []string{"edge", "--once"}, wantUnknown: true},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, err := execute(t, tt.args...)
			if gotUnknown := err != nil && strings.Contains(err.Error(), "unknown flag"); gotUnknown != tt.wantUnknown {
				t.Errorf("Execute() error = %v, want unknown flag %v", err, tt.wantUnknown)
			}
		})
	}
}

func TestKeygenCmd(t *testing.T) {
	out, err := execute(t, "keygen")
	if err != nil {
		t.Fatal(err)
	}

	var keypair struct {
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}
	if err := json.Unmarshal([]byte(out), &keypair); err != nil {
		t.Fatalf("keygen output is not a json keypair: %v: %q", err, out)
	}
	if keypair.PublicKey == "" || keypair.PrivateKey == "" {
		t.Errorf("keygen output misses a key: %q", out)
	}
}

func TestVersionCmd(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--version"}} {
		out, err := execute(t, args...)
		if err != nil {
			t.Fatal(err)
		}
		if out != versionString()+"\n" {
			t.Errorf("%v printed %q, want %q", args, out, versionString())
		}
	}
}

func TestLoadConfig_SetsModeFromSubcommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "server:\n  hosted_zone_id: x\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	config, err := loadConfig(path, internal.ServerMode)
	if err != nil {
		t.Fatal(err)
	}
	if config.Mode != internal.ServerMode {
		t.Errorf("Mode = %q, want %q", config.Mode, internal.ServerMode)
	}

	// The client section is not needed in server mode, but is in client mode.
	if _, err := loadConfig(path, internal.ClientMode); err == nil {
		t.Error("loadConfig() in client mode without a client section succeeded")
	}
}
