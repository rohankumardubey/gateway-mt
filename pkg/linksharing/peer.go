// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package linksharing

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"os"
	"runtime"
	pkgmiddleware "storj.io/gateway-mt/pkg/middleware"

	"github.com/oschwald/maxminddb-golang"
	"github.com/zeebo/errs"
	"go.uber.org/zap"

	"storj.io/gateway-mt/pkg/httpserver"
	"storj.io/gateway-mt/pkg/linksharing/objectmap"
	"storj.io/gateway-mt/pkg/linksharing/sharing"
)

// Config contains configurable values for sno registration Peer.
type Config struct {
	Server  httpserver.Config
	Handler sharing.Config

	// Maxmind geolocation database path.
	GeoLocationDB string
}

// Peer is the representation of a Linksharing service itself.
//
// architecture: Peer
type Peer struct {
	Log    *zap.Logger
	Mapper *objectmap.IPDB
	Server *httpserver.Server
}

// New is a constructor for Linksharing Peer.
func New(log *zap.Logger, config Config) (_ *Peer, err error) {
	peer := &Peer{
		Log: log,
	}

	_, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(context.Background(), "Linksharing Startup")
	span.AddEvent("linksharing service starting")
	span.End()

	if config.GeoLocationDB != "" {
		reader, err := maxminddb.Open(config.GeoLocationDB)
		if err != nil {
			return nil, errs.New("unable to open geo location db: %w", err)
		}
		peer.Mapper = objectmap.NewIPDB(reader)
	}

	handle, err := sharing.NewHandler(log, peer.Mapper, config.Handler)
	if err != nil {
		return nil, errs.New("unable to create handler: %w", err)
	}

	handleWithTracing := otelhttp.NewHandler(handle, "")

	//handleWithTracing := http.TraceHandler(handle, mon)
	//instrumentedHandle := middleware.Metrics("linksharing", handleWithTracing)
	handleWithRequestID := pkgmiddleware.AddRequestID(handleWithTracing)

	peer.Server, err = httpserver.New(log, handleWithRequestID, config.Server)
	if err != nil {
		return nil, errs.New("unable to create httpserver: %w", err)
	}

	return peer, nil
}

// Run starts the server.
func (peer *Peer) Run(ctx context.Context) (err error) {
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()

	return peer.Server.Run(ctx)
}

// Close shuts down the server and all underlying resources.
func (peer *Peer) Close() error {
	var errlist errs.Group

	if peer.Server != nil {
		errlist.Add(peer.Server.Shutdown())
	}

	if peer.Mapper != nil {
		errlist.Add(peer.Mapper.Close())
	}

	return errlist.Err()
}
