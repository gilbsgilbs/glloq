package mysqllocker

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gilbsgilbs/glloq"
	_ "github.com/go-sql-driver/mysql"
)

// Locker supports MySQL / MariaDB.
type Locker struct {
	glloq.SQLLocker
}

func (l *Locker) SupportsDSN(dsn string) bool {
	return l.DBUrlDriver(dsn) == "mysql"
}

func (l *Locker) quoteIdentifier(identifier string) string {
	// no driver built-in for this?
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func (l *Locker) withTx(db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = fn(tx)
	commitErr := tx.Commit()
	if err != nil {
		return err
	}
	return commitErr
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
	quotedTableName := l.quoteIdentifier(tableName)

	// Can't use an advisory lock because they can't be tight to a transaction.
	// So it could leak locks.
	if _, err := db.ExecContext(
		ctx,
		"CREATE TABLE IF NOT EXISTS "+quotedTableName+" ("+
			"	id VARCHAR(100) NOT NULL PRIMARY KEY"+
			");",
	); err != nil {
		return err
	}

	if _, err := db.ExecContext(
		ctx,
		"INSERT IGNORE INTO "+quotedTableName+"(id) VALUES (?);",
		opts.Key,
	); err != nil {
		return err
	}

	return l.withTx(db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			"SELECT * FROM "+quotedTableName+" "+
				"WHERE id = ? "+
				"FOR UPDATE;",
			opts.Key,
		); err != nil {
			return err
		}

		return fn()
	})
}
