package glloq_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gilbsgilbs/glloq"
	"github.com/gilbsgilbs/glloq/anylocker"
	"github.com/gilbsgilbs/glloq/filelocker"
	"github.com/gilbsgilbs/glloq/mysqllocker"
	"github.com/gilbsgilbs/glloq/postgreslocker"
)

func TestGlloq(t *testing.T) {
	postgresDSN := os.Getenv("POSTGRES_DSN")
	if postgresDSN == "" {
		postgresDSN = "postgresql://root:root@localhost:5432/glloq?sslmode=disable"
	}

	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		mysqlDSN = "mysql://root:root@localhost:3306/glloq"
	}

	testCases := []struct {
		name   string
		dsn    string
		params map[string]string
	}{
		{
			name: "File",
			dsn:  "file://.lock",
		},
		{
			name: "PostgreSQL",
			dsn:  postgresDSN,
		},
		{
			name:   "MySQL",
			dsn:    mysqlDSN,
			params: map[string]string{"poll_delay": "1"},
		},
	}

	for _, testCase := range testCases {
		dsn := testCase.dsn
		t.Run(testCase.name, func(t *testing.T) {
			t.Run("test lock function is called", func(t *testing.T) {
				called := false
				err := glloq.UseLocker(
					&anylocker.Locker{},
					&glloq.Options{
						DSN:    dsn,
						Key:    t.Name(),
						Params: testCase.params,
					},
					func() error {
						called = true
						return nil
					},
				)

				assert.Nil(t, err)
				assert.True(t, called)
			})

			t.Run("test lock function forwards errors", func(t *testing.T) {
				myError := errors.New("my custom error")
				err := glloq.UseLocker(
					&anylocker.Locker{},
					&glloq.Options{
						DSN:    dsn,
						Key:    t.Name(),
						Params: testCase.params,
					},
					func() error {
						return myError
					},
				)

				assert.Equal(t, myError, err)
			})

			t.Run("test concurrent locks", func(t *testing.T) {
				oneStarted := false
				oneDone := false
				oneCh := make(chan bool, 1)
				go func() {
					if err := glloq.UseLocker(
						&anylocker.Locker{},
						&glloq.Options{
							DSN:    dsn,
							Key:    t.Name(),
							Params: testCase.params,
						},
						func() error {
							oneStarted = true
							<-oneCh
							oneDone = true
							return nil
						},
					); err != nil {
						panic(err)
					}
				}()

				time.Sleep(50 * time.Millisecond)

				twoStarted := false
				twoDone := false
				twoCh := make(chan bool, 1)
				go func() {
					if err := glloq.UseLocker(
						&anylocker.Locker{},
						&glloq.Options{
							DSN:    dsn,
							Key:    t.Name(),
							Params: testCase.params,
						},
						func() error {
							twoStarted = true
							<-twoCh
							twoDone = true
							return nil
						},
					); err != nil {
						panic(err)
					}
				}()

				time.Sleep(50 * time.Millisecond)

				assert.True(t, oneStarted)
				assert.False(t, oneDone)
				assert.False(t, twoStarted)
				assert.False(t, twoDone)

				oneCh <- true
				time.Sleep(50 * time.Millisecond)

				assert.True(t, oneDone)
				assert.False(t, twoDone)

				twoCh <- true
				time.Sleep(50 * time.Millisecond)

				assert.True(t, twoDone)
			})

			t.Run("test timeout", func(t *testing.T) {
				ch := make(chan bool, 1)
				go func() {
					if err := glloq.UseLocker(
						&anylocker.Locker{},
						&glloq.Options{
							DSN:    dsn,
							Key:    t.Name(),
							Params: testCase.params,
						},
						func() error {
							<-ch
							return nil
						},
					); err != nil {
						panic(err)
					}
				}()

				time.Sleep(50 * time.Millisecond)

				err := glloq.UseLocker(
					&anylocker.Locker{},
					&glloq.Options{
						DSN:     dsn,
						Key:     t.Name(),
						Timeout: 50 * time.Millisecond,
						Params:  testCase.params,
					},
					func() error {
						return nil
					},
				)
				assert.Equal(t, glloq.ErrTimeout, err)

				close(ch)
			})
		})
	}

	t.Run("test connection timeout", func(t *testing.T) {
		called := false
		err := glloq.UseLocker(
			&anylocker.Locker{},
			&glloq.Options{
				DSN:     "postgresql://127.254.254.254:1333",
				Timeout: 100 * time.Millisecond,
			},
			func() error {
				called = true
				return nil
			},
		)

		assert.False(t, called)
		assert.Equal(t, err, glloq.ErrTimeout)
	})

	t.Run("unsupported DSN", func(t *testing.T) {
		err := glloq.UseLocker(
			&anylocker.Locker{},
			&glloq.Options{
				DSN: "foobar://barbaz",
			},
			func() error {
				return nil
			},
		)
		assert.Equal(t, glloq.ErrUnsupportedDSN, err)
	})
}

// The simplest way to use glloq is by using a data source name (DSN). UseLocker will take
// care of opening and closing any connection or socket, will wait for the backend to be
// available and will hold the lock and release it when you're finished.
func Example() {
	err := glloq.UseLocker(
		&anylocker.Locker{},
		&glloq.Options{
			// This is a connection string to your backend. For SQL-based backends,
			// dburl (https://github.com/xo/dburl) is used.
			DSN: "postgres://user:password@localhost:5432/db?sslmode=disable",
			// DSN: "mysql://user:password@localhost:3006/db",
			// DSN: "file://.lock",

			// Maximum time to wait for the backend and the lock. Defaults to 1 minute.
			Timeout: 1 * time.Hour,

			// An optional lock key, if supported by the backend.
			Key: "someUniqueKey",

			// backend-specific parameters.
			Params: map[string]string{},
		},
		func() error {
			// You can run any synchronized operation here, such as database migrations.
			// It won't run concurrently.
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}

// The file locker allows you take a lock using a local file.
func Example_fileLocker() {
	locker := filelocker.Locker{}
	err := locker.WithLock(
		context.Background(),
		&glloq.Options{
			DSN: "file:///tmp/myAppLockFile.lock",
		},
		func() error {
			// ...
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}

func Example_postgresLocker() {
	var db *sql.DB

	locker := postgreslocker.Locker{}
	locker.DB = db

	err := locker.WithLock(
		context.Background(),
		&glloq.Options{
			Params: map[string]string{
				// Name of the table that will be created to take locks.
				// Defaults to glloq.
				"table_name": "my_lock_table",
			},
		},
		func() error {
			// ...
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}

func Example_mysqlLocker() {
	var db *sql.DB

	locker := mysqllocker.Locker{}
	locker.DB = db

	err := locker.WithLock(
		context.Background(),
		&glloq.Options{
			Params: map[string]string{
				// Name of the table that will be created to take locks.
				// Defaults to glloq.
				"table_name": "my_lock_table",
			},
		},
		func() error {
			// ...
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}
