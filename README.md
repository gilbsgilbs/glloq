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
# Supported DSNs include PostgreSQL (CockroachDB, ...), MySQL (Maria, ...), Files.
export GLLOQ_DSN=postgres://user:password@postgres:5432/mydb?sslmode=disable

# This wont run concurrently
glloq run_migrations.sh

# Override default timeout of 60 seconds to 10 minutes.
export GLLOQ_TIMEOUT=600

# You can specify a lock ID (if the back-end supports it)
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
