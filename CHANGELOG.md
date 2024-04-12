# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Execution times have been added to console output on migration apply or revert.

## [0.8.2] - 2024-03-23

### Added

- No code changes. New release is being created to be able to trigger a Github Action to release pre-compiled binaries with the release.

## [0.8.1] - 2024-03-23

### Fixed

- If a local migration with the same version already exists, bolt will now error out when listing out the migrations and indicate that there is a conflicting migration. Previously, it would ignore the first migration and use the second one with the same version.

## [0.8.0] - 2024-03-16

### Changed

- Migrations are now a single file with an upgrade and downgrade script delineated by a `-- migrate:up` or `--migrate:down` comment instead of the prior approach where the migration was a directory with a `upgrade.sql` and `downgrade.sql` file within it.

- Default version style is now timestamp.

- `bolt_migrations` `version` column is now a `VARCHAR(255)` instead of `CHARACTER(14)`.

## [0.7.0] - 2024-03-14

### Changed

- Renamed the following environment variables names:
  - `BOLT_DB_CONN_HOST` -> `BOLT_DB_HOST`
  - `BOLT_DB_CONN_PORT` -> `BOLT_DB_PORT`
  - `BOLT_DB_CONN_USER` -> `BOLT_DB_USER`
  - `BOLT_DB_CONN_PASSWORD` -> `BOLT_DB_PASSWORD`
  - `BOLT_DB_CONN_DBNAME` -> `BOLT_DB_NAME`
  - `BOLT_DB_CONN_DRIVER` -> `BOLT_DB_DRIVER`

- Renamed the `[connection]` section of `bolt.toml` to `[database]`.

### Fixed

- Panic that occurs when `bolt new` is run without a remote database connection setup.

## [0.6.0] - 2024-03-11

### Added

- Support for Microsoft SQL Server.
- Support for SQLite3.

### Changed

- Default the version style to sequential.

## [0.5.0] - 2024-03-10

### Added

- Support for MySQL.

### Changed

- The "postgres" driver name is now called "postgresql".

## [0.4.0] - 2024-03-09

### Changed

- Replaced the lib/pq postgres driver with pgx.

## [0.3.0] - 2024-03-02

### Added

- Documentation on `bolt version` command.
- Ability to use sequentially versioned or timestamp-based migration versions.
- More detailed error messages to help understand when things go wrong.

### Changed

- Moved `cli.Run` from `pkg` to `internal`

### Fixed

- A handful of unhandled errors which were leading to some silent failures.
- The migration version wasn't being parsed out and converted to its native type during sort. That is now being done, for either version style (sequential or timestamp).

## [0.2.1] - 2024-01-14

### Fixed

- Fixed a bug where the `bolt.toml` file was not being read for the documented format.

## [0.2.0] - 2024-01-13

### Added

- Support for environment variables to configure bolt. You may now use the TOML configuration file and/or environment variables to specify the configuration options. The environment variables will take precedence over any values set in the TOML configuration file.

  New environment variables
  - `BOLT_MIGRATIONS_DIR_PATH`
  - `BOLT_DB_CONN_HOST`
  - `BOLT_DB_CONN_PORT`
  - `BOLT_DB_CONN_USER`
  - `BOLT_DB_CONN_PASSWORD`
  - `BOLT_DB_CONN_DBNAME`
  - `BOLT_DB_CONN_DRIVER`

- Support for Windows when looking up the user's bolt configuration file.
  Previously, as the filesystem was traversed, it would look for "/" as the
  root directory.

- `bolt down` command to enable rollback of all migrations or optionally to a specific version.

- `bolt up` now supports a `-v` or `-version` flag which allows upgrading to a specific version.

- `bolt version` command to display the current version of bolt.
  
- Support for executing migration scripts without a transaction. Use `-- bolt: no-transaction` as the first line of your migration script to disable transactions for that migration.

### Changed

- `bolt_migrations` `version` column is now 14 characters long instead of 32.

- `bolt migrate` has been renamed to `bolt up`.
  
- `bolt list` has been renamed to `bolt status`.
  
- `bolt create` has been renamed to `bolt new`.

- The `up` command will now exit with a status code of 1
  if it encounters an error while running a migration. Previously,
  it was silently failing.

- `bolt status` output has been updated to be a tabular format.

- `[connection]` in `bolt.toml` should now be `[bolt.db.connection]`.

- `migrations_dir` in `bolt.toml` should now be in a `[bolt.migrations]` section and should now be called `directory_path`.
  
## [0.1.0] - 2023-12-30

### Added

- Initial release

<!-- template
## [X.Y.Z] - YYYY-MM-DD
### Added
- for new features.
### Changed
- for changes in existing functionality.
### Deprecated
- for soon-to-be removed features.
### Removed
- for now removed features.
### Fixed
- for any bug fixes.
### Security
- in case of vulnerabilities.
-->
