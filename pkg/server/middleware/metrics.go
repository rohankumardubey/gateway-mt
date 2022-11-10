// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package middleware

import (
	"net/http"
)

var (
	_ http.ResponseWriter = (*flusherDelegator)(nil)
	_ http.Flusher        = (*flusherDelegator)(nil)
)

// measureFunc is a common type for functions called at particular points of the
// request that we wish to measure, such as when a header is written. It receives
// the HTTP status code that was written as an argument.
type measureFunc func(int)

// flusherDelegator acts as a gatherer of status code and bytes written.
//
// It calls atWriteHeaderFunc only once for WriteHeader (so that
// atWriteHeaderFunc executes expectedly), but it still delegates WriteHeader
// from the caller. It's "illegal" to call WriteHeader twice, but we don't want
// to mask any bugs.
//
// flusherDelegator is loosely inspired by the design of
// prometheus/client_golang/prometheus/promhttp package.
type flusherDelegator struct {
	http.ResponseWriter

	// atWriteHeaderFunc is called at the call to WriteHeader.
	atWriteHeaderFunc measureFunc

	// atTimeToFirstByteFunc is called when bytes are first written.
	atTimeToFirstByteFunc measureFunc

	status                  int
	written                 int64
	wroteHeader             bool
	observedTimeToFirstByte bool
}

func (f *flusherDelegator) WriteHeader(code int) {
	if f.atWriteHeaderFunc != nil && !f.wroteHeader {
		f.atWriteHeaderFunc(code)
	}
	f.status = code
	f.wroteHeader = true
	f.ResponseWriter.WriteHeader(code)
}

func (f *flusherDelegator) Write(b []byte) (int, error) {
	if !f.wroteHeader {
		f.WriteHeader(http.StatusOK)
	}
	n, err := f.ResponseWriter.Write(b)
	if f.atTimeToFirstByteFunc != nil && !f.observedTimeToFirstByte {
		f.atTimeToFirstByteFunc(f.status)
		f.observedTimeToFirstByte = true
	}
	f.written += int64(n)
	return n, err
}

func (f flusherDelegator) Flush() {
	if !f.wroteHeader {
		f.WriteHeader(http.StatusOK)
	}
	f.ResponseWriter.(http.Flusher).Flush()
}
