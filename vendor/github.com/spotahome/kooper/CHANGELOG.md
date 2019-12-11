## [unreleased]

- Support for Kubernetes 1.13.

## [0.6.0] - 2019-06-01

### Added

- Support for Kubernetes 1.12.

### Removed

- Glog logger.

## [0.5.1] - 2019-01-19

### Added

- Shortnames on CRD registration.

## [0.5.0] - 2018-10-24

### Added

- Support for Kubernetes 1.11.

## [0.4.1] - 2018-10-07

### Added

- Enable subresources support on CRD registration.
- Category support on CRD registration.

## [0.4.0] - 2018-07-21

This release breaks Prometheus metrics.

### Added

- Grafana dashboard for the refactored Prometheus metrics.

### Changed

- Refactor metrics in favor of less metrics but simpler and more meaningful.

## [0.3.0] - 2018-07-02

This release breaks handler interface to allow passing a context (used to allow tracing).

### Added

- Context as first argument to handler interface to pass tracing context (Breaking change).
- Tracing through opentracing.
- Leader election for controllers and operators.
- Let customizing (using configuration) the retries of event processing errors on controllers.
- Controllers now can be created using a configuration struct.
- Add support for Kubernetes 1.10.

## [0.2.0] - 2018-02-24

This release breaks controllers constructors to allow passing a metrics recorder backend.

### Added

- Prometheus metrics backend.
- Metrics interface.
- Concurrent controller implementation.
- Controllers record metrics about queued and processed events.

### Fixed

- Fix passing a nil logger to make controllers execution break.

## [0.1.0] - 2018-02-15

### Added

- CRD client check for kubernetes apiserver (>=1.7)
- CRD ensure (waits to be present after registering a CRD)
- CRD client tooling
- multiple CRD and multiple controller operator.
- single CRD and single controller operator.
- sequential controller implementation.
- Dependencies managed by dep and vendored.

[unreleased]: https://github.com/spotahome/kooper/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/spotahome/kooper/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/spotahome/kooper/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/spotahome/kooper/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/spotahome/kooper/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/spotahome/kooper/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/spotahome/kooper/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/spotahome/kooper/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/spotahome/kooper/releases/tag/v0.1.0
