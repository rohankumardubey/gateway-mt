// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package badgerauth

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"storj.io/common/sync2"
)

// BackupError is a class of backup errors.
var BackupError = errs.Class("backup")

// Client is the interface for the object store.
type Client interface {
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64,
		opts minio.PutObjectOptions) (info minio.UploadInfo, err error)
}

// BackupConfig provides options for creating a backup.
type BackupConfig struct {
	Enabled         bool          `user:"true" help:"enable backups" default:"false"`
	Endpoint        string        `user:"true" help:"backup bucket endpoint hostname, e.g. s3.amazonaws.com"`
	Bucket          string        `user:"true" help:"bucket name where database backups are stored"`
	Prefix          string        `user:"true" help:"database backup object path prefix"`
	Interval        time.Duration `user:"true" help:"how often full backups are run" default:"1h"`
	AccessKeyID     string        `user:"true" help:"access key for backup bucket"`
	SecretAccessKey string        `user:"true" help:"secret key for backup bucket"`
}

// Backup represents a backup job that backs up the database.
type Backup struct {
	db        *DB
	Client    Client
	SyncCycle *sync2.Cycle
	prefix    string
}

// NewBackup returns a new Backup. Note that BadgerDB does not support opening
// multiple connections to the same database, so we must use the same DB
// connection as normal KV operations.
func NewBackup(db *DB, client Client) *Backup {
	syncCycle := sync2.NewCycle(db.config.Backup.Interval)
	syncCycle.SetDelayStart()
	return &Backup{
		db:        db,
		SyncCycle: syncCycle,
		Client:    client,
		prefix:    path.Join(db.config.Backup.Prefix, db.config.ID.String()),
	}
}

// RunOnce performs a full backup of the database
//
// Each backup is split into separate prefix parts. For example:
//
//	mybucket/myprefix/mynodeid/2022/04/13/2022-04-13T03:42:07Z
func (b *Backup) RunOnce(ctx context.Context) (err error) {
	pc, _, _, _ := runtime.Caller(0)
	ctx, span := otel.Tracer(os.Getenv("SERVICE_NAME")).Start(ctx, runtime.FuncForPC(pc).Name(), trace.WithAttributes(attribute.String("node_id", b.db.config.ID.String())))
	defer span.End()

	r, w := io.Pipe()
	t := time.Now().UTC()
	key := path.Join(b.prefix, t.Format("2006/01/02"), t.Format(time.RFC3339))

	var group errgroup.Group
	group.Go(func() error {
		stream := b.db.db.NewStream()
		stream.LogPrefix = "DB.Backup"
		stream.SinceTs = 0
		stream.NumGo = 1
		_, err := stream.Backup(w, 0)
		return w.CloseWithError(err)
	})

	_, err = b.Client.PutObject(ctx, b.db.config.Backup.Bucket, key, r, -1, minio.PutObjectOptions{})
	if err != nil {
		return BackupError.New("upload object: %w", err)
	}

	return BackupError.Wrap(group.Wait())
}
