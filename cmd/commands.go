package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/signal"
	"syscall"

	"github.com/soerenschneider/dyndns/v2/internal"
	"github.com/soerenschneider/dyndns/v2/internal/conf"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/internal/verification"
	"github.com/spf13/cobra"
)

type globalOpts struct {
	configPath string
	debug      bool
}

// modeOpts are the flags of the subcommands that run dyndns in a mode.
type modeOpts struct {
	once            bool
	forceSendUpdate bool
}

func newRootCmd() *cobra.Command {
	opts := &globalOpts{}
	root := &cobra.Command{
		Use:          "dyndns",
		Short:        "Updates DNS records of hosts that don't have a static IP",
		Version:      versionString(),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		// Without a subcommand there is no mode to run in.
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = cmd.Help()
			return errors.New("a subcommand is required")
		},
	}
	root.SetVersionTemplate("{{.Version}}\n")
	root.PersistentFlags().StringVar(&opts.configPath, "config", defaultConfigPath, "Path to the config file")
	root.PersistentFlags().BoolVar(&opts.debug, "debug", false, "Print debug logs")

	root.AddCommand(
		newClientCmd(opts),
		newServerCmd(opts),
		newEdgeCmd(opts),
		newKeygenCmd(),
		newVersionCmd(),
	)

	return root
}

func newClientCmd(opts *globalOpts) *cobra.Command {
	mode := &modeOpts{}
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Detect the public IP and send signed update requests to the server",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMode(opts, internal.ClientMode, mode)
		},
	}
	cmd.Flags().BoolVar(&mode.once, "once", false, "Resolve and update the record once, then exit")
	cmd.Flags().BoolVar(&mode.forceSendUpdate, "force", false, "Force sending an update request at start")
	return cmd
}

func newServerCmd(opts *globalOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Verify update requests from clients and apply them to Route 53",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMode(opts, internal.ServerMode, &modeOpts{})
		},
	}
}

func newEdgeCmd(opts *globalOpts) *cobra.Command {
	mode := &modeOpts{}
	cmd := &cobra.Command{
		Use:   "edge",
		Short: "Run client and server in one process, updating Route 53 directly",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMode(opts, internal.EdgeMode, mode)
		},
	}
	cmd.Flags().BoolVar(&mode.forceSendUpdate, "force", false, "Force sending an update request at start")
	return cmd
}

func newKeygenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keygen",
		Short: "Generate a keypair for a client and print it as JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return generateKeypair(cmd.OutOrStdout())
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), versionString())
		},
	}
}

func versionString() string {
	return fmt.Sprintf("%s (commit: %s) go%s", internal.BuildVersion, internal.CommitHash, internal.GoVersion)
}

func runMode(opts *globalOpts, mode string, modeOpts *modeOpts) error {
	initLogging(opts.debug)

	metrics.Version.WithLabelValues(internal.BuildVersion, internal.CommitHash, internal.GoVersion).Set(1)
	metrics.ProcessStartTime.SetToCurrentTime()

	config, err := loadConfig(opts.configPath, mode)
	if err != nil {
		return err
	}
	config.Client.Once = modeOpts.once
	config.Client.ForceSendUpdate = modeOpts.forceSendUpdate

	// Cancelled on SIGINT and SIGTERM, which shuts dyndns down gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if !modeOpts.once {
		go func() {
			err := metrics.StartMetricsServer(ctx, config.MetricsListener)
			dieOnError(err, "could not start metrics server")
		}()
		go metrics.StartHeartbeat(ctx)
	}

	app, err := buildApp(ctx, config)
	if err != nil {
		return fmt.Errorf("could not build dependencies: %w", err)
	}

	return run(ctx, app)
}

func loadConfig(path, mode string) (*conf.Conf, error) {
	config, err := conf.ReadConfig(path)
	if err != nil {
		// Without a config file at the default path, dyndns is configured via
		// environment variables only.
		if path != defaultConfigPath {
			return nil, fmt.Errorf("couldn't read config file: %w", err)
		}
		config = conf.GetDefaultConfig()
	}
	config.Mode = mode

	if err := conf.ParseEnvVariables(config); err != nil {
		return nil, fmt.Errorf("could not parse env variables: %w", err)
	}

	if err := conf.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func generateKeypair(out io.Writer) error {
	keypair, err := verification.NewKeyPair()
	if err != nil {
		return fmt.Errorf("can not create keypair: %w", err)
	}

	jsonEncoded, err := keypair.AsJson()
	if err != nil {
		return fmt.Errorf("could not marshal keypair: %w", err)
	}

	_, err = fmt.Fprintf(out, "%s\n", jsonEncoded)
	return err
}
