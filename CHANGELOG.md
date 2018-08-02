# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- layouts for blog index and single posts
- page headings to template
 
## [0.4.0] 2018-08-02

### Fixed

- Update style checker to work with auto-generated plugin docs

### Added

- Created plugin interface for providers
- A demo of using Secretless in Kubernetes exists in `demos/k8s-demo`
- The project uses the ASL 2.0 License
- The project has a website with initial styling
- The project has a logo
- A tutorial exists on the website of using Secretless in Kubernetes
- The website has documentation and quick start
- There is a basic auth http handler
- Golint runs as part of the Jenkins pipeline
- Project has a contributing and style guide

### Changed

- Bumped the Golang version from 1.10.3 to Go1.11beta
- Converted from using dep to using go modules
- Updated test suite to split out unit and integration tests
- Updated README to be in sync with website documentation
- Improved Vault provider, SSH, and SSH Agent test suites
- Secretless runs as a limited user in the Docker image
- Secretless defaults to /sock for socket files
- Old demos were removed
- Improvements to SSH handler / listener for better error handling / debugging
- Style updates were made to code based on golint output
- The plugin package was renamed from `plugin_v1` to `plugin/v1`
- Added support for soft-reloading of listeners

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
[0.4.0]: https://github.com/conjurinc/secretless/compare/0.3.0...0.4.0

