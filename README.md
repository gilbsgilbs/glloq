[![License Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Godoc](https://godoc.org/github.com/gilbsgilbs/glloq?status.svg)](https://pkg.go.dev/github.com/gilbsgilbs/glloq)
[![Actions Status](https://github.com/gilbsgilbs/glloq/workflows/CI/badge.svg)](https://github.com/gilbsgilbs/glloq/actions)
[![Coverage Status](https://coveralls.io/repos/github/gilbsgilbs/glloq/badge.svg?branch=master)](https://coveralls.io/github/gilbsgilbs/glloq?branch=master)

# Glloq

Glloq is a simple command line utility and Go library that lets you take an advisory lock on
various backends before running some action. This is especially useful when you want to avoid
running database migrations (for example) concurrently.

Officially supported backends currently include:

- PostgreSQL and derivatives (CockroachDB, …)
- MySQL and derivatives (MariaDB, …)
- Local files

but you can very easily implement your own.

## Usage

### As a CLI

```bash
# By default, glloq will use a ".glloq" lock file in the current working directory.
glloq sleep 10 &
glloq echo ok  # This displays "ok" in 10 seconds

# Supported backends include PostgreSQL (CockroachDB, ...), MySQL (Maria, ...) and local files.
export GLLOQ_DSN="postgres://user:password@postgres:5432/mydb?sslmode=disable"

# This wont run concurrently
glloq run_db_migrations.sh

# You can override default timeout of 1 minute to 10 minutes.
export GLLOQ_TIMEOUT=600

# You can specify a lock key (if supported by the backend).
GLLOQ_KEY=concurrent0 glloq run_migrations_0.sh
GLLOQ_KEY=concurrent1 glloq run_migrations_1.sh
```

### As a library

For detailed usage and examples, refer to [the godoc page](
https://pkg.go.dev/github.com/gilbsgilbs/glloq).

```go
import "github.com/gilbsgilbs/glloq"

func lockWithDSN(dsn string) error {
    return glloq.UseLocker(
        &postgreslocker.Locker{}
        &glloq.Options{
            DSN: dsn,
        },
        func() error {
            // Run DB migrations or anything.
        },
    })
}

import "github.com/gilbsgilbs/glloq/postgreslocker"

func lockWithDB(db *sql.DB) error {
    locker := postgreslocker.Locker{}
    locker.DB = db

    return locker.WithLock(
        context.Background(),
        &glloq.Options{},
        func() error {
            // ...
        },
    )
}
```
