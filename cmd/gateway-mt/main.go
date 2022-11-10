// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zeebo/errs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"storj.io/common/errs2"
	"storj.io/common/fpath"
	"storj.io/gateway-mt/pkg/authclient"
	"storj.io/gateway-mt/pkg/server"
	"storj.io/gateway-mt/pkg/trustedip"
	"storj.io/private/cfgstruct"
	"storj.io/private/process"
)

var (
	// Error is the default gateway setup errs class.
	Error = errs.Class("gateway setup")

	// rootCmd represents the base gateway command when called without any
	// subcommands.
	rootCmd = &cobra.Command{
		Use:   "gateway",
		Short: "Multi-tenant, S3-compatible gateway",
		Args:  cobra.OnlyValidArgs,
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the service",
		Args:  cobra.ExactArgs(0),
		RunE:  cmdRun,
	}
	setupCmd = &cobra.Command{
		Use:         "setup",
		Short:       "Create configuration file",
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"type": "setup"},
		RunE:        cmdSetup,
		Hidden:      true,
	}

	runCfg   server.Config
	setupCfg server.Config

	confDir string
)

func init() {
	defaultConfDir := fpath.ApplicationDir("storj", "gateway")
	cfgstruct.SetupFlag(zap.L(), rootCmd, &confDir, "config-dir", defaultConfDir, "main directory for gateway configuration")
	defaults := cfgstruct.DefaultsFlag(rootCmd)

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)

	process.Bind(runCmd, &runCfg, defaults, cfgstruct.ConfDir(confDir))
	process.Bind(setupCmd, &setupCfg, defaults, cfgstruct.ConfDir(confDir), cfgstruct.SetupMode())

	// The loop below sets all flags in GatewayFlags to show up without the
	// `--advanced` flag until we decide which flags we want to hide.
	runCmd.Flags().VisitAll(func(f *pflag.Flag) {
		cfgstruct.SetBoolAnnotation(runCmd.Flags(), f.Name, cfgstruct.BasicHelpAnnotationName, true)
	})

	rootCmd.PersistentFlags().BoolVar(new(bool), "advanced", false, "if used in with -h, print advanced flags help")
	cfgstruct.SetBoolAnnotation(rootCmd.PersistentFlags(), "advanced", cfgstruct.BasicHelpAnnotationName, true)
	cfgstruct.SetBoolAnnotation(rootCmd.PersistentFlags(), "config-dir", cfgstruct.BasicHelpAnnotationName, true)
	setUsageFunc(rootCmd)
}

func cmdRun(cmd *cobra.Command, _ []string) (err error) {
	ctx, cancel := process.Ctx(cmd)
	defer cancel()

	log := zap.L()

	if err := process.InitMetricsWithHostname(ctx, log, nil); err != nil {
		return errs.New("Failed to initialize telemetry batcher: %w", err)
	}

	// setup tracing
	err = initTracer()
	if err != nil {
		return errs.New("failed to initialize open telemetry: %w", err)
	}
	err = initMeter()
	if err != nil {
		return errs.New("failed to initialize open telemetry: %w", err)
	}

	// setup environment variables for Minio
	validate := func(value, configName string) {
		if value == "" {
			err = errs.Combine(err, Error.New("required parameter --%s not set", configName))
		}
	}
	set := func(value, envName string) {
		err = errs.Combine(err, Error.Wrap(os.Setenv(envName, value)))
	}
	validate(runCfg.DomainName, "domain-name")
	set(runCfg.DomainName, "MINIO_DOMAIN") // MINIO_DOMAIN supports comma-separated domains.
	set("off", "MINIO_BROWSER")
	set("dummy-key-to-satisfy-minio", "MINIO_ACCESS_KEY")
	set("dummy-key-to-satisfy-minio", "MINIO_SECRET_KEY")
	if err != nil {
		return err
	}

	log.Info("Starting Storj DCS S3 Gateway")

	if runCfg.InsecureLogAll {
		log.Info("Insecurely logging all errors, paths, and headers")
	}

	var trustedClientIPs trustedip.List

	if runCfg.UseClientIPHeaders {
		if len(runCfg.ClientTrustedIPSList) > 0 {
			trustedClientIPs = trustedip.NewList(runCfg.ClientTrustedIPSList...)
		} else {
			trustedClientIPs = trustedip.NewListTrustAll()
		}
	} else {
		trustedClientIPs = trustedip.NewListUntrustAll()
	}

	corsAllowedOrigins := strings.Split(runCfg.CorsOrigins, ",")

	if err := runCfg.Auth.Validate(); err != nil {
		return err
	}
	peer, err := server.New(runCfg, log, trustedClientIPs, corsAllowedOrigins,
		authclient.New(runCfg.Auth), strings.Split(runCfg.DomainName, ","), runCfg.ConcurrentAllowed)
	if err != nil {
		return err
	}

	var g errgroup.Group

	g.Go(func() error {
		<-ctx.Done()
		return errs2.IgnoreCanceled(peer.Close())
	})

	g.Go(func() error {
		return errs2.IgnoreCanceled(peer.Run(ctx))
	})

	return g.Wait()
}

func cmdSetup(cmd *cobra.Command, _ []string) error {
	setupDir, err := filepath.Abs(confDir)
	if err != nil {
		return err
	}

	valid, _ := fpath.IsValidSetupDir(setupDir)
	if !valid {
		return errs.New("configuration already exists (%v)", setupDir)
	}

	if err = os.MkdirAll(setupDir, 0700); err != nil {
		return err
	}

	return process.SaveConfig(cmd, filepath.Join(setupDir, "config.yaml"))
}

/*
`setUsageFunc` is a bit unconventional but cobra didn't leave much room for
extensibility here. `cmd.SetUsageTemplate` is fairly useless for our case without
the ability to add to the template's function map (see: https://golang.org/pkg/text/template/#hdr-Functions).

Because we can't alter what `cmd.Usage` generates, we have to edit it afterwards.
In order to hook this function *and* get the usage string, we have to juggle the
`cmd.usageFunc` between our hook and `nil`, so that we can get the usage string
from the default usage func.
*/
func setUsageFunc(cmd *cobra.Command) {
	if findBoolFlagEarly("advanced") {
		return
	}

	reset := func() (set func()) {
		original := cmd.UsageFunc()
		cmd.SetUsageFunc(nil)

		return func() {
			cmd.SetUsageFunc(original)
		}
	}

	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		set := reset()
		usageStr := cmd.UsageString()
		defer set()

		usageScanner := bufio.NewScanner(bytes.NewBufferString(usageStr))

		var basicFlags []string
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			basic, ok := flag.Annotations[cfgstruct.BasicHelpAnnotationName]
			if ok && len(basic) == 1 && basic[0] == "true" {
				basicFlags = append(basicFlags, flag.Name)
			}
		})

		for usageScanner.Scan() {
			line := usageScanner.Text()
			trimmedLine := strings.TrimSpace(line)

			var flagName string
			if _, err := fmt.Sscanf(trimmedLine, "--%s", &flagName); err != nil {
				fmt.Println(line)
				continue
			}

			// TODO: properly filter flags with short names
			if !strings.HasPrefix(trimmedLine, "--") {
				fmt.Println(line)
			}

			for _, basicFlag := range basicFlags {
				if basicFlag == flagName {
					fmt.Println(line)
				}
			}
		}
		return nil
	})
}

func findBoolFlagEarly(flagName string) bool {
	for i, arg := range os.Args {
		arg := arg
		argHasPrefix := func(format string, args ...interface{}) bool {
			return strings.HasPrefix(arg, fmt.Sprintf(format, args...))
		}

		if !argHasPrefix("--%s", flagName) {
			continue
		}

		// NB: covers `--<flagName> false` usage
		if i+1 != len(os.Args) {
			next := os.Args[i+1]
			if next == "false" {
				return false
			}
		}

		if !argHasPrefix("--%s=false", flagName) {
			return true
		}
	}
	return false
}

func initTracer() error {
	ctx := context.Background()

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(os.Getenv("EXPORTER_ENDPOINT")))
	sctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	traceExp, err := otlptrace.New(sctx, traceClient)
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(os.Getenv("SERVICE_NAME")),
		),
	)
	if err != nil {
		return err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)
	return nil
}

func initMeter() error {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	wrappedRegisterer := prometheus.WrapRegistererWithPrefix("gateway_", prometheus.NewRegistry())
	exporter, err := otelprom.New(otelprom.WithRegisterer(wrappedRegisterer), otelprom.WithoutUnits())
	if err != nil {
		log.Fatal(err)
	}
	global.SetMeterProvider(metric.NewMeterProvider(metric.WithReader(exporter)))

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics()
	return nil
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:9153/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":9153", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}

func main() {
	process.Exec(rootCmd)
}
