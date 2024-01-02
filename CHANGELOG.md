# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed

- The `migrate` command will now exit with a status code of 1
  if it encounters an error while running a migration. Previously,
  it was silently failing.

- `bolt_migrations` `version` column is now 14 characters long instead of 32.

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
