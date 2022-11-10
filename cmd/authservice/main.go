// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
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
	"time"

	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"storj.io/common/errs2"
	"storj.io/common/fpath"
	"storj.io/gateway-mt/internal/register"
	"storj.io/gateway-mt/pkg/auth"
	"storj.io/private/cfgstruct"
	"storj.io/private/process"
)

var (
	rootCmd = &cobra.Command{
		Use:   "authservice",
		Short: "Auth Service (used mainly with Gateway-MT and Link Sharing Service)",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the service",
		RunE:  cmdRun,
	}
	runMigrationCmd = &cobra.Command{
		Use:   "migration",
		Short: "Create or update the database schema, then quit",
		Args:  cobra.ExactArgs(0),
		RunE:  cmdMigrationRun,
	}
	setupCmd = &cobra.Command{
		Use:         "setup",
		Short:       "Create configuration file",
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"type": "setup"},
		RunE:        cmdSetup,
		Hidden:      true,
	}
	registerCmd = &cobra.Command{
		Use:    "register",
		Short:  "Register credentials @ authservice via HTTP or DRPC",
		Args:   cobra.ExactArgs(1),
		RunE:   cmdRegister,
		Hidden: true,
	}

	runCfg   auth.Config
	setupCfg auth.Config

	confDir string

	registerCfg struct {
		Address   string `help:"authservice to register access to" dev:"drpc://localhost:20002" release:"drpcs://auth.storjshare.io:7777"`
		Public    bool   `help:"whether access grant can be retrieved from authservice by providing only Access Key ID without Secret Access Key" default:"false"`
		FormatEnv bool   `help:"environmental-variable format of credentials; for using in scripts" default:"false"`
	}
)

func init() {
	defaultConfDir := fpath.ApplicationDir("storj", "authservice")
	cfgstruct.SetupFlag(zap.L(), rootCmd, &confDir, "config-dir", defaultConfDir, "main directory for configuration")
	defaults := cfgstruct.DefaultsFlag(rootCmd)

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(registerCmd)

	runCmd.AddCommand(runMigrationCmd)

	process.Bind(runCmd, &runCfg, defaults, cfgstruct.ConfDir(confDir))
	process.Bind(runMigrationCmd, &runCfg, defaults, cfgstruct.ConfDir(confDir))
	process.Bind(setupCmd, &setupCfg, defaults, cfgstruct.ConfDir(confDir), cfgstruct.SetupMode())
	process.Bind(registerCmd, &registerCfg, defaults)
}

func main() {
	process.Exec(rootCmd)
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx, _ := process.Ctx(cmd)

	if runCfg.Migration {
		if err = cmdMigrationRun(cmd, args); err != nil {
			return err
		}
	}

	log := zap.L()

	if err := process.InitMetricsWithHostname(ctx, log, nil); err != nil {
		return errs.New("failed to initialize telemetry batcher: %w", err)
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

	p, err := auth.New(ctx, log, runCfg, confDir)
	if err != nil {
		return err
	}

	defer func() {
		err = errs.Combine(err, p.Close())
	}()

	return errs2.IgnoreCanceled(p.Run(ctx))
}

func cmdMigrationRun(cmd *cobra.Command, _ []string) (err error) {
	ctx, _ := process.Ctx(cmd)

	migrationCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	kv, err := auth.OpenKV(migrationCtx, zap.L().Named("migration"), runCfg)
	if err != nil {
		return errs.Wrap(err)
	}
	defer func() { err = errs.Combine(err, kv.Close()) }()

	var g errgroup.Group

	g.Go(func() error {
		return errs2.IgnoreCanceled(kv.Run(migrationCtx))
	})

	migrator, ok := kv.(interface {
		MigrateToLatest(ctx context.Context) error
	})
	if !ok {
		return errs.New("database backend %T does not support migrations", kv)
	}

	if err = migrator.MigrateToLatest(migrationCtx); err != nil {
		return errs.Wrap(err)
	}

	cancel()

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

func cmdRegister(cmd *cobra.Command, args []string) error {
	ctx, _ := process.Ctx(cmd)

	res, err := register.Access(ctx, registerCfg.Address, args[0], registerCfg.Public)
	if err != nil {
		return err
	}

	if registerCfg.FormatEnv {
		fmt.Printf("AWS_ACCESS_KEY_ID=%s\nAWS_SECRET_ACCESS_KEY=%s\nAWS_ENDPOINT=%s\n",
			res.AccessKeyID, res.SecretKey, res.Endpoint)
	} else {
		fmt.Println(res)
	}

	return nil
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
