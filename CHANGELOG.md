# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.1] 2018-08-02

 ### Added
 - layouts for blog index and single posts
 - page headings to template
 

## [0.3.0] 2018-06-28

### Added

- Connection managers can be loaded with factories
- Listeners, handlers and managers can all now run from external plugins
- External plugin versioning now enforced
- Multi-stage container builds used
- Plugin test is now part of our CI pipeline
- Ability to notify connection managers of graceful shutdowns
- Added helper for creating changelog entries

### Changed

- Internal listeners and handlers use the same plugin architecture as
external plugins
- Made Docker images have Secretless in the path for easier launching
- Fixed CI test suite
- Optimized many aspects of container builds
- Pinned Golang version to 1.10.3
- Standardized plugin API

## [0.2.0] - 2018-05-17

### Changed
- Added initial support for plugins
- Update CI to push images to Docker registry

## [0.1.0] - 2018-05-15

The first tagged version.

[Unreleased]: https://github.com/conjurinc/secretless/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/conjurinc/secretless/compare/v0.1.0...v0.2.0
[0.3.0]: https://github.com/conjurinc/secretless/compare/0.2.0...0.3.0
