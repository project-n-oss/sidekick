# Changelog

## [0.2.3](https://github.com/project-n-oss/sidekick/compare/v0.2.2...v0.2.3) (2023-11-20)


### Features

* add 404 failover behavior control via cli arg ([#93](https://github.com/project-n-oss/sidekick/issues/93)) ([40d9ac0](https://github.com/project-n-oss/sidekick/commit/40d9ac02887da7c8f373172449b6384bc0df0e6e))

## [0.2.2](https://github.com/project-n-oss/sidekick/compare/v0.2.1...v0.2.2) (2023-11-16)


### Features

* gcp object path fixes & dataproc integration docs ([#91](https://github.com/project-n-oss/sidekick/issues/91)) ([219ba99](https://github.com/project-n-oss/sidekick/commit/219ba99bbf4f4b8353821c844dfa1e7fee8513ad))

## [0.2.1](https://github.com/project-n-oss/sidekick/compare/v0.2.0...v0.2.1) (2023-09-20)


### Bug Fixes

* make staticcheck actually work in ci and fix warnings ([#87](https://github.com/project-n-oss/sidekick/issues/87)) ([4d17c53](https://github.com/project-n-oss/sidekick/commit/4d17c53a1073d1767c89255cb029d4db90ec3b25))

## [0.2.0](https://github.com/project-n-oss/sidekick/compare/v0.1.35...v0.2.0) (2023-09-20)


### ⚠ BREAKING CHANGES

* **KRY-644:** GCP support ([#83](https://github.com/project-n-oss/sidekick/issues/83))

### Features

* **KRY-644:** GCP support ([#83](https://github.com/project-n-oss/sidekick/issues/83)) ([60553fe](https://github.com/project-n-oss/sidekick/commit/60553fe311a57d04e0bbe33edf88926711822088))

## [0.1.35](https://github.com/project-n-oss/sidekick/compare/v0.1.34...v0.1.35) (2023-09-12)


### Features

* Change logging level from Info to Debug for reduced console noise ([#85](https://github.com/project-n-oss/sidekick/issues/85)) ([14bed8a](https://github.com/project-n-oss/sidekick/commit/14bed8afa07b769125a276f3206e357a981b9a38))
* Update Sidekick logo with Granica ([#84](https://github.com/project-n-oss/sidekick/issues/84)) ([6bd3fcb](https://github.com/project-n-oss/sidekick/commit/6bd3fcb8651aba8a756ad48a7784651ed271ae32))

## [0.1.34](https://github.com/project-n-oss/sidekick/compare/v0.1.33...v0.1.34) (2023-08-23)


### Features

* Added a separate config flag to control 404 failover ([16e02b6](https://github.com/project-n-oss/sidekick/commit/16e02b603f4814f058a9bfc43fcb839c1677b1c2))
* Enhance Initial Request Target Selection ([#78](https://github.com/project-n-oss/sidekick/issues/78)) ([dbbe3c4](https://github.com/project-n-oss/sidekick/commit/dbbe3c41e6640db2dd482101995fe9d2ada269d9))
* set failover to default false. Update readme ([#82](https://github.com/project-n-oss/sidekick/issues/82)) ([c582e88](https://github.com/project-n-oss/sidekick/commit/c582e8807a3e15fc6e5c9a444d065096cd1ed31b))


### Bug Fixes

* Avoid dumping resp in logger.Debug calls ([b9956b9](https://github.com/project-n-oss/sidekick/commit/b9956b958b67c7526903bd5bf23a16841b812d50))
* Convert metadata headers starting with "x-amz-meta" to lower case ([160a49a](https://github.com/project-n-oss/sidekick/commit/160a49a45e26a7c6d7eab186371f647ea493f731))
* Dump endpoint used for the request on completion ([e2c26e0](https://github.com/project-n-oss/sidekick/commit/e2c26e0d1f5d16523f743314c99862635b1ca363))
* Fix tests which were passing crunch_traffic_percent as integer instead of string ([512c877](https://github.com/project-n-oss/sidekick/commit/512c87741880b0f04c6fd5de464a957edbd9b6e0))
* Fix the regression caused by local sidekick logic ([#62](https://github.com/project-n-oss/sidekick/issues/62)) ([9d67bbb](https://github.com/project-n-oss/sidekick/commit/9d67bbb1130d2b3ba11841d014d5b2d3ee25755a))
* misc. code and documentation cleanup ([04c4f4c](https://github.com/project-n-oss/sidekick/commit/04c4f4c6e8a29239b77d7661d563450f6c52dc02))
* Prevent routing to offline bolt endpoints ([f7894eb](https://github.com/project-n-oss/sidekick/commit/f7894ebefaad5bd384e7e25c67e1f1f4fae0edf8))
* Switch Quicksilver mock in tests to use httpmock responder ([79333b2](https://github.com/project-n-oss/sidekick/commit/79333b20ba35a3a3c47c7116c2f8efb50f8c70eb))

## [0.1.32](https://github.com/project-n-oss/sidekick/compare/v0.1.31...v0.1.32) (2023-08-16)


### Bug Fixes

* Fix failover typo ([435a71b](https://github.com/project-n-oss/sidekick/commit/435a71b685bf1a005824f9a79c6ab2bdf8f07e85))
* Trying to release again ([#74](https://github.com/project-n-oss/sidekick/issues/74)) ([bf8e1cd](https://github.com/project-n-oss/sidekick/commit/bf8e1cdc43e5336d98ca27c6c20940ed0346b7e3))

## [0.1.31](https://github.com/project-n-oss/sidekick/compare/v0.1.30...v0.1.31) (2023-08-14)


### Features

* catch panics + additional AWS failover behavior ([#70](https://github.com/project-n-oss/sidekick/issues/70)) ([1d8c26e](https://github.com/project-n-oss/sidekick/commit/1d8c26e7834e8aa822160343a2ca3248740696d5))

## [0.1.30](https://github.com/project-n-oss/sidekick/compare/v0.1.29...v0.1.30) (2023-08-11)


### Features

* always log bolt request analytics in debug mode ([#68](https://github.com/project-n-oss/sidekick/issues/68)) ([f8d8ed0](https://github.com/project-n-oss/sidekick/commit/f8d8ed06da2d6b29dea35627af6d8f6fb4ff4a7d))

## [0.1.29](https://github.com/project-n-oss/sidekick/compare/v0.1.28...v0.1.29) (2023-08-09)


### Features

* log analytics in debug mode for each request made by Sidekick ([#66](https://github.com/project-n-oss/sidekick/issues/66)) ([2bfc2fa](https://github.com/project-n-oss/sidekick/commit/2bfc2fa6972ced37cee6f2eeb37cc69721f324d1))

## [0.1.28](https://github.com/project-n-oss/sidekick/compare/v0.1.27...v0.1.28) (2023-07-26)


### Features

* return proper error if boltInfo is empty ([#63](https://github.com/project-n-oss/sidekick/issues/63)) ([83a35cb](https://github.com/project-n-oss/sidekick/commit/83a35cb4352014747d13306e4440f08089944f56))

## [0.1.27](https://github.com/project-n-oss/sidekick/compare/v0.1.26...v0.1.27) (2023-07-26)


### Bug Fixes

* revert "feat(KRY-634): Add sidekick local run to support local bolt endpoint" ([#60](https://github.com/project-n-oss/sidekick/issues/60)) ([9e8e919](https://github.com/project-n-oss/sidekick/commit/9e8e91918f7db86b87e6cb11d20f1691b921a395))

## [0.1.26](https://github.com/project-n-oss/sidekick/compare/v0.1.25...v0.1.26) (2023-07-25)


### Bug Fixes

* letter casing ([#59](https://github.com/project-n-oss/sidekick/issues/59)) ([e45caa7](https://github.com/project-n-oss/sidekick/commit/e45caa7f4f1c3322b562f43e4a50a110f7de09a0))
* trigger releases ([c0107cb](https://github.com/project-n-oss/sidekick/commit/c0107cbfc78be09526d52fa4cb0e5bb0624e56ad))

## [0.1.25](https://github.com/project-n-oss/sidekick/compare/v0.1.24...v0.1.25) (2023-07-25)


### Features

* **KRY-634:** Add sidekick local run to support local bolt endpoint ([#49](https://github.com/project-n-oss/sidekick/issues/49)) ([7bf67c7](https://github.com/project-n-oss/sidekick/commit/7bf67c7568e28de2d2f400bc66c38e8deb172c6e))
* **KRY-636:** Add support for multiple logging levels and clean up logging ([#56](https://github.com/project-n-oss/sidekick/issues/56)) ([28737ad](https://github.com/project-n-oss/sidekick/commit/28737ad5a1660208dde4d74adf58b4496fc9f9b5))


### Bug Fixes

* Add prebuilt binary info ([#54](https://github.com/project-n-oss/sidekick/issues/54)) ([1397abc](https://github.com/project-n-oss/sidekick/commit/1397abcf87cc89074738e4a89eaf2540740c33a5))

## [0.1.25](https://github.com/project-n-oss/sidekick/compare/v0.1.24...v0.1.25) (2023-07-20)


### Features

* **KRY-634:** Add sidekick local run to support local bolt endpoint ([#49](https://github.com/project-n-oss/sidekick/issues/49)) ([7bf67c7](https://github.com/project-n-oss/sidekick/commit/7bf67c7568e28de2d2f400bc66c38e8deb172c6e))


### Bug Fixes

* Add prebuilt binary info ([#54](https://github.com/project-n-oss/sidekick/issues/54)) ([1397abc](https://github.com/project-n-oss/sidekick/commit/1397abcf87cc89074738e4a89eaf2540740c33a5))

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
