# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2024-09-25

### Fixed

- Do not migrate `<WC>-cluster-values` config map or secret from vintage to CAPI MC since then cluster-app-operator cannot create those objects with the new values

## [0.2.0] - 2024-08-07

### Changed

- Update dependencies

### Fixed

- Fix label detection for app filtering
- Fix error handling for removing finalizer on namespace

## [0.1.0] - 2024-03-27

- Initial release

[Unreleased]: https://github.com/giantswarm/app-migration-cli/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/giantswarm/app-migration-cli/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/app-migration-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/app-migration-cli/releases/tag/v0.1.0
