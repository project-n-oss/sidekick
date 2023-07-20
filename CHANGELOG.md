# Changelog

## [0.1.24](https://github.com/project-n-oss/sidekick/compare/v0.1.23...v0.1.24) (2023-07-20)


### Bug Fixes

* fix 2 typos in readme ([#52](https://github.com/project-n-oss/sidekick/issues/52)) ([82eea5c](https://github.com/project-n-oss/sidekick/commit/82eea5c1b2d2451ead3b1c10dcb02286bfe042ed))

## [0.1.23](https://github.com/project-n-oss/sidekick/compare/v0.1.22...v0.1.23) (2023-07-20)


### Bug Fixes

* new release ([b4cf490](https://github.com/project-n-oss/sidekick/commit/b4cf490c98c86513eaf69ba2caf2e5d3568839d5))

## [0.1.22](https://github.com/project-n-oss/sidekick/compare/v0.1.21...v0.1.22) (2023-07-20)


### Features

* **KRY-624:** A more intelligent SideKick ([#47](https://github.com/project-n-oss/sidekick/issues/47)) ([0735c85](https://github.com/project-n-oss/sidekick/commit/0735c85e243d61abd7700e89cdaffef58ada9f72))

## [0.1.21](https://github.com/project-n-oss/sidekick/compare/v0.1.20...v0.1.21) (2023-07-17)


### Features

* **KRY-627:** set X-Bolt-Availability-Zone header in Bolt requests ([#46](https://github.com/project-n-oss/sidekick/issues/46)) ([951623b](https://github.com/project-n-oss/sidekick/commit/951623b7e0bb02cdee535ed897556629382de95c))

## [0.1.20](https://github.com/project-n-oss/sidekick/compare/v0.1.19...v0.1.20) (2023-07-05)


### Features

* default back to aws region ([#44](https://github.com/project-n-oss/sidekick/issues/44)) ([61096f1](https://github.com/project-n-oss/sidekick/commit/61096f121a821e5fd6ea9b77770f32aaf22af551))

## [0.1.19](https://github.com/project-n-oss/sidekick/compare/v0.1.18...v0.1.19) (2023-06-14)


### Bug Fixes

* fix default failover value ([#41](https://github.com/project-n-oss/sidekick/issues/41)) ([d23fa41](https://github.com/project-n-oss/sidekick/commit/d23fa4148c0c9a2cba34cf83b81bfe32aaa3bacc))

## [0.1.18](https://github.com/project-n-oss/sidekick/compare/v0.1.17...v0.1.18) (2023-06-13)


### Features

* changed BOLT_CUSTOM_DOMAIN to GRANICA_CUSTOM_DOMAIN ([6575e0b](https://github.com/project-n-oss/sidekick/commit/6575e0b22be9cafb3a9aee36529e22165b4cff50))


### Bug Fixes

* fix terminal logger for windows ([9336da4](https://github.com/project-n-oss/sidekick/commit/9336da40895efbee6befabc3e218a9b4182724e3))

## [0.1.17](https://github.com/project-n-oss/sidekick/compare/v0.1.16...v0.1.17) (2023-06-11)


### Features

* put object tests and better logging ([#38](https://github.com/project-n-oss/sidekick/issues/38)) ([43e84ab](https://github.com/project-n-oss/sidekick/commit/43e84ab091188c9b3fb7d3f4c8b77e22430edce8))

## [0.1.16](https://github.com/project-n-oss/sidekick/compare/v0.1.15...v0.1.16) (2023-06-06)


### Features

* add CyberDuck integration docs ([#29](https://github.com/project-n-oss/sidekick/issues/29)) ([1f09f9f](https://github.com/project-n-oss/sidekick/commit/1f09f9f31b4efda9005cc9e94d28caa1963a99f4))
* allow boltrouter config to be set from env ([#36](https://github.com/project-n-oss/sidekick/issues/36)) ([17aa433](https://github.com/project-n-oss/sidekick/commit/17aa433229efd7f367d7250fdb5018d1dd586132))
* changed BOLT_CUSTOM_DOMAIN to GRANICA_CUSTOM_DOMAIN ([#37](https://github.com/project-n-oss/sidekick/issues/37)) ([c19ff28](https://github.com/project-n-oss/sidekick/commit/c19ff28c674325238fd2e3bb5a7bbfa8782ae43b))
* cleaned up bolt_vars for non ec2 use ([#27](https://github.com/project-n-oss/sidekick/issues/27)) ([97829ed](https://github.com/project-n-oss/sidekick/commit/97829edce7fe041b3d3ab1f78c196c56338b2aba))
* Multi region support ([#30](https://github.com/project-n-oss/sidekick/issues/30)) ([bb62e1e](https://github.com/project-n-oss/sidekick/commit/bb62e1ecae7b95d93bd59fe532975b8fca12c876))
* refresh AWS credentials periodically ([#31](https://github.com/project-n-oss/sidekick/issues/31)) ([ffbba32](https://github.com/project-n-oss/sidekick/commit/ffbba32d3086404cede424acd76b757b68619495))
* S3 GUI client integration instructions ([#35](https://github.com/project-n-oss/sidekick/issues/35)) ([c3fc76f](https://github.com/project-n-oss/sidekick/commit/c3fc76f9535d9f443525d0670ad5369c10260957))


### Bug Fixes

* add instrucitons when using temp credentials ([#34](https://github.com/project-n-oss/sidekick/issues/34)) ([c30ec05](https://github.com/project-n-oss/sidekick/commit/c30ec05d369449235e5587266c32dad191cc400d))
* complete Docker instructions,  aware of different credential scenarios ([#33](https://github.com/project-n-oss/sidekick/issues/33)) ([3402935](https://github.com/project-n-oss/sidekick/commit/3402935dc83f9a218f9bb63676edcc42d3bf000f))

## [0.1.15](https://github.com/project-n-oss/sidekick/compare/v0.1.14...v0.1.15) (2023-05-04)


### Bug Fixes

* Fix unicode awsfailover path error ([7d61971](https://github.com/project-n-oss/sidekick/commit/7d6197102187793835ebcb535132177269e79e22))

## [0.1.14](https://github.com/project-n-oss/sidekick/compare/v0.1.13...v0.1.14) (2023-05-03)


### Bug Fixes

* fix docker tag versions ([4905807](https://github.com/project-n-oss/sidekick/commit/4905807c934105f48a41a5bf600153e8d723dc2c))

## [0.1.13](https://github.com/project-n-oss/sidekick/compare/v0.1.12...v0.1.13) (2023-05-03)


### Bug Fixes

* fix release.yml ([ad02307](https://github.com/project-n-oss/sidekick/commit/ad0230765ae4c0b0b4609329115a0cfd3bc5d8d1))

## [0.1.12](https://github.com/project-n-oss/sidekick/compare/v0.1.11...v0.1.12) (2023-05-03)


### Features

* added ci-cd for docker version tags ([#22](https://github.com/project-n-oss/sidekick/issues/22)) ([b231ae7](https://github.com/project-n-oss/sidekick/commit/b231ae78b9bc7ec88fb80347fb67f76fa475a8bc))

## [0.1.11](https://github.com/project-n-oss/sidekick/compare/v0.1.10...v0.1.11) (2023-04-25)


### Features

* added version output when running ([9cff822](https://github.com/project-n-oss/sidekick/commit/9cff822385d6ab8b44e639bd2bdf166b55cf06d1))

## [0.1.10](https://github.com/project-n-oss/sidekick/compare/v0.1.9...v0.1.10) (2023-04-25)


### Features

* added health check ([8f3b4da](https://github.com/project-n-oss/sidekick/commit/8f3b4da23d8a5fa9b081b0029354a5129edfd00a))


### Bug Fixes

* fix status code failover check ([040eb76](https://github.com/project-n-oss/sidekick/commit/040eb768f46fa2763fbee00571e74f123563e798))

## [0.1.9](https://github.com/project-n-oss/sidekick/compare/v0.1.8...v0.1.9) (2023-04-19)


### Features

* added docker image CICD ([#18](https://github.com/project-n-oss/sidekick/issues/18)) ([df827f0](https://github.com/project-n-oss/sidekick/commit/df827f0a5937695473208ee480d0541e204d6ea2))
* log warning for non 2xx aws response ([9593881](https://github.com/project-n-oss/sidekick/commit/95938819a760b083354c89fd1c75e55021a26f21))


### Bug Fixes

* fix non 2xx warn logging ([601ad53](https://github.com/project-n-oss/sidekick/commit/601ad532fa1e363925842d6bbf83844c28402065))

## [0.1.8](https://github.com/project-n-oss/sidekick/compare/v0.1.7...v0.1.8) (2023-04-12)


### Bug Fixes

* fix github action for docker registry ([079df58](https://github.com/project-n-oss/sidekick/commit/079df58b6f7af7e6b9ecfc176049d1c208e3d14b))

## [0.1.7](https://github.com/project-n-oss/sidekick/compare/v0.1.6...v0.1.7) (2023-04-12)


### Features

* docker registry build ([e1d7342](https://github.com/project-n-oss/sidekick/commit/e1d73420477f096e59d2d04e677bf14c6631215a))

## [0.1.6](https://github.com/project-n-oss/sidekick/compare/v0.1.5...v0.1.6) (2023-04-11)


### Features

* version cmd ([fa8b663](https://github.com/project-n-oss/sidekick/commit/fa8b6635746cd75ae129cab1604280f29ab5720e))

## [0.1.5](https://github.com/project-n-oss/sidekick/compare/v0.1.4...v0.1.5) (2023-04-11)


### Bug Fixes

* added standard http client for quicksilver ([612110b](https://github.com/project-n-oss/sidekick/commit/612110be01608f8a8e60dd73eee7536b4033160e))
* overwrite bin in release ([97e28cd](https://github.com/project-n-oss/sidekick/commit/97e28cdd22a8606fe1b0822bae72855f34232e37))

## [0.1.4](https://github.com/project-n-oss/sidekick/compare/v0.1.3...v0.1.4) (2023-04-07)


### Features

* added load tests ([#11](https://github.com/project-n-oss/sidekick/issues/11)) ([557531e](https://github.com/project-n-oss/sidekick/commit/557531e05214fc1c32782da41d8e3807d0e5a209))

## [0.1.3](https://github.com/project-n-oss/sidekick/compare/v0.1.2...v0.1.3) (2023-04-05)


### Bug Fixes

* rollback to stylepath ([4398f44](https://github.com/project-n-oss/sidekick/commit/4398f447ece0230f53a20f77d281b71a2838f579))

## [0.1.2](https://github.com/project-n-oss/sidekick/compare/v0.1.1...v0.1.2) (2023-04-04)


### Features

* failover tests and non path-style requests ([#8](https://github.com/project-n-oss/sidekick/issues/8)) ([48910dd](https://github.com/project-n-oss/sidekick/commit/48910dd06d29e0b9aa9ca1121516c2672e8afcf2))

## [0.1.1](https://github.com/project-n-oss/sidekick/compare/v0.1.0...v0.1.1) (2023-03-27)


### Features

* updated documentation and sidekick_service_init for databricks ([#6](https://github.com/project-n-oss/sidekick/issues/6)) ([83f8595](https://github.com/project-n-oss/sidekick/commit/83f8595aa633a9864c572c02380abee3345ea049))

## 0.1.0 (2023-03-27)


### Miscellaneous Chores

* release 0.1.0 ([5c4fad4](https://github.com/project-n-oss/sidekick/commit/5c4fad4f81fc62f48080c515dc84441026527540))
