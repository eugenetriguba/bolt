<div align="center">
  <img src="https://github.com/eugenetriguba/bolt/raw/main/docs/logo.png" width="150" />
  <h1>Bolt</h1>
</div>

<p>
    <a href="https://godoc.org/github.com/eugenetriguba/bolt">
        <img src="https://godoc.org/github.com/eugenetriguba/bolt?status.svg" alt="GoDoc">
    </a>
    <a href="https://goreportcard.com/report/github.com/eugenetriguba/bolt">
        <img src="https://goreportcard.com/badge/github.com/eugenetriguba/bolt" alt="Go Report Card Badge">
    </a>
    <img alt="Version Badge" src="https://img.shields.io/badge/version-0.5.0-blue" style="max-width:100%;">
</p>


Bolt is a command-line tool designed to simplify and streamline your database migration process.

🌟 Key Features:
- **SQL-Powered Flexibility**: 🧩 - Write your migration scripts in plain SQL.

- **Bidirectional Migrations**: 🔀 - With support for both 'up' (migrating) and 'down' (downgrading) migrations, Bolt provides complete control over your database versions.

- **Time-Travel with Ease**: ⏳ - Migrate or downgrade to any specific point in time. Bolt allows you to specify a specific migration version to migrate or downgrade to, allowing you to easily time-travel to any point in your database's history.

## Table of Contents

- [Installation](#installation)
- [Tutorial](#tutorial)
  - [Creating your first migration](#creating-your-first-migration)
  - [Writing your migration scripts](#writing-your-migration-scripts)
  - [Starting up a Database Container](#starting-up-a-database-container)
  - [Configuring Bolt](#configuring-bolt)
  - [Applying your migration](#applying-your-migration)
  - [Checking Migration Status](#checking-migration-status)
  - [Verifying the Migration](#verifying-the-migration)
  - [Downgrading Your Migration](#downgrading-your-migration)
  - [Verifying Migration Reversion](#verifying-migration-reversion)
  - [Next Steps](#next-steps)
- [How-to](#how-to)
  - [How to execute a migration script without a transaction](#how-to-execute-a-migration-script-without-a-transaction)
- [Reference](#reference)
  - [Database Compatibility](#database-compatibility)
  - [Configuration](#configuration)
    - [Configuration File](#configuration-file)
    - [Environment Variables](#environment-variables)
  - [Commands](#commands)
    - [`bolt new`](#bolt-new)
    - [`bolt up`](#bolt-up)
    - [`bolt down`](#bolt-down)
    - [`bolt status`](#bolt-status)
    - [`bolt version`](#bolt-version)
  - [Script Execution Options](#script-execution-options)
  - [Version Styles](#version-styles)
- [Explanation](#explanation)
  - [How Are Migration Scripts Executed?](#how-are-migration-scripts-executed)
  - [How Does Bolt Know What Migrations Have Been Applied?](#how-does-bolt-know-what-migrations-have-been-applied)
  - [How Are Migrations Applied?](#how-are-migrations-applied)
  - [How Are Migrations Reverted?](#how-are-migrations-reverted)
  - [What are Migration Version Styles?](#what-are-migration-version-styles)
  - [Why can't I change between version styles?](#why-cant-i-change-between-version-styles)
  - [How is the migration message used by Bolt?](#how-is-the-migration-message-used-by-bolt)

## Installation

To install Bolt, run the following command:

```bash
$ go install github.com/eugenetriguba/bolt/cmd/bolt@latest
```

This will install Bolt to your `$GOBIN` directory. Make sure that `$GOBIN` is in your `$PATH`.

## Tutorial

In this tutorial, we'll guide you through the basics of using Bolt by running through how you can create a migration, apply it, revert it, and see the status of your migrations.

### Creating your first migration

Run the following command to create your first migration:

```bash
$ bolt new -m "my first migration"
```

Upon completion, a migrations directory will be created with the following structure:

```bash
migrations/
└── 20240113162817_my_first_migration
    ├── downgrade.sql
    └── upgrade.sql

1 directories, 2 files
```

Bolt automatically creates the migrations directory and your migration scripts inside it.

### Writing your migration scripts

Bolt utilizes plain SQL for migration scripts. Each migration directory contains `upgrade.sql` for applying migrations and `downgrade.sql` for reverting them.

1. Write the Upgrade Script

Edit `upgrade.sql` with the following SQL command to create a new table:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

2. Write the Downgrade Script

Edit `downgrade.sql` with the following SQL command to remove the table:

```sql
DROP TABLE users;
```

### Starting up a Database Container

We'll use Docker to run a PostgreSQL database container for this tutorial:

1. Install Docker

If you don't have docker installed, view the installation instructions [here](https://docs.docker.com/get-docker/). 

2. Startup the PostgreSQL container

```bash
$ docker run --name bolt-postgres -e POSTGRES_USER=bolt_user -e POSTGRES_PASSWORD=bolt_password -e POSTGRES_DB=bolt_tutorial_db -p 5432:5432 --detach postgres:16.1-alpine3.19
```

### Configuring Bolt

Bolt can be configured with a `bolt.toml` file or environment variables. For simplicity, we'll use a `bolt.toml` file in this tutorial.

1. Create a `bolt.toml` file:

```bash
$ touch bolt.toml
```

2. Add the following contents to the file:

```toml
[connection]
host = "localhost"
port = 5432
user = "bolt_user"
password = "bolt_password"
dbname = "bolt_tutorial_db"
driver = "postgres"
```

### Applying your migration

Apply your migration, which will execute the `upgrade.sql` script:

```bash
$ bolt up
Applying migration 20240113162817_my_first_migration..
Successfully applied migration 20240113162817_my_first_migration!
```

### Checking Migration Status

To view the status of your migration:

```bash
$ bolt status
Version           Message               Applied
20240113162817    my_first_migration    X
```

This command displays the migration's version, name, and whether the migration was applied or not.

### Verifying the Migration

Connect to the database container to confirm the users table was created:

```bash
$ docker exec -it bolt-postgres psql -U bolt_user -d bolt_tutorial_db
```

```psql
bolt_tutorial_db=# \dt
              List of relations
 Schema |      Name       | Type  |   Owner
--------+-----------------+-------+-----------
 public | bolt_migrations | table | bolt_user
 public | users           | table | bolt_user
(2 rows)
```

We can see that our `users` table was created successfully. We also see another table that we didn't explicitly create called `bolt_migrations`. This table is used by Bolt to keep track of which migrations have been applied to your database. You should not modify this table manually.

### Downgrading Your Migration

To revert your migration:

```bash
$ bolt down
Reverting migration 20240113162817_my_first_migration..
Successfully reverted migration 20240113162817_my_first_migration!
```

### Verifying Migration Reversion

Reconnect to the database container and check that the users table has been removed:

```bash
$ docker exec -it bolt-postgres psql -U bolt_user -d bolt_tutorial_db
```

```psql
bolt_tutorial_db=# \dt
              List of relations
 Schema |      Name       | Type  |   Owner
--------+-----------------+-------+-----------
 public | bolt_migrations | table | bolt_user
(1 row)
```

### Next Steps

Congrats! You've learned the core features of Bolt. If you want learn more, see the [Reference](#reference) or [Explanation](#explanation) section.

## How-to

### How to execute a migration script without a transaction

In your `upgrade.sql` or `downgrade.sql`, add the following line as the first line of your script:

```sql
-- bolt: no-transaction
```

## Reference

### Database Compatibility

- PostgreSQL
- MySQL
- Microsoft SQL Server

### Configuration

There are two ways to configure Bolt: via a `bolt.toml` file or via environment variables. If you use both methods, the environment variables will always take precedence.

#### Configuration File

Bolt attempts to find a `bolt.toml` file in your current working directory
or in any parent directory.

```toml
[migrations]
# The directory where your migration scripts are located.
# Defaults to "migrations". This directory is relative to
# the current working directory. You may also use an absolute
# path.
directory_path = "migrations"
# The migration versioning style you prefer. Supported options
# are "timestamp" and "sequential". Defaults to "timestamp".
#
# Note: It is not supported to change migration version styles
# i.e. you can't have a mix of sequential and timestamp migrations.
version_style = "timestamp"

# Connection parameters for the database Bolt will be
# applying migrations to. All connection parameters are
# required.
[connection]
# The host to use to connect to your database.
host = 
# The port to use to connect to your database.
# Note: This should be an integer.
port = 
# The user to use to connect to your database.
user = 
# The password to use to connect to your database.
password = 
# The name of the database within your DBMS.
dbname = 
# The name of the database driver to use to connect to
# the database. Either "postgresql", "mysql", or "mssql".
driver = 
```

#### Environment Variables

All configuration file settings have corresponding environment variables.

- `BOLT_MIGRATIONS_DIR_PATH`
- `BOLT_MIGRATIONS_VERSION_STYLE`
- `BOLT_DB_CONN_HOST`
- `BOLT_DB_CONN_PORT`
- `BOLT_DB_CONN_USER`
- `BOLT_DB_CONN_PASSWORD`
- `BOLT_DB_CONN_DBNAME`
- `BOLT_DB_CONN_DRIVER`

### Commands

#### `bolt new`

```bash
$ bolt help new
new [-message|-m]:
	Create a new database migration
    -m string
    	alias for -message (default "autogenerated")
  -message string
    	Message to use for the migration (default "autogenerated")
```

#### `bolt up`

```bash
$ bolt help up
up [-version|-v]:
	Apply migrations against the database
    -v string
    	alias for -version
  -version string
    	The version to upgrade up and including to.
```

#### `bolt down`

```bash
$ bolt help up
down [-version|-v]:
	Downgrade migrations against the database
    -v string
    	alias for -version
  -version string
    	The version to downgrade down and including to.
```

#### `bolt status`

```bash
$ bolt help status
status:
	List the database migrations and their statuses
```

#### `bolt version`

```bash
$ bolt help version
version:
	Show the current version of Bolt
```

### Script Execution Options

Bolt provides a way for you to customize how your migration scripts are executed if you need something different than the default behavior.

You can do this by adding a comment to the top of your migration script. The comment must be in the following format: `-- bolt: <option>`. The following options are available:

- `no-transaction`: Execute the migration script without a transaction. By default, every migration script will be attempted to be executed within a transaction, however, some SQL commands cannot be executed within a transaction so you'll need to opt out of that behavior in those cases.

### Version Styles

The following version styles are supported:
- Timestamp
- Sequential

## Explanation

### How Are Migration Scripts Executed?

Bolt executes upgrade and downgrade migration scripts in a transaction. This ensures that if any errors occur during the execution of any migration script, the transaction will be rolled back and the migration will be marked as failed. Bolt will then exit with an error code and output what error has occurred to standard error.

### How Does Bolt Know What Migrations Have Been Applied?

Bolt keeps track of which migrations have been applied to your database by creating a table called `bolt_migrations`. This table contains a single column called `version` which is the version of the migration that was applied. That version is compared to the versions you have locally.

### How Are Migrations Applied?

When you run `bolt up`, Bolt will look at your local migration scripts and compares the version part of the migration directory names to the versions that have been inserted into the `bolt_migrations` table. Any versions that aren't in the table will be applied in order, starting with the oldest migration. Applying a migration entails executing the `upgrade.sql` script in a transaction and inserting in the migration's version into the `bolt_migrations` table.

### How Are Migrations Reverted?

When you run `bolt down`, Bolt will look at the `bolt_migrations` table and compare the versions to the versions of your local migration scripts. Any versions that are in the table but not in your local migration scripts will be reverted in order, starting with the newest migration. Reverting a migration entails executing the `downgrade.sql` script in a transaction and removing the migration's version from the `bolt_migrations` table.

### What are Migration Version Styles?

Whenever you create a migration, it'll be prefixed with a "version". This is what is used by Bolt to keep track of what order to apply or revert migrations. Version styles are different supported options for what this prefix will be.

Bolt supports two different options for this:
- Sequential versions
- Timestamp versions

Sequential versions are incrementing integers. When you create a migration, it'll be prefixed with a number like "001", "002", and so on. Timestamp versions use the current time on migration creation as the prefix.

### Why can't I change between version styles?

The main reason that switching back and forth between "sequential" and "timestamp" versions is not supported is because Bolt will no longer know how to properly sort your migrations in the right order (and therefore, apply or revert them in the right order). If you've been using sequential migrations, such as "001" and "002", and then move over to timestamp migrations, such as "20200101000000", Bolt won't know how to order these correctly when both exist. Furthermore, because the "version" is parsed and sorted according to the version style that is configured, the parsing will fail if a mix of version styles are used.

### How is the migration message used by Bolt?

All migrations are created on the local filesystem in the format `<version>_<message>`. Bolt uses the `<version>` part of the name for keeping track of the migrations and how they should be applied or reverted. The `<message>` part is purely informational so you know what that migration is for. Bolt does not use or store it anywhere. You may alter this message whenever you'd like as long as the format is kept consistent (`_` after the `<version>`). However, the version should not be directly manipulated.
