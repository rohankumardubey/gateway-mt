// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package middleware

import (
	"context"
	"encoding/hex"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"runtime"

	"storj.io/common/grant"
	"storj.io/common/useragent"
	"storj.io/gateway-mt/pkg/trustedip"
)

// CollectEvent collects event data to send to eventkit.
func CollectEvent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pc, _, _, _ := runtime.Caller(0)
		_, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(context.Background(), runtime.FuncForPC(pc).Name())
		defer span.End()
		agents, err := useragent.ParseEntries([]byte(r.UserAgent()))
		product := "unknown"
		if err == nil && len(agents) > 0 && agents[0].Product != "" {
			product = agents[0].Product
			if len(product) > 32 {
				product = product[:32]
			}
		}

		var macHead string
		credentials := GetAccess(r.Context())
		if credentials != nil && credentials.AccessGrant != "" {
			access, err := grant.ParseAccess(credentials.AccessGrant)
			if err == nil {
				macHead = hex.EncodeToString(access.APIKey.Head())
			}
		}

		span.AddEvent("gmt",
			trace.WithAttributes(attribute.String("user-agent", product)),
			trace.WithAttributes(attribute.String("macaroon-head", macHead)),
			trace.WithAttributes(attribute.String("remote-ip", trustedip.GetClientIP(trustedip.NewListTrustAll(), r))))

		next.ServeHTTP(w, r)
	})
}
