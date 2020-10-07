package anylocker

import (
	"context"

	"github.com/gilbsgilbs/glloq"
	"github.com/gilbsgilbs/glloq/filelocker"
	"github.com/gilbsgilbs/glloq/mysqllocker"
	"github.com/gilbsgilbs/glloq/postgreslocker"
)

// Locker supports locking across multiple backends.
type Locker struct {
	locker glloq.Locker
}

func (l *Locker) newLockerForDSN(dsn string) glloq.Locker {
	lockers := []glloq.Locker{
		&postgreslocker.Locker{},
		&mysqllocker.Locker{},
		&filelocker.Locker{},
	}

	for _, locker := range lockers {
		if locker.SupportsDSN(dsn) {
			return locker
		}
	}

	return nil
}

func (l *Locker) SupportsDSN(dsn string) bool {
	return l.newLockerForDSN(dsn) != nil
}

func (l *Locker) Open(ctx context.Context, dsn string) error {
	l.locker = l.newLockerForDSN(dsn)
	if l.locker == nil {
		return glloq.ErrUnsupportedDSN
	}

	return l.locker.Open(ctx, dsn)
}

func (l *Locker) Close() error {
	return l.locker.Close()
}

func (l *Locker) WithLock(
	ctx context.Context,
	opts *glloq.Options,
	fn func() error,
) error {
	return l.locker.WithLock(ctx, opts, fn)
}
