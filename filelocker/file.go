package filelocker

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gilbsgilbs/glloq"
	"github.com/theckman/go-flock"
)

// Locker supports locking a local file.
type Locker struct {
}

func (l *Locker) SupportsDSN(dsn string) bool {
	return strings.HasPrefix(dsn, "file://") || strings.HasPrefix(dsn, "/")
}

func (l *Locker) Open(ctx context.Context, dsn string) error {
	return nil
}

func (l *Locker) Close() error {
	return nil
}

func (l *Locker) WithLock(
	ctx context.Context,
	opts *glloq.Options,
	fn func() error,
) error {
	dsn := strings.TrimPrefix(opts.DSN, "file://")

	fl := flock.New(dsn)
	defer func() {
		if err := fl.Unlock(); err != nil {
			log.Println("Warning, couldn't unlock file:", err)
		}
	}()

	pollDelayMilliseconds, err := strconv.Atoi(opts.Params["poll_delay"])
	if err != nil {
		pollDelayMilliseconds = 1
	}

	if _, err := fl.TryLockContext(
		ctx,
		time.Duration(pollDelayMilliseconds)*time.Millisecond,
	); err != nil {
		return err
	}

	return fn()
}
