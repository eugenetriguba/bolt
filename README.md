# Bolt

Bolt is a command-line tool that helps you manage your database migrations. It allows you to write
your database migrations in plain SQL, and Bolt will handle ensuring they are applied and executed
in the correct order.

## Installation

You may download the binary from the Releases page or install it from source using Go:

```bash
$ git clone github.com/eugenetriguba/bolt
$ cd bolt
$ GOBIN=/usr/local/bin/ go install ./cmd/bolt/bolt.go
```

This will install a `bolt` binary under `/usr/local/bin/` that you may then use by running `bolt`
from a command-line.

## Database Compatibility

At the moment, only PostgreSQL is supported. You're welcome to contribute support for other databases.
