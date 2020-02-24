# Changelog

## [v1.0.0] - 2020-02-24

### Major changes
- Custom Resource Definition moved to `databases.spotahome.com`
- Rolling updates are aware of cluster topology and nodes roles to follow minimum impact strategy
- Better readiness probes for redis nodes.
- More customizable options for kubernetes objects.
- Better bootstrap times.
- Improve security with password protected redis and security pod policies.
- Redis 5 as default version
- Update dependencies

For detailed changelogs see rc relases

## [v1.0.0-rc.5] - 2020-02-07

### Changes
- Custom annotations for services #216 @alecjacobs5401 
- Update redis-exporter #222 @VerosK
- Pod security policy to run as non root #228 @logdnalf
- Custom command renames #234 @logdnalf

### Fix
- Add fsGroup to security context #215 @ese
- Pod disruption budget lower than replicas #229 @tkrop
- Add password support for readiness probes #235 @teamon
  
## [v1.0.0-rc.4] - 2019-12-17

### Changes
- Update kooper to v0.8.0
- Update kubernetes to v1.15.6
- Add support for `hostNetwork` and `dnsPolicy` in Redis and Sentinel pods #212 @paol

## [v1.0.0-rc.3] - 2019-12-10

### Action required

Since update logic has been moved to operator, `PodManagementPolicy` has been set to `Parallel` in redis statefulSet. This improve bootstrap times.
This field is immutable so to upgrade from previous rc releases you need to delete statefulSets manually. 
*Note:* you can use `--cascade=false` flag to avoid disruption, pods will be adopted by the new statefulSet created by the operator.
example: `kubectl delete statefulset --cascade=false rfr-redisfailover`

### Changes

- Move rolling update strategy to redis-operator to be cluster-aware #203 @chusAlvarez
- Readiness probe check nodes belong to the cluster and are synced #206 @chusAlvarez
- Support label propagation filter #195 @adamhf
- Support for sentinel prometheus exporter #207 @shonge

### Fix

- Documentation and examples #204 @shonge
- Add RBAC policy to access secrets #208 @hoffoo

## [v1.0.0-rc.2] - 2019-11-15

### Changes

- Add custom annotations for pods in the CRD `podAnnotations` @alecjacobs5401
- Add redis authentication @hoffoo
- Configurable imagePullSecret @romanfurst 
- Configurable imagePullPolicy @mcdiae 
- Support for node selector `nodeSelector` @sergeunity 

### Fix

- Add RBAC policy for the CRD finalizer @mcanevet
- Examples documentation  @SataQiu @marcemq 
- Chart service labels @timmyers
- Memory requests and limits for sentinel @marcemq
- Execution permissions in shutdown script @glebpom
- Makefile uid passthrough @adamhf

## [v1.0.0-rc.1] - 2019-05-10

### Changed

- Minimum Kubernetes version needed is 1.9.
- Custom Resource Definition moved to `databases.spotahome.com`.
- API version moved to v1.
- Standardize labels with the Kubernetes recommended ones.
- Update Kubernetes libraries to 1.11.9.
- Update Kooper to v0.5.1.
- Update Golang used to 1.12.
- Use new versioning standard.

### Fixed

- Chart unused values removed.
- Remove double loops for checking Sentinels data in memory.

## [0.5.8] - 2019-03-26

### Fixed

- Now all errors makes a `redisfailover` be marked as failed on metrics, to prevent that some errors were never alerted.

## [0.5.7] - 2019-03-06

### Added

- Command for Redis and Sentinel containers is now configurable.

### Fixed

- Panic if checking the `StartTime` of a pod that was not started yet (nil pointer exception).

## [0.5.6] - 2019-02-27

### Added

- Add tolerations to Redis and Sentinel pods.

### Changed

- Improve management of `customConfig` so they admit any type of configuration.

## [0.5.5] - 2019-02-19

### Added

- Create flag to disable exporter probes.

### Changed

- Increase default memory.
- Improve readability of code.

## [0.5.4] - 2018-10-15

### Changed

- Improve the checker to make it more resilient.
- Reduce startup time.
- When force one master, choose the oldest one.

## [0.5.3] - 2018-09-18

### Added

- Limit length of redis-failovers name to prevent errors when creating the redis statefulsets.
- Add set as failure on metrics when cannot fix the status of redis/sentinel by the operator.
- Remove the redis-failover from metrics if deleted.

## [0.5.2] - 2018-09-04

### Changed

- Higher `InitialDelaySeconds` probes times.

### Fixed

- Default values for spec and validator (lost when release of 1alpha2 api version).

## [0.5.1] - 2018-09-03

### Added

- Persist Redis data on disk.

## [0.5.0] - 2018-08-24

### Added

- Add redis and sentinel custom configuration array.

### Removed

- A `ConfigMap` name for the custom configuration is no longer available.

## [0.4.1] - 2018-08-17

### Added

- Elect a new master when the master pod is terminated.

## [0.4.0] - 2018-07-18

### Added

- Persistence for Redis data in persistent volumes is now available.

## [0.3.0] - 2018-07-03

### Added

- Make name of the Redis Operator container configurable.

### Changed

- Update kooper to v0.3.0, updating the Kubernetes clients to v1.10.5.

## [0.2.5] - 2018-05-25

### Added

- Add the possibility to use a volumen for redis data.

### Changed

- Use the RedisImage to copy the Sentinel configuration in order to use one image less.

## [0.2.4] - 2018-05-24

### Added

- Add the possibility to set the configMap to be used on both Redis and Sentinel.
- Add the possibility to set the redis/sentinel image.
- Add the possibility to set the redis-exporter image and version.

## [0.2.3] - 2018-04-06

### Added

- Add the possibility to use a `NodeAffinity`.

## [0.2.2] - 2018-04-06

### Added

- Add Prometheus Annotations to Redis Exporter.

## [0.2.1] - 2018-03-28

### Fixed

- Create a init-container on sentinel pods so the sentinel.conf is writable.

## [0.2.0] - 2018-02-19

### Added

- Use [Kooper](https://github.com/spotahome/kooper).
- New API version: `storage.spotahome.com/v1alpha2`.

### Changed

- Simplified metrics.
- New client that allows interaction with the redis failovers created.
- New ensurer that checks all pieces are created.
- New checker and healer that puts the nodes into their expected state.

### Removed

- There is no path for upgrade from <0.2.0. You need to create new resources and delete the deprecated CRD resource with `kubectl delete crd redisfailovers.spotahome.com`.

## [0.1.6] - 2018-02-01

### Added

- Add flag to disable `hardaffinity`.
- Wait for CDR before running operator.

## [0.1.5] - 2018-01-03

### Added

- Ensure scheduling on different nodes.
- Export port for gather metrics.
- Add service to chart.

### Changed

- Change waiters so not blocking multiple edits of same resources.

### Fixed

- Only add the redis exporter container if it does not exists.

## [0.1.4] - 2018-01-02

### Added

- Add timeout on waiters.

### Fixed

- Fix WaitForPod unlimited waiting.

## [0.1.3] - 2017-12-29

### Added

- Add/Delete exporter when updating.

### Changed

- Refactor waiters.
- Change concurrency approach. New default limits.

## [0.1.2] - 2017-12-18

### Fixed

- Change kind of response when calling sentinel.

## 0.1.1 - 2017-12-15

### Added

- Initial open-sourced release

[v1.0.0]: https://github.com/spotahome/redis-operator/compare/0.5.8...v1.0.0
[v1.0.0-rc.5]: https://github.com/spotahome/redis-operator/compare/v1.0.0-rc.4...v1.0.0-rc.5
[v1.0.0-rc.4]: https://github.com/spotahome/redis-operator/compare/v1.0.0-rc.3...v1.0.0-rc.4
[v1.0.0-rc.3]: https://github.com/spotahome/redis-operator/compare/v1.0.0-rc.2...v1.0.0-rc.3
[v1.0.0-rc.2]: https://github.com/spotahome/redis-operator/compare/v1.0.0-rc.1...v1.0.0-rc.2
[v1.0.0-rc.1]: https://github.com/spotahome/redis-operator/compare/0.5.8...v1.0.0-rc.1
[0.5.8]: https://github.com/spotahome/redis-operator/compare/0.5.7...0.5.8
[0.5.7]: https://github.com/spotahome/redis-operator/compare/0.5.6...0.5.7
[0.5.6]: https://github.com/spotahome/redis-operator/compare/0.5.5...0.5.6
[0.5.5]: https://github.com/spotahome/redis-operator/compare/0.5.4...0.5.5
[0.5.4]: https://github.com/spotahome/redis-operator/compare/0.5.3...0.5.4
[0.5.3]: https://github.com/spotahome/redis-operator/compare/0.5.2...0.5.3
[0.5.2]: https://github.com/spotahome/redis-operator/compare/0.5.1...0.5.2
[0.5.1]: https://github.com/spotahome/redis-operator/compare/0.5.0...0.5.1
[0.5.0]: https://github.com/spotahome/redis-operator/compare/0.4.1...0.5.0
[0.4.1]: https://github.com/spotahome/redis-operator/compare/0.4.0...0.4.1
[0.4.0]: https://github.com/spotahome/redis-operator/compare/0.3.0...0.4.0
[0.3.0]: https://github.com/spotahome/redis-operator/compare/0.2.5...0.3.0
[0.2.5]: https://github.com/spotahome/redis-operator/compare/0.2.4...0.2.5
[0.2.4]: https://github.com/spotahome/redis-operator/compare/0.2.3...0.2.4
[0.2.3]: https://github.com/spotahome/redis-operator/compare/0.2.2...0.2.3
[0.2.2]: https://github.com/spotahome/redis-operator/compare/0.2.1...0.2.2
[0.2.1]: https://github.com/spotahome/redis-operator/compare/0.2.0...0.2.1
[0.2.0]: https://github.com/spotahome/redis-operator/compare/0.1.6...0.2.0
[0.1.6]: https://github.com/spotahome/redis-operator/compare/0.1.5...0.1.6
[0.1.5]: https://github.com/spotahome/redis-operator/compare/0.1.4...0.1.5
[0.1.4]: https://github.com/spotahome/redis-operator/compare/0.1.3...0.1.4
[0.1.3]: https://github.com/spotahome/redis-operator/compare/0.1.2...0.1.3
[0.1.2]: https://github.com/spotahome/redis-operator/compare/0.1.1...0.1.2
