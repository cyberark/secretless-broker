# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0] 2019-10-21

### Added

- Added a new public [plugin interface](pkg/secretless/plugin) for building connector plugins
- Added a new public [log interface](pkg/secretless/log) for standardizing logging
- Added code coverage reporting to unit test output
- Added ability to run k8s-demo test on GKE

### Changed

- Refactored existing connectors to use new public connector plugin interface
- Changed the core proxy and plugin manager to support the new public connector
  plugin interface
- Edited website Google Group links to link to Discourse
- Updated the [example plugin](test/plugin) to implement the new plugin interface
- Minor format changes to Apache 2.0 license
- Project structure reorganized
- Internal code updated to use v2 config instead of v1 config
- Goreleaser build updated to cross-compile linux and darwin
- Updated Conjur tests to use official CLI image

### Fixed

- Improve namespace cleanup in k8s-ci/test
- Add COMPOSE_PROJECT_NAME to tests to fix namespace collision errors
- Updated k8s-demo to use LoadBalancer on Services to avoid NodePort conflicts
- Clarified quick demo directions
- Improved error-handling / retry logic in k8s-ci

### Deprecated

- `Protocol` key in v2 config is replaced with `connector` key

## [1.1.0] 2019-08-09

### Added
- Added version output to logs on startup
- Added NOTICES.txt to the project
- Added dependency tracking tools and info
- Added ability to configure PG connector with `host`/`port` combination
- Added gitleaks config to enable running gitleaks pre-push

### Changed
- Deprecated support for PG connector configurations with `address` field
- Minor edits to website quick start instructions
- Updated versioning method for the project to use version.go
- Parallelized integration tests
- Upgraded summon module dependency to 0.7.0
- Cleaned up go.mod and go.sum with `go mod tidy`
- Only pin to vault/api submodule rather than larger vault module
- MySQL port defaults to 3306 if not specified
- Updated health check test to wait longer for server to come up to prevent
  test failures
- Revised README for simplicity and to describe available releases

### Removed
- Removed custom script to check style in favor of code climate
- Removed old benchmark proof of concepts
- Removed GitLab pipeline
- Removed ability to pass `dbname` in the `address` field of the PostgreSQL
  config - the PostgreSQL `address` config now only accepts `host:[port]`

### Fixed
- Resolved shellcheck errors
- Standardized spacing in `testutil` package
- Fixed changelog prefill script

## [1.0.0] 2019-07-03

### Added
- Added aggregation script to performance test code

### Changed
- Revised "service authenticator" to "service connector" and updated docs/links
- Moved plugin interfaces to internal pending redesign
- Updated project so internal dev tags push to internal registry instead of
  DockerHub
- Removed beta label from project and updated README
- Updated configuration samples in demos to use v2 config

### Fixed
- Fixed go lint errors
- Fixed broken homepage link
- Fixed bug with MySQL connector (#766) that returned "Malformed packet" for all
  errors

### Removed
- Removed deprecated full-demo

## [0.8.0] 2019-06-18

### Added
- Added a performance testing tool to bin/juxtaposer
- Added a v2 configuration syntax that is simpler and easier to use

### Fixed
- Updated the Conjur Kubernetes authenticator client to 0.13.0 to fix a bug
  that caused the token refresh to fail after the cert expired

### Changed
- Revised "k8s-demo"
- Upgraded to Golang v1.12.5 from v1.11.4
- Updated `conjur-authn-k8s-client` dependency to v0.13.0
- Updated `conjur-api-go` dependency to v0.5.2
- Removed third-party module for evaluating home directory path
- Updated goreleaser config to address deprecated `archive` tag
- Revised PR template to remove unneeded manual tests

## [0.7.1] 2019-05-16

### Added
- Added several issue templates
- Added improved tutorial flow to webpage

### Changed
- Noted alpha support for HCV provider in README
- Improved CRD testing
- Updated base image used for GitLab CI
- Updated contributor info for documentation
- Updated to use universal `psql` command throughout repo`

### Fixed
- Corrected tutorial issues with code snippets and spacing

## [0.7.0] 2019-03-26

### Added

- Add ability to verify plugin checksums
- Add kubernetes secrets provider to README.md
- Note styling in Kubernetes tutorial
- Add link to /tutorials in the top nav
- Add daily build trigger
- Add redirect link capabilities
- Add version to README.md
- Add a README for the shared library
- C shared library exposing secret providers (POC)
- Add custom 404 page

### Changed

- Update Kubernetes Tutorial for Simplicity and Clarity
- Simplify fast k8s tutorial
- Update CTA links
- Refactor mysql/NativePassword to take bytes
- Clean up Go memory of secrets
- Refactor MySQL handler for readability and consistency
- Updating website build to gen godocs in go img

### Fixed

- Fix kubernetes secrets example in README
- Fix kubernetes-secrets-provider hash
- Remove target=blank from footer links
- Fix broken website publishing
- Fix all non-TODO CodeClimate issues
- Fix ssh hadler test naming
- Make ssh-handler integration test pull images before build
- Remove references to doc layout and update links
- Remove hashicorp root cert to fix broken build
- Fix the vault test that broke due to vault CLI updates
- Re-enable ssh-handler tests

## [0.6.4] 2019-02-01

### Added
- Added a design proposal for credential zeroization
- Improved dev functionality in handler integration tests

### Changed
- Removed checksum hacks for client-go from Dockerfiles, since this is fixed
  in Go 1.11.4
- Improved and refactored database integration test suite

### Fixed
- Updated MySQL handler to handle authPluginName mismatch and to have consistent
  sequenceIds

## [0.6.3] 2019-01-11

### Added
- Database handlers support private-key pair as sslkey and sslcert

### Changed
- Permissions have been fixed for OpenShift non-root integration and use

## [0.6.2] 2019-01-09

### Added
- Added Kubernetes authenticator documentation for Conjur credential provider

### Changed
- Sanitized remaining listeners/handlers from dumping data on the CLI when debug mode is on
- Removed developer-only debug mode from demos and examples

## [0.6.1] 2019-01-08

### Changed
- Updated conjur-api-go dependency

### Added
- Added `/ready` and `/live` endpoints on port 5335 for checking if the broker is ready/live

## [0.6.0] 2018-12-20

### Added
- SSL support for MySQL and PostgreSQL handlers
- Improved test utilities
- Added flag for CPU or memory profiling

### Changed
- Updated demos to support databases configured with SSL
- Allow ./bin/test_integration to specify individual test_folders + local flag
- Updated goreleaser process to use new image

## [0.5.2] 2018-11-26

### Fixed
- Updated Kubernetes secrets provider to retrieve secrets from current namespace
- Fixed broken GitLab build referencing non-existent image
- Fixed broken keychain provider tests, and made easier to run manually

## [0.5.1] 2018-11-20

### Added
- Tests for Kubernetes Secrets provider
- Initial benchmark data is compiled during build
- Project now builds in GitLab
- Goreleaser support for deb/rpm packages
- Initial implementation of AWS Secrets provider

### Changed
- Removed bash4 dependency
- Documentation updates
- Updated Jekyll dependency to use version 3.8.4

### Removed
- Moved the sidecar injector functionality to its [own repo](https://github.com/cyberark/sidecar-injector)

## [0.5.0] 2018-09-06

### Fixed

- Fix for "no matching manifest for linux/amd64 in the manifest" error
- Linter fixes
- Fixed fast-restart http listener error
- Fixed soft-reload 100% CPU bug
- Cleaned up channel closing in main proxy loop
- Update pg test to use sslmode=disable
- Fix Proxy#Run SHUTDOWN event deadlock
- Secretless shutdown ensures handlers shutdown; inform clients of closed
  connections
- Fixed panic when using server plugin with "match" config field

### Added

- Added support for Conjur Kubernetes authenticator in Conjur provider
- Added Kubernetes secrets provider
- Added support for a K8s custom resource definition of Secretless Broker config
- Updated standard config file reading to be in the form of a config manager
  plugin
- Added ability to watch for configuration changes through CRDs
- Add test for clean listener shutdown
- Added sidecar injector admission-webhook-controller
- Add BaseHandler and BaseListener
- Added Goreleaser for automated binary archive building (for tags)
- Added http credential zeroization
- Publish quick start Docker image

### Changed

- Repo moved to `cyberark`, images pushed to DockerHub
- Updated K8s demo to use K8s secrets provider
- Upgraded to Go1.11
- Conjur handler updated to instantiate Conjur provider
- Updates to website style, homepage, copy to clipboard, and minor content edits
- Update demos to use Dockerhub image
- Name updated to Secretless Broker

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

[Unreleased]: https://github.com/cyberark/secretless-broker/compare/v1.2.0...HEAD
[0.2.0]: https://github.com/cyberark/secretless-broker/compare/v0.1.0...v0.2.0
[0.3.0]: https://github.com/cyberark/secretless-broker/compare/v0.2.0...v0.3.0
[0.4.0]: https://github.com/cyberark/secretless-broker/compare/v0.3.0...v0.4.0
[0.5.0]: https://github.com/cyberark/secretless-broker/compare/v0.4.0...v0.5.0
[0.5.1]: https://github.com/cyberark/secretless-broker/compare/v0.5.0...v0.5.1
[0.5.2]: https://github.com/cyberark/secretless-broker/compare/v0.5.1...v0.5.2
[0.6.0]: https://github.com/cyberark/secretless-broker/compare/v0.5.2...v0.6.0
[0.6.1]: https://github.com/cyberark/secretless-broker/compare/v0.6.0...v0.6.1
[0.6.2]: https://github.com/cyberark/secretless-broker/compare/v0.6.1...v0.6.2
[0.6.3]: https://github.com/cyberark/secretless-broker/compare/v0.6.2...v0.6.3
[0.6.4]: https://github.com/cyberark/secretless-broker/compare/v0.6.3...v0.6.4
[0.7.0]: https://github.com/cyberark/secretless-broker/compare/v0.6.4...v0.7.0 
[0.7.1]: https://github.com/cyberark/secretless-broker/compare/v0.7.0...v0.7.1 
[0.8.0]: https://github.com/cyberark/secretless-broker/compare/v0.7.1...v0.8.0 
[1.0.0]: https://github.com/cyberark/secretless-broker/compare/v0.8.0...v1.0.0 
[1.1.0]: https://github.com/cyberark/secretless-broker/compare/v1.0.0...v1.1.0 
[1.2.0]: https://github.com/cyberark/secretless-broker/compare/v1.1.0...v1.2.0 
