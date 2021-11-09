# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.7.8] - 2021-11-09

### Fixed
- Version bump to resolve flakey test on tagged master.
  [cyberark/secretless-broker#1438](https://github.com/cyberark/secretless-broker/pull/1438)

## [1.7.7] - 2021-11-03

### Fixed
- Request-signing on the AWS connector was updated to address a bug that was
  causing failed integrity checks, where the request-signing by Secretless was
  incorporating more headers than were used on the original request-signing. The
  fix limits the headers used by Secretless to those used in the original
  request. [cyberark/secretless-broker#1432](https://github.com/cyberark/secretless-broker/issues/1432)

### Security
- Updated containerd to v1.4.11 to close CVE-2020-15257 (Not vulnerable)
  [cyberark/secretless-broker#1431](https://github.com/cyberark/secretless-broker/pull/1431)

## [1.7.6] - 2021-09-10

### Added
- Secretless and secretless-redhat containers now use Alpine 3.14 as their base
  image. [PR cyberark/secretless-broker#1423](https://github.com/cyberark/secretless-broker/pull/1423)

## [1.7.5] - 2021-08-04

### Security
- Updated addressable to 2.8.0 in docs/Gemfile.lock to resolve GHSA-jxhc-q857-3j6g
  [cyberark/secretless-broker#1418](https://github.com/cyberark/secretless-broker/pull/1418)
- Updated github.com/gogo/protobuf to 1.3.2 to resolve CVE-2021-3121
  [cyberark/secretless-broker#1418](https://github.com/cyberark/secretless-broker/pull/1418)

## [1.7.4] - 2021-06-30

### Changed
- Update RH base image to `ubi8/ubi` instead of `rhel7/rhel`.
  [PR cyberark/secretless-broker#1411](https://github.com/cyberark/secretless-broker/pull/1411)

## [1.7.3] - 2021-03-09

### Changed
- Updated k8s authenticator client version to
  [0.19.1](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/CHANGELOG.md#0191---2021-02-08),
  which streamlines the parsing of authentication responses, updates the
  project Golang version to v1.15, and improves error messaging.

### Fixed
- When configured for the SSL mode of `require` or `prefer`, Secretless now sends
  a valid "SSL is not supported" response per the PostgreSQL protocol standard when
  a client attempts to open an SSL connection using the PostgreSQL connector. When
  the client is configured for SSL mode `prefer`, the updated response enables the
  client to downgrade to an insecure connection and continue. Previously, clients
  sending requests using the SSL mode of either `require` or `prefer` would receive
  a generic error from Secretless, which made it harder to determine the root cause
  of the problem and caused the `prefer` SSL mode to not function correctly.
  [cyberark/secretless-broker#1377](https://github.com/cyberark/secretless-broker/issues/1377)

### Deprecated
- Support for OpenShift 4.3 has been deprecated as of this release.

### Added
- Support for OpenShift 4.6 has been certified as of this release.
- Support for OpenShift 4.7 has been certified as of this release.

## [1.7.2] - 2021-02-05

### Added
- Support for OpenShift 4.3 and 4.5.
  [conjurdemos/kubernetes-conjur-demo#122](https://github.com/conjurdemos/kubernetes-conjur-demo/issues/122)

### Deprecated
- Support for OpenShift 3.9 and 3.10 is removed in this release.
  [conjurdemos/kubernetes-conjur-demo#122](https://github.com/conjurdemos/kubernetes-conjur-demo/issues/122)

### Fixed
- Automatic endpoint discovery for the AWS connector was updated to address two
  bugs where (1) the request host header was not being updated to the discovered
  endpoint, and (2) the request modification was being done after signing the
  request which would result in a failing integrity check.
  [cyberark/secretless-broker#1369](https://github.com/cyberark/secretless-broker/issues/1369)

## [1.7.1] - 2020-10-20

### Added
- The `vault` provider now supports loading secrets from the KV Version 2 secret
  engine. Reference a secret in Vault using the right path and a field
  navigation in the Secretless configuration.
  [cyberark/secretless-broker#1331](https://github.com/cyberark/secretless-broker/issues/1331)

### Changed
- Update k8s authenticator client version to
  [0.19.0](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/CHANGELOG.md#0190---2020-10-08),
  which adds some fixes around cert injection failure (see also changes in
  [0.18.1](https://github.com/cyberark/conjur-authn-k8s-client/blob/master/CHANGELOG.md#0181---2020-09-13)).
  [cyberark/secretless-broker#1352](https://github.com/cyberark/secretless-broker/pull/1352)

## [1.7.0] - 2020-09-11

### Added
- Secretless and secretless-redhat containers now use Alpine 3.12 as their base
  image. [PR cyberark/secretless-broker#1296](https://github.com/cyberark/secretless-broker/pull/1296)
- MySQL and PostgreSQL connectors support SSL host name verification with
  `verify-full` SSL mode. Also adds optional `sslhost` configuration parameter
  that is compared to the server's certificate SAN.
  [cyberark/secretless-broker#548](https://github.com/cyberark/secretless-broker/issues/548)
- Generic HTTP connector now supports `queryParam` as a configurable section
  in the Secretless configuration file, under `config`. This allows the
  construction of a query string which can have credentials injected
  as needed.
  [cyberark/secretless-broker#1290](https://github.com/cyberark/secretless-broker/issues/1290)
- Generic HTTP connector now supports `oauth1` as a configurable section in the
  secretless configuration file, under `config`. This allows the construction of
  a header for an OAuth 1.0 request. The OAuth 1.0 feature currently only supports
  HMAC-SHA1, but there is an [issue](https://github.com/cyberark/secretless-broker/issues/1324)
  logged to support other hashing methods.
  [cyberark/secretless-broker#1297](https://github.com/cyberark/secretless-broker/issues/1297)
- Many (20+) example generic connector configurations were added to the project,
  to demonstrate support for a broad set of popular APIs and to serve as an
  example for other APIs users may need to use Secretless with their apps.
  See [here](https://github.com/cyberark/secretless-broker/tree/master/examples/generic_connector_configs)
  for the full list of examples.
  [cyberark/secretless-broker#1248](https://github.com/cyberark/secretless-broker/issues/1248)

## [1.6.0] - 2020-05-04

### Added
- Support for a `SECRETLESS_HTTP_CA_BUNDLE` environment variable that specifies
  the path to a CA cert bundle and enables users to configure Secretless with
  additional CA certificates for server cert verification when using HTTP
  connectors.
  [PR #1180](https://github.com/cyberark/secretless-broker/pull/1180)
- TLS support for the Secretless-to-server connections of the MSSQL connector.
  This is the recommended way to secure this connection and achieves feature
  parity with other TLS connectors.
  [#1163](https://github.com/cyberark/secretless-broker/issues/1163),
  [#1164](https://github.com/cyberark/secretless-broker/issues/1164),
  [#1165](https://github.com/cyberark/secretless-broker/issues/1165)
- MSSQL connector supports SSL host name verification with `verify-full` SSL
  mode. Also adds optional `sslhost` configuration parameter that is compared to
  the server's certificate SAN.
  [#1199](https://github.com/cyberark/secretless-broker/issues/1199)

### Fixed
- PostgreSQL connector log messages were updated to improve formatting, fixing
  a previous issue where the log messages were improperly formatted and were
  garbled in the logs. [PR #1192](https://github.com/cyberark/secretless-broker/pull/1192)

### Security
- TCP connectors all automatically zeroize the connection credentials in memory
  after successfully opening a connection; previously, credentials were only
  zeroized in memory on error. [#1188](https://github.com/cyberark/secretless-broker/issues/1188)

## [1.5.2] - 2020-02-24

### Changed
- Bump authn-k8s client to v0.16.1 (cyberark/conjur-authn-k8s-client#70)

### Fixed
- Updated RH image push to ensure we're logged into the RH container registry
  appropriately before pushing (#1149)
- Fixed a stack overflow issue when running multiple multiple connections to an MsSQL
  server consecutively

## [1.5.1] - 2020-02-12

### Added
- Added RedHat certified image build to pipeline (#1141)
- Added pipeline step to validate changelog (#1138)
- Added MSSQL support to juxtaposer perf testing tool (#1135)
- Added SIGPIPE to signals handled by Secretless Juxtaposer (#1136)
- Added JDBC Integration tests for Postgres (#1130)
- Added JDBC Tests for MSSQL (#1124)
- Added client params propagation to MSSQL integration tests (#1103)

### Changed
- Default logging level changed from `Warn` to `Info`. Some logging message
  levels were readjusted to retain the same UX. (#1127)
- Update `bin/prefill_changelog` to generate valid CHANGELOG / ensure current
  CHANGELOG parses (#1138)
- Converted integration tests to use configs.v2 (#1120)

### Fixed
- Fixed broken documentation links (#1122)

## [1.5.0] - 2020-01-29

### Added
- Added option to specify MSSQL edition in tests (#1093)
- Added debug image that can be used with a debugger like delve (#1056)
- Added template READMEs to connector templates (#1020)

### Changed
- Updated release instructions (#1080)
- Improved MSSQL connector tests (#1107, #1089, #1098)
- Improved handling of `io.EOF` errors on TCP `proxy_service`
- Conjur authn-k8s client version bumped to v0.16.0
- Added links to SDK docs in README (#1104)
- Ensure external connector plugins will not override built-in connectors (#1085)
- MSSQL connector moved to beta

### Fixed
- Updated pg connector to better validate packet length (#1095)
- MSSQL connector faithfully propagates login response (#1106)
- MSSQL connector faithfully propagates login request (#1107)

## [1.4.2] - 2020-01-08

### Added
- Updated CONTRIBUTING.md with instructions for using `go-mssqldb` submodule (#1044)
- Added gosec security scan to pipeline (#976)
- Added integration tests for MSSQL against additional MSSQL versions (#1017)
- Added `gofmt` to CodeClimate checks (#1055)
- Added support for MSSQL client parameter propagation (#1012)

### Changed
- Bumped the `conjur-authn-k8s-client` version for the Conjur provider k8s
  authenticator to `v0.15.0` (#1060)
- Example plugin updated for clarity (#1061)
- Plugin SDK templates updated for clarity (#1054)
- Removed hardcoded PreloginResponse from MSSQL connector (#1014)
- Bumped Go version in Dockerfile to 1.13

### Fixed
- Secretless doesn't exit when it can't start a configured connector (#1057)
- Secretless has insufficient logs when the config file has trouble loading (#1062)

## [1.4.1] - 2019-12-11

### Added
- Added [README](https://github.com/cyberark/secretless-broker/blob/master/internal/plugin/connectors/tcp/mssql/README.md) for the MSSQL connector (#1003)

### Changed
- Added `go-mssqldb` dependency as a submodule (#1038)

### Fixed
- Updated Conjur provider to log and exit on repeated authentication failure
  (#1035)

## [1.4.0] - 2019-12-04

### Added
- Added generic HTTP connector to enable writing new HTTP connectors via
  config (#995)

### Changed
- Improved logs for k8s CRD test failure debugging (#1027)
- Updated Ruby version in docs container (#1028)
- Updated Conjur HTTP connector to leverage the generic HTTP connector (#1009)
- Reorganized integration tests (#958)
- Updated Basic Auth HTTP connector to leverage the generic HTTP connector
  (#1007)
- Replaced "honnef.co/go/tools" dependency in go.sum with a github link
- Updated "ozzo-validation" dependency to latest version
- Make forceSSL setting explicit in e2e tests

## [1.3.0] - 2019-11-18

### Added
- Added trivy security scan to project pipeline (#986)
- Added unit tests to ConfigEnv, profile and signal packages
- Added alpha MSSQL connector (#964)
- Added template skeleton for connector plugins (#967)

### Changed
- Extract config validation from ProxyServices and add unit tests
- Improved available_plugins unit tests
- Updated juxtaposer configs for perf tests (#969)

### Fixed
- Ensure MySQL uses appropriate default sslmode value (#928)
- Improved pg error propagation (#974)

## [1.2.0] - 2019-10-21

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

## [1.1.0] - 2019-08-09

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

## [1.0.0] - 2019-07-03

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

## [0.8.0] - 2019-06-18

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

## [0.7.1] - 2019-05-16

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

## [0.7.0] - 2019-03-26

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

## [0.6.4] - 2019-02-01

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

## [0.6.3] - 2019-01-11

### Added
- Database handlers support private-key pair as sslkey and sslcert

### Changed
- Permissions have been fixed for OpenShift non-root integration and use

## [0.6.2] - 2019-01-09

### Added
- Added Kubernetes authenticator documentation for Conjur credential provider

### Changed
- Sanitized remaining listeners/handlers from dumping data on the CLI when debug mode is on
- Removed developer-only debug mode from demos and examples

## [0.6.1] - 2019-01-08

### Changed
- Updated conjur-api-go dependency

### Added
- Added `/ready` and `/live` endpoints on port 5335 for checking if the broker is ready/live

## [0.6.0] - 2018-12-20

### Added
- SSL support for MySQL and PostgreSQL handlers
- Improved test utilities
- Added flag for CPU or memory profiling

### Changed
- Updated demos to support databases configured with SSL
- Allow ./bin/test_integration to specify individual test_folders + local flag
- Updated goreleaser process to use new image

## [0.5.2] - 2018-11-26

### Fixed
- Updated Kubernetes secrets provider to retrieve secrets from current namespace
- Fixed broken GitLab build referencing non-existent image
- Fixed broken keychain provider tests, and made easier to run manually

## [0.5.1] - 2018-11-20

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

## [0.5.0] - 2018-09-06

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

## [0.4.0] - 2018-08-02

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

## [0.3.0] - 2018-06-28

### Added
- Connection managers can be loaded with factories
- Listeners, handlers and managers can all now run from external plugins
- External plugin versioning now enforced
- Multi-stage container builds used
- Plugin test is now part of our CI pipeline
- Ability to notify connection managers of graceful shutdowns
- Added helper for creating changelog entries

### Changed
- Internal listeners and handlers use the same plugin architecture as external plugins
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

### Added
- The first tagged version.

[Unreleased]: https://github.com/cyberark/secretless-broker/compare/v1.7.8...HEAD
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
[1.3.0]: https://github.com/cyberark/secretless-broker/compare/v1.2.0...v1.3.0
[1.4.0]: https://github.com/cyberark/secretless-broker/compare/v1.3.0...v1.4.0
[1.4.1]: https://github.com/cyberark/secretless-broker/compare/v1.4.0...v1.4.1
[1.4.2]: https://github.com/cyberark/secretless-broker/compare/v1.4.1...v1.4.2
[1.5.0]: https://github.com/cyberark/secretless-broker/compare/v1.4.2...v1.5.0
[1.5.1]: https://github.com/cyberark/secretless-broker/compare/v1.5.0...v1.5.1
[1.5.2]: https://github.com/cyberark/secretless-broker/compare/v1.5.1...v1.5.2
[1.6.0]: https://github.com/cyberark/secretless-broker/compare/v1.5.2...v1.6.0
[1.7.0]: https://github.com/cyberark/secretless-broker/compare/v1.6.0...v1.7.0
[1.7.1]: https://github.com/cyberark/secretless-broker/compare/v1.7.0...v1.7.1
[1.7.2]: https://github.com/cyberark/secretless-broker/compare/v1.7.1...v1.7.2
[1.7.3]: https://github.com/cyberark/secretless-broker/compare/v1.7.2...v1.7.3
[1.7.4]: https://github.com/cyberark/secretless-broker/compare/v1.7.3...v1.7.4
[1.7.5]: https://github.com/cyberark/secretless-broker/compare/v1.7.4...v1.7.5
[1.7.6]: https://github.com/cyberark/secretless-broker/compare/v1.7.5...v1.7.6
[1.7.7]: https://github.com/cyberark/secretless-broker/compare/v1.7.6...v1.7.7
[1.7.8]: https://github.com/cyberark/secretless-broker/compare/v1.7.7...v1.7.8
