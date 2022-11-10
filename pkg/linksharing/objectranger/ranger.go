// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package objectranger

import (
	"context"
	"go.opentelemetry.io/otel"
	"io"
	"os"
	"runtime"

	"storj.io/common/ranger"
	"storj.io/uplink"
)

// ObjectRanger holds all the data needed to make object downloadable.
type ObjectRanger struct {
	p      *uplink.Project
	o      *uplink.Object
	bucket string
}

// New creates a new object ranger.
func New(p *uplink.Project, o *uplink.Object, bucket string) ranger.Ranger {
	return &ObjectRanger{
		p:      p,
		o:      o,
		bucket: bucket,
	}
}

// Size returns object size.
func (ranger *ObjectRanger) Size() int64 {
	return ranger.o.System.ContentLength
}

// Range returns object read/close interface.
func (ranger *ObjectRanger) Range(ctx context.Context, offset, length int64) (_ io.ReadCloser, err error) {
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	return ranger.p.DownloadObject(ctx, ranger.bucket, ranger.o.Key, &uplink.DownloadOptions{Offset: offset, Length: length})
}
