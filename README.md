<div align="center">
  <img src="https://github.com/eugenetriguba/bolt/raw/main/docs/logo.png" width="150" />
  <h1>Bolt</h1>
</div>

Bolt is a command-line tool that helps you manage your database migrations.
It allows you to write your database migrations in plain SQL, and Bolt will
handle ensuring they are applied and executed in the correct order.

:warning: This project is in early development.

## Installation

You may install Bolt using the following command:

```bash
$ go install github.com/eugenetriguba/bolt@latest
```

## Database Compatibility

At the moment, only PostgreSQL is supported. You're welcome to contribute support for other databases.
