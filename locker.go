package glloq

import (
	"context"
	"database/sql"

	"github.com/xo/dburl"
)

type Locker interface {
	// SupportsDSN returns true if the DSN is supported by the Locker.
	SupportsDSN(dsn string) bool

	// Open allows the locker to open a connection to the backend.
	Open(ctx context.Context, dsn string) error

	// Close allows the locker to close the connection to the backend
	Close() error

	// Holds the lock. Returns ErrTimeout if context is done.
	WithLock(ctx context.Context, opts *Options, fn func() error) error
}

// SQLLocker implements some base locker methods for SQL-based backends.
type SQLLocker struct {
	DB *sql.DB
}

func (l *SQLLocker) DBUrlDriver(dsn string) string {
	u, err := dburl.Parse(dsn)
	if err != nil {
		return ""
	}
	return u.Driver
}

func (l *SQLLocker) Open(ctx context.Context, dsn string) error {
	db, err := dburl.Open(dsn)
	if err != nil {
		return err
	}

	if err := db.PingContext(ctx); err != nil {
		defer db.Close()
		return err
	}

	l.DB = db
	return nil
}

func (l *SQLLocker) Close() error {
	if l.DB != nil {
		return l.DB.Close()
	}

	return nil
}
