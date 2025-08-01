# Changelog

## [1.3.0](https://github.com/g2a-com/klio/compare/v1.2.0...v1.3.0) (2025-07-29)


### Features

* add --upgrade flag ([#88](https://github.com/g2a-com/klio/issues/88)) ([deffa47](https://github.com/g2a-com/klio/commit/deffa4742c8279a8c0efb23ad9bb1d565aa82fd0))

## [1.2.0](https://github.com/g2a-com/klio/compare/v1.1.4...v1.2.0) (2025-02-10)


### Features

* auto download divergent project commands ([#86](https://github.com/g2a-com/klio/issues/86)) ([bf2893a](https://github.com/g2a-com/klio/commit/bf2893adb1b03eeb8b3efba192cc2fdbb477a78f))

## [1.1.4](https://github.com/g2a-com/klio/compare/v1.1.3...v1.1.4) (2024-10-10)


### Bug Fixes

* properly format warn log ([#80](https://github.com/g2a-com/klio/issues/80)) ([db1e9de](https://github.com/g2a-com/klio/commit/db1e9de67ae92d35fb80250c46040a4a7a916fc3))

## [1.1.3](https://github.com/g2a-com/klio/compare/v1.1.2...v1.1.3) (2024-07-30)


### Bug Fixes

* remove unused release-please param ([#78](https://github.com/g2a-com/klio/issues/78)) ([383091d](https://github.com/g2a-com/klio/commit/383091dfcca16801e4ed421d0cb640c527fb0566))

### [1.1.2](https://www.github.com/g2a-com/klio/compare/v1.1.1...v1.1.2) (2024-06-03)


### Bug Fixes

* update flush func to respect log level and add new line ([#73](https://www.github.com/g2a-com/klio/issues/73)) ([e503488](https://www.github.com/g2a-com/klio/commit/e5034886ec5411ebe2d3a59a61d87bccb465495c))

### [1.1.1](https://www.github.com/g2a-com/klio/compare/v1.1.0...v1.1.1) (2024-05-24)


### Bug Fixes

* remove bugged SetOuput and allow to create new logger instance ([#71](https://www.github.com/g2a-com/klio/issues/71)) ([f8ec6fb](https://www.github.com/g2a-com/klio/commit/f8ec6fbc62d432cb38966295c9605cbdb426c0d8))

## [1.1.0](https://www.github.com/g2a-com/klio/compare/v1.0.1...v1.1.0) (2024-05-21)


### Features

* **g2a-com#48:** expose internal logger in pkg ([#69](https://www.github.com/g2a-com/klio/issues/69)) ([362e2df](https://www.github.com/g2a-com/klio/commit/362e2df6e68838419a366acb9014076a8aac44e0))
* **g2a-com#52:** add 'remove' command ([#59](https://www.github.com/g2a-com/klio/issues/59)) ([d05de37](https://www.github.com/g2a-com/klio/commit/d05de37757853add6796a84fdfcd18f2e7519aa5))


### Bug Fixes

* change Manager interface, remove fatal with defer in place ([24a0f6f](https://www.github.com/g2a-com/klio/commit/24a0f6fd65048e5165b5fe0dec455ecab1b42c40))
* Global scope in local scope (should be project) ([fab3765](https://www.github.com/g2a-com/klio/commit/fab3765551b54a3ad0a67962cdead9ef88f03dff))
* success message and implicit dependencies ([2ad5504](https://www.github.com/g2a-com/klio/commit/2ad55047c021d3f278591243a996a350a1fdd1aa))
* traverse the whole path to find project config ([0234848](https://www.github.com/g2a-com/klio/commit/02348485061170f568eb87928a1d6e10037f24ad))
* various fixes after refactor ([bc081ef](https://www.github.com/g2a-com/klio/commit/bc081efcdf2bbb28cc39a4bbf917810f8d712943))
* wrong logging on exit ([ca1137e](https://www.github.com/g2a-com/klio/commit/ca1137e8c3ecb635709543abbf6283d3836c5a3c))
