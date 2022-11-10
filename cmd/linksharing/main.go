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
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"storj.io/common/errs2"
	"storj.io/common/fpath"
	"storj.io/gateway-mt/pkg/authclient"
	"storj.io/gateway-mt/pkg/httpserver"
	"storj.io/gateway-mt/pkg/linksharing"
	"storj.io/gateway-mt/pkg/linksharing/sharing"
	"storj.io/private/cfgstruct"
	"storj.io/private/process"
	"storj.io/uplink"
)

// LinkSharing defines link sharing configuration.
//
// TODO(artur): some of these options could be grouped, e.g. into Security.
type LinkSharing struct {
	Address                string        `user:"true" help:"public address to listen on" default:":20020"`
	AddressTLS             string        `user:"true" help:"public tls address to listen on" default:":20021"`
	LetsEncrypt            bool          `user:"true" help:"use lets-encrypt to handle TLS certificates" default:"false"`
	InsecureDisableTLS     bool          `user:"true" help:"listen using insecure connections only" releaseDefault:"false" devDefault:"true"`
	CertFile               string        `user:"true" help:"server certificate file"`
	KeyFile                string        `user:"true" help:"server key file"`
	PublicURL              string        `user:"true" help:"comma separated list of public urls for the server" devDefault:"http://localhost:20020" releaseDefault:""`
	GeoLocationDB          string        `user:"true" help:"maxmind database file path"`
	TxtRecordTTL           time.Duration `user:"true" help:"max ttl (seconds) for website hosting txt record cache" devDefault:"10s" releaseDefault:"1h"`
	AuthService            authclient.Config
	DNSServer              string        `user:"true" help:"dns server address to use for TXT resolution" default:"1.1.1.1:53"`
	StaticSourcesPath      string        `user:"true" help:"the path to where web assets are located" default:"./pkg/linksharing/web/static"`
	Templates              string        `user:"true" help:"the path to where renderable templates are located" default:"./pkg/linksharing/web"`
	LandingRedirectTarget  string        `user:"true" help:"the url to redirect empty requests to" default:"https://www.storj.io/"`
	RedirectHTTPS          bool          `user:"true" help:"redirect to HTTPS" devDefault:"false" releaseDefault:"true"`
	DialTimeout            time.Duration `help:"timeout for dials" default:"10s"`
	UseQosAndCC            bool          `user:"true" help:"use congestion control and QOS settings" default:"true"`
	ClientTrustedIPSList   []string      `user:"true" help:"list of clients IPs (comma separated) which are trusted; usually used when the service run behinds gateways, load balancers, etc."`
	UseClientIPHeaders     bool          `user:"true" help:"use the headers sent by the client to identify its IP. When true the list of IPs set by --client-trusted-ips-list, when not empty, is used" default:"true"`
	StandardRendersContent bool          `user:"true" help:"enable standard (non-hosting) requests to render content and not only download it" default:"false"`
	StandardViewsHTML      bool          `user:"true" help:"serve HTML as text/html instead of text/plain for standard (non-hosting) requests" default:"false"`
	ConnectionPool         connectionPoolConfig
}

// connectionPoolConfig is a config struct for configuring RPC connection pool options.
type connectionPoolConfig struct {
	Capacity       int           `user:"true" help:"RPC connection pool capacity" default:"100"`
	KeyCapacity    int           `user:"true" help:"RPC connection pool key capacity" default:"5"`
	IdleExpiration time.Duration `user:"true" help:"RPC connection pool idle expiration" default:"2m0s"`
}

var (
	rootCmd = &cobra.Command{
		Use:   "link sharing service",
		Short: "Link Sharing Service",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the link sharing service",
		RunE:  cmdRun,
	}
	setupCmd = &cobra.Command{
		Use:         "setup",
		Short:       "Create config files",
		RunE:        cmdSetup,
		Annotations: map[string]string{"type": "setup"},
	}

	runCfg   LinkSharing
	setupCfg LinkSharing

	confDir string
)

func init() {
	defaultConfDir := fpath.ApplicationDir("storj", "linksharing")
	cfgstruct.SetupFlag(zap.L(), rootCmd, &confDir, "config-dir", defaultConfDir, "main directory for link sharing configuration")
	defaults := cfgstruct.DefaultsFlag(rootCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(setupCmd)
	process.Bind(runCmd, &runCfg, defaults, cfgstruct.ConfDir(confDir))
	process.Bind(setupCmd, &setupCfg, defaults, cfgstruct.ConfDir(confDir), cfgstruct.SetupMode())
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx, cancel := process.Ctx(cmd)
	defer cancel()

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

	publicURLs := strings.Split(runCfg.PublicURL, ",")

	var tlsConfig *httpserver.TLSConfig
	if !runCfg.InsecureDisableTLS {
		tlsConfig = &httpserver.TLSConfig{
			LetsEncrypt: runCfg.LetsEncrypt,
			CertFile:    runCfg.CertFile,
			KeyFile:     runCfg.KeyFile,
			PublicURLs:  publicURLs,
			ConfigDir:   confDir,
		}
	}

	peer, err := linksharing.New(log, linksharing.Config{
		Server: httpserver.Config{
			Name:            "Link Sharing",
			Address:         runCfg.Address,
			AddressTLS:      runCfg.AddressTLS,
			TrafficLogging:  true,
			TLSConfig:       tlsConfig,
			ShutdownTimeout: -1,
		},
		Handler: sharing.Config{
			URLBases:               publicURLs,
			Templates:              runCfg.Templates,
			StaticSourcesPath:      runCfg.StaticSourcesPath,
			RedirectHTTPS:          runCfg.RedirectHTTPS,
			LandingRedirectTarget:  runCfg.LandingRedirectTarget,
			TxtRecordTTL:           runCfg.TxtRecordTTL,
			AuthServiceConfig:      runCfg.AuthService,
			DNSServer:              runCfg.DNSServer,
			ConnectionPool:         sharing.ConnectionPoolConfig(runCfg.ConnectionPool),
			UseQosAndCC:            runCfg.UseQosAndCC,
			ClientTrustedIPsList:   runCfg.ClientTrustedIPSList,
			UseClientIPHeaders:     runCfg.UseClientIPHeaders,
			StandardViewsHTML:      runCfg.StandardViewsHTML,
			StandardRendersContent: runCfg.StandardRendersContent,
			Uplink: &uplink.Config{
				UserAgent:   "linksharing",
				DialTimeout: runCfg.DialTimeout,
			},
		},
		GeoLocationDB: runCfg.GeoLocationDB,
	})
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

func cmdSetup(cmd *cobra.Command, args []string) (err error) {
	setupDir, err := filepath.Abs(confDir)
	if err != nil {
		return err
	}

	valid, _ := fpath.IsValidSetupDir(setupDir)
	if !valid {
		return fmt.Errorf("link sharing configuration already exists (%v)", setupDir)
	}

	err = os.MkdirAll(setupDir, 0700)
	if err != nil {
		return err
	}

	return process.SaveConfig(cmd, filepath.Join(setupDir, "config.yaml"))
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
