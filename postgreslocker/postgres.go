package postgreslocker

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/cockroach-go/crdb"
	"github.com/gilbsgilbs/glloq"
	"github.com/lib/pq"
)

// Locker supports PostgreSQL (CockroachDB, …).
type Locker struct {
	glloq.SQLLocker
}

func (l *Locker) SupportsDSN(dsn string) bool {
	return l.DBUrlDriver(dsn) == "postgres"
}

func (l *Locker) wrapError(err error) error {
	pqErr, isPqErr := err.(*pq.Error)
	if isPqErr && pqErr.Code.Name() == "query_canceled" {
		return context.DeadlineExceeded
	}

	return err
}

func (l *Locker) WithLock(
	ctx context.Context,
	opts *glloq.Options,
	fn func() error,
) error {
	db := l.DB

	tableName := opts.Params["table_name"]
	if tableName == "" {
		tableName = "glloq"
	}
	quotedTableName := pq.QuoteIdentifier(tableName)

	// can't use an advisory lock because CockroachDB doesn't support them.
	if _, err := db.ExecContext(
		ctx,
		"CREATE TABLE IF NOT EXISTS "+quotedTableName+"("+
			"	key text,"+
			"	CONSTRAINT glloq_pk PRIMARY KEY(key)"+
			");",
	); err != nil {
		return l.wrapError(err)
	}

	return crdb.ExecuteTx(context.Background(), db, &sql.TxOptions{}, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			"INSERT INTO "+quotedTableName+"(key) "+
				"VALUES ($1) "+
				"ON CONFLICT ON CONSTRAINT glloq_pk DO NOTHING;",
			opts.Key,
		); err != nil {
			return l.wrapError(err)
		}

		if _, err := tx.ExecContext(
			ctx,
			"SELECT * FROM "+quotedTableName+" "+
				"WHERE key = $1 "+
				"FOR UPDATE;",
			opts.Key,
		); err != nil {
			return l.wrapError(err)
		}

		return fn()
	})
}
