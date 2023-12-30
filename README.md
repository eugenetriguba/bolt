# Bolt

Bolt is a command-line tool written that allows you to write and manage plain SQL upgrade and downgrade migrations.

## Installation

You may download the binary from the Releases page or install it from source using Go:

```bash
$ git clone github.com/eugenetriguba/bolt
$ cd bolt
$ GOBIN=/usr/local/bin/ go install ./cmd/bolt/bolt.go
```

This will install a `bolt` binary under `/usr/local/bin/` that you may then use by running `bolt`
from a command-line.
