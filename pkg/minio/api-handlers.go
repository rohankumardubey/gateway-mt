// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package minio

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"net/http"
	"os"
	"runtime"
	"strings"

	"storj.io/minio/cmd"
)

// objectAPIHandlers is linked to Minio's cmd.objectAPIHandlers and should not be changed.
type objectAPIHandlers struct {
	ObjectAPI func() cmd.ObjectLayer
	CacheAPI  func() cmd.CacheObjectLayer
}

// objectAPIHandlersWrapper should be used to extend objectAPIHandlers.
type objectAPIHandlersWrapper struct {
	core               objectAPIHandlers
	corsAllowedOrigins []string
}

func (h objectAPIHandlersWrapper) HeadObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	HeadObjectHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) CopyObjectPartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	CopyObjectPartHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutObjectPartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutObjectPartHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListObjectPartsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListObjectPartsHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) CompleteMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	CompleteMultipartUploadHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) NewMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	NewMultipartUploadHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) AbortMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	AbortMultipartUploadHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetObjectACLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetObjectACLHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutObjectACLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutObjectACLHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetObjectTaggingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetObjectTaggingHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutObjectTaggingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutObjectTaggingHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteObjectTaggingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteObjectTaggingHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) SelectObjectContentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	SelectObjectContentHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetObjectRetentionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetObjectRetentionHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetObjectLegalHoldHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetObjectLegalHoldHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetObjectHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) CopyObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	CopyObjectHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutObjectRetentionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutObjectRetentionHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutObjectLegalHoldHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutObjectLegalHoldHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutObjectHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteObjectHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketLocationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketLocationHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketPolicyHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketLifecycleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketLifecycleHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketEncryptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketEncryptionHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketObjectLockConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketObjectLockConfigHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketReplicationConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketReplicationConfigHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketVersioningHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketVersioningHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketNotificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketNotificationHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListenNotificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListenNotificationHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketACLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketACLHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketACLHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketACLHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketCorsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	var sb strings.Builder
	sb.WriteString("<CORSConfiguration><CORSRule>")
	for _, o := range h.corsAllowedOrigins {
		fmt.Fprintf(&sb, "<AllowedOrigin>%s</AllowedOrigin>", o)
	}
	// CorsHandler's AllowedHeader list is duplicated here
	allowedMethods := []string{http.MethodGet, http.MethodPut, http.MethodHead, http.MethodPost,
		http.MethodDelete, http.MethodOptions, http.MethodPatch}
	for _, o := range allowedMethods {
		fmt.Fprintf(&sb, "<AllowedMethod>%s</AllowedMethod>", o)
	}
	// CorsHandler's AllowedHeader list is not implemented here, because it includes "*"
	sb.WriteString("<AllowedHeader>*</AllowedHeader><ExposeHeader>*</ExposeHeader></CORSRule></CORSConfiguration>")
	writeSuccessResponseXML(w, []byte(sb.String()))
}

func (h objectAPIHandlersWrapper) PutBucketCorsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	writeErrorResponse(r.Context(), w, GetAPIError(cmd.ErrNotImplemented), r.URL, false)
}

func (h objectAPIHandlersWrapper) DeleteBucketCorsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	writeErrorResponse(r.Context(), w, GetAPIError(cmd.ErrNotImplemented), r.URL, false)
}

func (h objectAPIHandlersWrapper) GetBucketWebsiteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketWebsiteHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketAccelerateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketAccelerateHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketRequestPaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketRequestPaymentHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketLoggingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketLoggingHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) GetBucketTaggingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	GetBucketTaggingHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteBucketWebsiteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteBucketWebsiteHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteBucketTaggingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteBucketTaggingHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListMultipartUploadsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListMultipartUploadsHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListObjectsV2MHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListObjectsV2MHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListObjectsV2Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListObjectsV2Handler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListObjectVersionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListObjectVersionsHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListObjectsV1Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListObjectsV1Handler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketLifecycleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketLifecycleHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketReplicationConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketReplicationConfigHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketEncryptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketEncryptionHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketPolicyHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketObjectLockConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketObjectLockConfigHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketTaggingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketTaggingHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketVersioningHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketVersioningHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketNotificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketNotificationHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PutBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PutBucketHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) HeadBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	HeadBucketHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PostPolicyBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PostPolicyBucketHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteMultipleObjectsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteMultipleObjectsHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteBucketPolicyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteBucketPolicyHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteBucketReplicationConfigHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteBucketReplicationConfigHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteBucketLifecycleHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteBucketLifecycleHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteBucketEncryptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteBucketEncryptionHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) DeleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	DeleteBucketHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) PostRestoreObjectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	PostRestoreObjectHandler(h.core, w, r)
}

func (h objectAPIHandlersWrapper) ListBucketsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name())
	defer span.End()
	ListBucketsHandler(h.core, w, r)
}
