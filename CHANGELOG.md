# Changelog

## [1.12.0](https://github.com/soerenschneider/dyndns/compare/v1.11.1...v1.12.0) (2023-06-02)


### Features

* add further metrics to enhance observability ([29e14b3](https://github.com/soerenschneider/dyndns/commit/29e14b33ba8aaff2cbfa5eeb4ab3b16f051658b5))
* add simple hash function to compare known hosts across hosts ([162b073](https://github.com/soerenschneider/dyndns/commit/162b0739326c007ac0f9a25ffcdeeafafc8a76c2))
* auto reload updated certificates ([b512a7a](https://github.com/soerenschneider/dyndns/commit/b512a7af74b7029bbcf3d0e4b87d7b9831cd852f))
* configurable address families ([89f05f8](https://github.com/soerenschneider/dyndns/commit/89f05f86f6fc622477ad80a387cb12322220a3f2))
* validate using 'go-playground/validator' ([3db56f7](https://github.com/soerenschneider/dyndns/commit/3db56f72314e2a3be90daabd7b1f5286f5526f36))


### Bug Fixes

* change validation flag ([9acd257](https://github.com/soerenschneider/dyndns/commit/9acd257a4019e060b3f4bd35408eec36e3cd54b1))
* define custom validator to fix panic ([7ca9cd5](https://github.com/soerenschneider/dyndns/commit/7ca9cd50dfc6d34eebe81e67c33249c55072326a))
* iterate over correct datastructure ([2c3bfcd](https://github.com/soerenschneider/dyndns/commit/2c3bfcdb13875574745e632320099a95b9e03b38))

## [1.11.1](https://github.com/soerenschneider/dyndns/compare/v1.11.0...v1.11.1) (2023-01-31)


### Bug Fixes

* add missing label ([6952b85](https://github.com/soerenschneider/dyndns/commit/6952b85da62f88b81c62af30d285d57218a9e020))

## [1.11.0](https://github.com/soerenschneider/dyndns/compare/v1.10.0...v1.11.0) (2023-01-31)


### Features

* Notifications and preferred resolvers ([#216](https://github.com/soerenschneider/dyndns/issues/216)) ([c294004](https://github.com/soerenschneider/dyndns/commit/c2940047cbb13b42825698d068f904d1a47d8b32))

## [1.10.0](https://github.com/soerenschneider/dyndns/compare/v1.9.0...v1.10.0) (2023-01-28)


### Features

* **server:** support multiple brokers simultaneously ([#212](https://github.com/soerenschneider/dyndns/issues/212)) ([12011da](https://github.com/soerenschneider/dyndns/commit/12011da5964346e7c026b184ab953fee2b22534d))

## [1.9.0](https://github.com/soerenschneider/dyndns/compare/v1.8.0...v1.9.0) (2023-01-27)


### Features

* decouple detection and change propagation  ([#210](https://github.com/soerenschneider/dyndns/issues/210)) ([c687a2d](https://github.com/soerenschneider/dyndns/commit/c687a2d30798257b3a8831a3ce5aad0ac0a63fba))

## [1.8.0](https://github.com/soerenschneider/dyndns/compare/v1.7.0...v1.8.0) (2023-01-15)


### Features

* add metric for currently active status ([72712f6](https://github.com/soerenschneider/dyndns/commit/72712f686ccbdee0b49886e47c4362a1ad90afe0))

## [1.7.0](https://github.com/soerenschneider/dyndns/compare/v1.6.1...v1.7.0) (2022-11-29)


### Features

* add configurable urls ([#193](https://github.com/soerenschneider/dyndns/issues/193)) ([48fe7e1](https://github.com/soerenschneider/dyndns/commit/48fe7e1b04f3f5ba2731be8779cb0edda59d1dbe))

## [1.6.1](https://github.com/soerenschneider/dyndns/compare/v1.6.0...v1.6.1) (2022-11-27)


### Bug Fixes

* Use multiple brokers in client ([#190](https://github.com/soerenschneider/dyndns/issues/190)) ([204d872](https://github.com/soerenschneider/dyndns/commit/204d8728f6bbacf4194411c206bdeddcd94684d4))

## [1.6.0](https://github.com/soerenschneider/dyndns/compare/v1.5.4...v1.6.0) (2022-11-26)


### Features

* support tls config ([20bcf82](https://github.com/soerenschneider/dyndns/commit/20bcf82d1e311729870c6c02939d99b2a2f73234))

## [1.5.4](https://github.com/soerenschneider/dyndns/compare/v1.5.3...v1.5.4) (2022-08-31)


### Bug Fixes

* don't initialize slice ([ad5da23](https://github.com/soerenschneider/dyndns/commit/ad5da2361fe34ffeca6082289021015baec49ad7))

## [1.5.4](https://github.com/soerenschneider/dyndns/compare/v1.5.3...v1.5.4) (2022-08-31)


### Bug Fixes

* don't initialize slice ([ad5da23](https://github.com/soerenschneider/dyndns/commit/ad5da2361fe34ffeca6082289021015baec49ad7))

## [1.5.3](https://github.com/soerenschneider/dyndns/compare/v1.5.2...v1.5.3) (2022-08-29)


### Bug Fixes

* don't fail on non-existent users ([c78bf83](https://github.com/soerenschneider/dyndns/commit/c78bf8320efc05af2d10e975a2e2c8b37fdbca95))

## [1.5.2](https://github.com/soerenschneider/dyndns/compare/v1.5.1...v1.5.2) (2022-08-20)


### Bug Fixes

* fix metrics listener config ([8e1fc2c](https://github.com/soerenschneider/dyndns/commit/8e1fc2cef32afccdc874536841975223910a00b0))

## [1.5.1](https://github.com/soerenschneider/dyndns/compare/v1.5.0...v1.5.1) (2022-08-20)


### Bug Fixes

* use configured client-id in client mode ([37b5b69](https://github.com/soerenschneider/dyndns/commit/37b5b696cb9beed2ab5cf329910a468c48660029))

## [1.5.0](https://github.com/soerenschneider/dyndns/compare/v1.4.1...v1.5.0) (2022-08-18)


### Features

* Allow reading config from env vars ([#152](https://github.com/soerenschneider/dyndns/issues/152)) ([4b7f669](https://github.com/soerenschneider/dyndns/commit/4b7f66976f3cd17516eec28dbfcadcd399a9bc51))
* Support multiple keypairs per host ([#154](https://github.com/soerenschneider/dyndns/issues/154)) ([edca9a5](https://github.com/soerenschneider/dyndns/commit/edca9a545cae598652b49354ad63d0ce86ee5d2d))

### [1.4.1](https://www.github.com/soerenschneider/dyndns/compare/v1.4.0...v1.4.1) (2022-06-02)


### Bug Fixes

* introduce build tags to suppress unneeded metrics ([#130](https://www.github.com/soerenschneider/dyndns/issues/130)) ([25f48a3](https://www.github.com/soerenschneider/dyndns/commit/25f48a30fa90190ebded1d705508dfab8978c67f))

## [1.4.0](https://www.github.com/soerenschneider/dyndns/compare/v1.3.0...v1.4.0) (2022-06-02)


### Features

* Multiple brokers ([#125](https://www.github.com/soerenschneider/dyndns/issues/125)) ([1cce676](https://www.github.com/soerenschneider/dyndns/commit/1cce67685c8ec0b3b9501a4bf417f059100f2776))

## [1.3.0](https://www.github.com/soerenschneider/dyndns/compare/v1.2.0...v1.3.0) (2022-05-07)


### Features

* require discovered ipv4 is not private and not loopback ([1ea3f45](https://www.github.com/soerenschneider/dyndns/commit/1ea3f450dedb8aa21fcbf7fa2a97cc22fb0c4bfc))


### Bug Fixes

* do not start metrics server with '-once' flag ([6f8759d](https://www.github.com/soerenschneider/dyndns/commit/6f8759d3487fc3fd84acb4ef5a7de738dfc4800e))

## [1.2.0](https://www.github.com/soerenschneider/dyndns/compare/v1.1.2...v1.2.0) (2021-09-28)


### Features

* print version ([529cc31](https://www.github.com/soerenschneider/dyndns/commit/529cc31d80f809dd29141edf654ced55a8e8ccf7))

### [1.1.2](https://www.github.com/soerenschneider/dyndns/compare/v1.1.1...v1.1.2) (2021-07-03)


### Bug Fixes

* Fix incorrect, duplicate logmessage ([b5cad1e](https://www.github.com/soerenschneider/dyndns/commit/b5cad1e0dcb6925db4479d25bba546afef8624c2))

### [1.1.1](https://www.github.com/soerenschneider/dyndns/compare/v1.1.0...v1.1.1) (2021-06-08)


### Bug Fixes

* fix typo ([009aae2](https://www.github.com/soerenschneider/dyndns/commit/009aae24146c149b09e6df8855df2f8f32a6d2f6))

## [1.1.0](https://www.github.com/soerenschneider/dyndns/compare/v1.0.0...v1.1.0) (2021-06-07)


### Features

* Use better logging ([a7bb520](https://www.github.com/soerenschneider/dyndns/commit/a7bb520ac029aba2cff8a02c7a2bc9332c219444))

## 1.0.0 (2021-06-07)


### Features

* Check if entry is already propagated ([af63491](https://www.github.com/soerenschneider/dyndns/commit/af634914943b4cf66bcd86c9321053292401bbe1))


### Bug Fixes

* Resolve interface issue ([b70ea24](https://www.github.com/soerenschneider/dyndns/commit/b70ea24af91a8d42ca48a6cdb8eda606dc83c63b))
