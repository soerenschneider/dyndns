# Changelog

## [1.22.0](https://github.com/soerenschneider/dyndns/compare/v1.21.1...v1.22.0) (2026-01-29)


### Features

* include go version metric ([be3de59](https://github.com/soerenschneider/dyndns/commit/be3de59784b64e9acb1c0331f734330a875b8e58))
* initial support for nats ([a53b222](https://github.com/soerenschneider/dyndns/commit/a53b2225a8728bf4c6d537540f6aff17f8181ec3))
* update list of built-in ip resolvers ([2269ef4](https://github.com/soerenschneider/dyndns/commit/2269ef40adab62760e3b56259029bb45091776b2))


### Bug Fixes

* **deps:** Update dependency go to v1.24.4 ([8133b2b](https://github.com/soerenschneider/dyndns/commit/8133b2b8ae3bba9cecf96d06a1c9ac1b0f12f321))
* **deps:** Update module github.com/aws/aws-lambda-go to v1.49.0 ([286c1af](https://github.com/soerenschneider/dyndns/commit/286c1af8be081b6c0eaf548c850b5fcf8e081684))
* **deps:** Update module github.com/aws/aws-sdk-go to v1.55.7 ([6e7654d](https://github.com/soerenschneider/dyndns/commit/6e7654d4f32c79b8b9008bcb919674b77753bdb2))
* **deps:** Update module github.com/caarlos0/env/v6 to v11 ([a4e8bbe](https://github.com/soerenschneider/dyndns/commit/a4e8bbed813e24b3f001a5dbf828a8a6aee48050))
* **deps:** Update module github.com/hashicorp/vault/api to v1.20.0 ([a004fdd](https://github.com/soerenschneider/dyndns/commit/a004fddff06445ed92c5db33099ceec6c2a0a704))
* **deps:** Update module github.com/hashicorp/vault/api/auth/approle to v0.10.0 ([72b8b1a](https://github.com/soerenschneider/dyndns/commit/72b8b1a7d0782c25b7b0e2867a2e4cef56eb58f6))
* **deps:** Update module github.com/nats-io/nats.go to v1.43.0 ([43b808c](https://github.com/soerenschneider/dyndns/commit/43b808cf33b6c9930af8594ce290d50c899341cd))
* **deps:** Update module golang.org/x/term to v0.32.0 ([dd4cf50](https://github.com/soerenschneider/dyndns/commit/dd4cf50ac1d4d51502ecccdfff72d92539d0fd99))
* fix linting issues ([f70711a](https://github.com/soerenschneider/dyndns/commit/f70711a15cdeb6c408c7f9bf3e7c645794a254e8))

## [1.21.1](https://github.com/soerenschneider/dyndns/compare/v1.21.0...v1.21.1) (2024-09-18)


### Bug Fixes

* **deps:** bump github.com/aws/aws-sdk-go from 1.54.20 to 1.55.5 ([3e7d5ee](https://github.com/soerenschneider/dyndns/commit/3e7d5ee1b79d5e6345e546c1736460358e8cd37e))
* **deps:** bump github.com/eclipse/paho.mqtt.golang from 1.4.3 to 1.5.0 ([22a59ba](https://github.com/soerenschneider/dyndns/commit/22a59ba05778a0055f4b7f113bade62e89cbaf3e))
* **deps:** bump github.com/hashicorp/vault/api/auth/approle ([2d7963c](https://github.com/soerenschneider/dyndns/commit/2d7963c753ae012df92b033a585372ec24665c1c))
* **deps:** bump github.com/rs/zerolog from 1.32.0 to 1.33.0 ([3474e37](https://github.com/soerenschneider/dyndns/commit/3474e37e1c463241ee36abe5da43be16342669cb))
* **deps:** bump golang from 1.22.5 to 1.23.1 ([6dcbc90](https://github.com/soerenschneider/dyndns/commit/6dcbc90b009db18ba43a59af17db8dd1df9ec662))
* fix validation for email configuration ([92a43ee](https://github.com/soerenschneider/dyndns/commit/92a43ee9bca448f7073d9e100f68c54437fe3d4c))

## [1.21.0](https://github.com/soerenschneider/dyndns/compare/v1.20.0...v1.21.0) (2024-07-24)


### Features

* allow force sending update request at application start ([80e0f77](https://github.com/soerenschneider/dyndns/commit/80e0f77f71316a1c437ce36578af9b2a4f0fb01d))
* allow Lambda handler be triggered by SQS and API Gateway ([087319a](https://github.com/soerenschneider/dyndns/commit/087319a132a3bff6947b93df46365eae0091c496))


### Bug Fixes

* **deps:** bump github.com/aws/aws-lambda-go from 1.46.0 to 1.47.0 ([196de5a](https://github.com/soerenschneider/dyndns/commit/196de5a4b7b43d0e6e7bff356b9c4764983f0242))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.35 to 1.53.15 ([e55b5f5](https://github.com/soerenschneider/dyndns/commit/e55b5f5e6e67c7fcb7f43fb9983c21d5b1f4daa8))
* **deps:** bump github.com/aws/aws-sdk-go from 1.53.15 to 1.54.20 ([5e2977e](https://github.com/soerenschneider/dyndns/commit/5e2977ed0d2c5871fba0db49aa8d13508369fa94))
* **deps:** bump github.com/go-playground/validator/v10 ([6512b54](https://github.com/soerenschneider/dyndns/commit/6512b54f70442211a871ef034727f911a1e38e9d))
* **deps:** bump github.com/hashicorp/go-retryablehttp ([d3afcee](https://github.com/soerenschneider/dyndns/commit/d3afcee5f0a862437d6394de0c2d6c81b965e6a2))
* **deps:** bump github.com/hashicorp/vault/api from 1.12.0 to 1.14.0 ([57601cf](https://github.com/soerenschneider/dyndns/commit/57601cf1b3d52075e7f0c199b9387e1f427e6123))
* **deps:** bump github.com/prometheus/client_golang ([5150598](https://github.com/soerenschneider/dyndns/commit/51505984ab61cfeea3a4a113e224101e7a64ed9d))
* **deps:** bump golang from 1.22.1 to 1.22.3 ([bbdbee7](https://github.com/soerenschneider/dyndns/commit/bbdbee7d8aac69a3928677865bada55a95d7ed57))
* **deps:** bump golang from 1.22.3 to 1.22.5 ([066d4a7](https://github.com/soerenschneider/dyndns/commit/066d4a7dcf276c35ed7dfe0646641216c79565cf))
* **deps:** bump golang.org/x/term from 0.17.0 to 0.20.0 ([e63a958](https://github.com/soerenschneider/dyndns/commit/e63a95803cecbc1572260996b649c4350f48b438))
* **deps:** bump golang.org/x/term from 0.20.0 to 0.22.0 ([3b59825](https://github.com/soerenschneider/dyndns/commit/3b59825daa62ddb4dba68e3521f083939e754cf6))
* fix syntax error ([21472d2](https://github.com/soerenschneider/dyndns/commit/21472d2e90136d3dc69740c56bbbdbfc634a2bc6))

## [1.20.0](https://github.com/soerenschneider/dyndns/compare/v1.19.0...v1.20.0) (2024-03-17)


### Features

* improve logging ([181fc19](https://github.com/soerenschneider/dyndns/commit/181fc19f9cd6848c91f69b2a90ea54fafd12fab9))

## [1.19.0](https://github.com/soerenschneider/dyndns/compare/v1.18.0...v1.19.0) (2024-03-16)


### Features

* add metrics for sqs api calls ([476c42a](https://github.com/soerenschneider/dyndns/commit/476c42ac015a767ad0bb51ec9bf84e5a70138104))
* allow stopping reconciliation early after first successful update request dispatch ([7dbf766](https://github.com/soerenschneider/dyndns/commit/7dbf76672cfbd3c240f2c6e5d26efaf4353d47a5))


### Bug Fixes

* allow setting explicit aws region ([0fc9f15](https://github.com/soerenschneider/dyndns/commit/0fc9f15ed2506cf6fb310a89dfa04f6d6b2446a2))
* fix config validation ([51012d3](https://github.com/soerenschneider/dyndns/commit/51012d3f3926382d7fe45fc79dcd06b63445a34e))

## [1.18.0](https://github.com/soerenschneider/dyndns/compare/v1.17.0...v1.18.0) (2024-03-15)


### Features

* Add support for AWS SQS ([#384](https://github.com/soerenschneider/dyndns/issues/384)) ([06b9cd7](https://github.com/soerenschneider/dyndns/commit/06b9cd78d8154813af1b73c3c6f6329f2f4de6c8))
* support yaml config files ([87248bb](https://github.com/soerenschneider/dyndns/commit/87248bbcff56a0cb7ce45d5b7c61defe2ab2b21f))


### Bug Fixes

* **deps:** bump github.com/aws/aws-lambda-go from 1.41.0 to 1.43.0 ([#350](https://github.com/soerenschneider/dyndns/issues/350)) ([d0903c4](https://github.com/soerenschneider/dyndns/commit/d0903c49398b08c8a000d2e131199a5fda87dae8))
* **deps:** bump github.com/aws/aws-lambda-go from 1.43.0 to 1.46.0 ([#366](https://github.com/soerenschneider/dyndns/issues/366)) ([e91095f](https://github.com/soerenschneider/dyndns/commit/e91095f84335cdee8e951c5296560636448c9c3b))
* **deps:** bump github.com/aws/aws-sdk-go from 1.48.13 to 1.49.4 ([#347](https://github.com/soerenschneider/dyndns/issues/347)) ([90da6f6](https://github.com/soerenschneider/dyndns/commit/90da6f689d0c1438334662ca6bb9346e2fcd9080))
* **deps:** bump github.com/aws/aws-sdk-go from 1.49.4 to 1.50.15 ([#372](https://github.com/soerenschneider/dyndns/issues/372)) ([d7560f2](https://github.com/soerenschneider/dyndns/commit/d7560f2b9bfab41024cee89d227851bcf018c32c))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.15 to 1.50.35 ([#383](https://github.com/soerenschneider/dyndns/issues/383)) ([a6d4451](https://github.com/soerenschneider/dyndns/commit/a6d4451843f53978af573ef04d2a6067dad41339))
* **deps:** bump github.com/go-playground/validator/v10 ([cb064de](https://github.com/soerenschneider/dyndns/commit/cb064de4100de3a2efaca9033b8631f78b100bbe))
* **deps:** bump github.com/go-playground/validator/v10 ([#371](https://github.com/soerenschneider/dyndns/issues/371)) ([cb3b248](https://github.com/soerenschneider/dyndns/commit/cb3b248b48c46a478d60f128b19f7c81b984c561))
* **deps:** bump github.com/hashicorp/vault/api from 1.10.0 to 1.12.0 ([#370](https://github.com/soerenschneider/dyndns/issues/370)) ([8e0f8df](https://github.com/soerenschneider/dyndns/commit/8e0f8dfd6083f8ded3ee4926e15f2cd04bfc3df4))
* **deps:** bump github.com/hashicorp/vault/api/auth/approle ([#375](https://github.com/soerenschneider/dyndns/issues/375)) ([d3668bc](https://github.com/soerenschneider/dyndns/commit/d3668bcbb8e7dc171f7a2a42cd271786278b67a1))
* **deps:** bump github.com/prometheus/client_golang ([82b5af6](https://github.com/soerenschneider/dyndns/commit/82b5af6333f02584ca45ad14e8ac5a0fb12933df))
* **deps:** bump github.com/prometheus/client_golang ([#353](https://github.com/soerenschneider/dyndns/issues/353)) ([c3b34e1](https://github.com/soerenschneider/dyndns/commit/c3b34e1117b87aa2c08475b9cc3be6ecd9a36ba0))
* **deps:** bump github.com/rs/zerolog from 1.31.0 to 1.32.0 ([#374](https://github.com/soerenschneider/dyndns/issues/374)) ([d9b59f2](https://github.com/soerenschneider/dyndns/commit/d9b59f2961ecbb0cfd8eec07215c649546ea32d0))
* **deps:** bump golang from 1.21.4 to 1.21.5 ([#344](https://github.com/soerenschneider/dyndns/issues/344)) ([163d012](https://github.com/soerenschneider/dyndns/commit/163d012e90d4f1c00c02630cdadced3cad48fecd))
* **deps:** bump golang from 1.21.5 to 1.22.0 ([#368](https://github.com/soerenschneider/dyndns/issues/368)) ([a95378d](https://github.com/soerenschneider/dyndns/commit/a95378dd06fe1ee2942f6f8268d93cb9773c25fe))
* **deps:** bump golang from 1.22.0 to 1.22.1 ([#381](https://github.com/soerenschneider/dyndns/issues/381)) ([2501216](https://github.com/soerenschneider/dyndns/commit/2501216b5dd1255010b405ed5d879057807bb999))

## [1.18.0](https://github.com/soerenschneider/dyndns/compare/v1.17.0...v1.18.0) (2024-03-15)


### Features

* Add support for AWS SQS ([#384](https://github.com/soerenschneider/dyndns/issues/384)) ([06b9cd7](https://github.com/soerenschneider/dyndns/commit/06b9cd78d8154813af1b73c3c6f6329f2f4de6c8))
* support yaml config files ([87248bb](https://github.com/soerenschneider/dyndns/commit/87248bbcff56a0cb7ce45d5b7c61defe2ab2b21f))


### Bug Fixes

* **deps:** bump github.com/aws/aws-lambda-go from 1.41.0 to 1.43.0 ([#350](https://github.com/soerenschneider/dyndns/issues/350)) ([d0903c4](https://github.com/soerenschneider/dyndns/commit/d0903c49398b08c8a000d2e131199a5fda87dae8))
* **deps:** bump github.com/aws/aws-lambda-go from 1.43.0 to 1.46.0 ([#366](https://github.com/soerenschneider/dyndns/issues/366)) ([e91095f](https://github.com/soerenschneider/dyndns/commit/e91095f84335cdee8e951c5296560636448c9c3b))
* **deps:** bump github.com/aws/aws-sdk-go from 1.48.13 to 1.49.4 ([#347](https://github.com/soerenschneider/dyndns/issues/347)) ([90da6f6](https://github.com/soerenschneider/dyndns/commit/90da6f689d0c1438334662ca6bb9346e2fcd9080))
* **deps:** bump github.com/aws/aws-sdk-go from 1.49.4 to 1.50.15 ([#372](https://github.com/soerenschneider/dyndns/issues/372)) ([d7560f2](https://github.com/soerenschneider/dyndns/commit/d7560f2b9bfab41024cee89d227851bcf018c32c))
* **deps:** bump github.com/aws/aws-sdk-go from 1.50.15 to 1.50.35 ([#383](https://github.com/soerenschneider/dyndns/issues/383)) ([a6d4451](https://github.com/soerenschneider/dyndns/commit/a6d4451843f53978af573ef04d2a6067dad41339))
* **deps:** bump github.com/go-playground/validator/v10 ([cb064de](https://github.com/soerenschneider/dyndns/commit/cb064de4100de3a2efaca9033b8631f78b100bbe))
* **deps:** bump github.com/go-playground/validator/v10 ([#371](https://github.com/soerenschneider/dyndns/issues/371)) ([cb3b248](https://github.com/soerenschneider/dyndns/commit/cb3b248b48c46a478d60f128b19f7c81b984c561))
* **deps:** bump github.com/hashicorp/vault/api from 1.10.0 to 1.12.0 ([#370](https://github.com/soerenschneider/dyndns/issues/370)) ([8e0f8df](https://github.com/soerenschneider/dyndns/commit/8e0f8dfd6083f8ded3ee4926e15f2cd04bfc3df4))
* **deps:** bump github.com/hashicorp/vault/api/auth/approle ([#375](https://github.com/soerenschneider/dyndns/issues/375)) ([d3668bc](https://github.com/soerenschneider/dyndns/commit/d3668bcbb8e7dc171f7a2a42cd271786278b67a1))
* **deps:** bump github.com/prometheus/client_golang ([#353](https://github.com/soerenschneider/dyndns/issues/353)) ([c3b34e1](https://github.com/soerenschneider/dyndns/commit/c3b34e1117b87aa2c08475b9cc3be6ecd9a36ba0))
* **deps:** bump github.com/rs/zerolog from 1.31.0 to 1.32.0 ([#374](https://github.com/soerenschneider/dyndns/issues/374)) ([d9b59f2](https://github.com/soerenschneider/dyndns/commit/d9b59f2961ecbb0cfd8eec07215c649546ea32d0))
* **deps:** bump golang from 1.21.4 to 1.21.5 ([#344](https://github.com/soerenschneider/dyndns/issues/344)) ([163d012](https://github.com/soerenschneider/dyndns/commit/163d012e90d4f1c00c02630cdadced3cad48fecd))
* **deps:** bump golang from 1.21.5 to 1.22.0 ([#368](https://github.com/soerenschneider/dyndns/issues/368)) ([a95378d](https://github.com/soerenschneider/dyndns/commit/a95378dd06fe1ee2942f6f8268d93cb9773c25fe))
* **deps:** bump golang from 1.22.0 to 1.22.1 ([#381](https://github.com/soerenschneider/dyndns/issues/381)) ([2501216](https://github.com/soerenschneider/dyndns/commit/2501216b5dd1255010b405ed5d879057807bb999))

## [1.17.0](https://github.com/soerenschneider/dyndns/compare/v1.16.0...v1.17.0) (2023-12-24)


### Features

* increase resilience by isolating brokers ([0ab3ae2](https://github.com/soerenschneider/dyndns/commit/0ab3ae20bff92c97094c58cd0943068e0e7765d2))
* suppor for parsing http dispatcher conf via env variables ([230c6de](https://github.com/soerenschneider/dyndns/commit/230c6dea1f297dc1751dddfb76dc5d5636369ac1))
* support http client- and server ([3862498](https://github.com/soerenschneider/dyndns/commit/38624987ab1a5c669edd442f0174866c45ed45c9))


### Bug Fixes

* **deps:** bump github.com/aws/aws-sdk-go from 1.45.21 to 1.45.24 ([#322](https://github.com/soerenschneider/dyndns/issues/322)) ([22d85e4](https://github.com/soerenschneider/dyndns/commit/22d85e45376b32f2bc42aa4ea16693cdb74790f1))
* **deps:** bump github.com/aws/aws-sdk-go from 1.45.24 to 1.45.25 ([#326](https://github.com/soerenschneider/dyndns/issues/326)) ([157c9cf](https://github.com/soerenschneider/dyndns/commit/157c9cf1bbfd66b2888e50c0cbee80623de21c6c))
* **deps:** bump github.com/aws/aws-sdk-go from 1.45.25 to 1.48.0 ([#336](https://github.com/soerenschneider/dyndns/issues/336)) ([5a286e9](https://github.com/soerenschneider/dyndns/commit/5a286e972814cdf910ebd8974bb534df0b55765a))
* **deps:** bump github.com/aws/aws-sdk-go from 1.48.0 to 1.48.13 ([#341](https://github.com/soerenschneider/dyndns/issues/341)) ([1d52db1](https://github.com/soerenschneider/dyndns/commit/1d52db14bc72d29ce0729e9a1bbcfb1a58467024))
* **deps:** bump github.com/go-playground/validator/v10 ([#329](https://github.com/soerenschneider/dyndns/issues/329)) ([f130c9f](https://github.com/soerenschneider/dyndns/commit/f130c9f3175ef0cd40a156582510e2fd3d12d56f))
* **deps:** bump github.com/hashicorp/go-retryablehttp ([#332](https://github.com/soerenschneider/dyndns/issues/332)) ([392057c](https://github.com/soerenschneider/dyndns/commit/392057c6ca2a1b30c825174937381353da10648d))
* **deps:** bump golang from 1.21.1 to 1.21.2 ([#323](https://github.com/soerenschneider/dyndns/issues/323)) ([ce28f8a](https://github.com/soerenschneider/dyndns/commit/ce28f8a055803a49a7613dc196bf23de7da3e1dc))
* **deps:** bump golang from 1.21.2 to 1.21.3 ([#325](https://github.com/soerenschneider/dyndns/issues/325)) ([0d804f6](https://github.com/soerenschneider/dyndns/commit/0d804f6ce50802d09f4e4485dc293dd2308cfbc7))
* **deps:** bump golang from 1.21.3 to 1.21.4 ([#334](https://github.com/soerenschneider/dyndns/issues/334)) ([2734d61](https://github.com/soerenschneider/dyndns/commit/2734d61d3fe313906ec432b33b11e49b4190238e))
* fix panic due to calling the wrong unlock method ([f537adc](https://github.com/soerenschneider/dyndns/commit/f537adc31e1d2fae978e405f3876aadd1a1d5c1d))

## [1.16.0](https://github.com/soerenschneider/dyndns/compare/v1.15.1...v1.16.0) (2023-10-04)


### Features

* client interacts with server running on Lambda ([858b1e2](https://github.com/soerenschneider/dyndns/commit/858b1e2fc53d9637763c7f1fd6225e2019bee84e))
* detect drift on dns records ([709caf6](https://github.com/soerenschneider/dyndns/commit/709caf665c08faf2d0d8c2245b5fe98188fa0c31))
* make server run on AWS lambda ([bfe821f](https://github.com/soerenschneider/dyndns/commit/bfe821f58ec27a17aab23991cfb6f4e9c1575cfd))


### Bug Fixes

* **deps:** Bump github.com/aws/aws-sdk-go from 1.45.15 to 1.45.16 ([#310](https://github.com/soerenschneider/dyndns/issues/310)) ([a54e0de](https://github.com/soerenschneider/dyndns/commit/a54e0de9e2847a348bc9012d334103310e3362e3))
* **deps:** bump github.com/aws/aws-sdk-go from 1.45.16 to 1.45.21 ([#321](https://github.com/soerenschneider/dyndns/issues/321)) ([9e22e47](https://github.com/soerenschneider/dyndns/commit/9e22e4713dd875a4b5a516232fc10b8cc8df0f16))
* **deps:** bump github.com/go-playground/validator/v10 ([#318](https://github.com/soerenschneider/dyndns/issues/318)) ([3a94897](https://github.com/soerenschneider/dyndns/commit/3a9489710dd5ada2c16995aa66a709692e843217))
* **deps:** bump github.com/prometheus/client_golang ([#319](https://github.com/soerenschneider/dyndns/issues/319)) ([5cf1dd4](https://github.com/soerenschneider/dyndns/commit/5cf1dd4cc086dcfd34feeca4e3a5927f65c457b4))

## [1.15.1](https://github.com/soerenschneider/dyndns/compare/v1.15.0...v1.15.1) (2023-07-07)


### Bug Fixes

* fix validate tag ([def1897](https://github.com/soerenschneider/dyndns/commit/def1897d100475fb7cba4353bde74013c75e3f0c))

## [1.15.0](https://github.com/soerenschneider/dyndns/compare/v1.14.3...v1.15.0) (2023-07-07)


### Features

* allow setting keypair via env variables / config file ([6fefcc5](https://github.com/soerenschneider/dyndns/commit/6fefcc50d1afee60ed8e420be1e1f2f7b0127274))
* parse server config via env vars ([63248c0](https://github.com/soerenschneider/dyndns/commit/63248c0187b092e8aa397ed75a8e280f05d9f4a1))


### Bug Fixes

* forgot to call ParseEnvVariables ([6d744e8](https://github.com/soerenschneider/dyndns/commit/6d744e8bd79aa4a684c718efb24ccf01d6f7c2b2))
* more precise log message ([9c68b32](https://github.com/soerenschneider/dyndns/commit/9c68b327cc8c2748e80c19d21a3e638b4cc7851f))
* return default config when no file is given ([cefe358](https://github.com/soerenschneider/dyndns/commit/cefe358b029f2bddcc4c19e09e339628f764b973))
* return error only when specified custom config path ([f22a814](https://github.com/soerenschneider/dyndns/commit/f22a8145bc6ceb38175ad7a50ba2f9df77cb606f))

## [1.14.3](https://github.com/soerenschneider/dyndns/compare/v1.14.2...v1.14.3) (2023-07-06)


### Bug Fixes

* add missing mux handler to server ([909e46e](https://github.com/soerenschneider/dyndns/commit/909e46eff71f0969356a2448cdb35f2b973d5d8b))
* encode public key for log output ([c0b1b77](https://github.com/soerenschneider/dyndns/commit/c0b1b77bb413ed170cc8bff451ad7f28c3581c4c))
* use common prefix for env var ([66073b1](https://github.com/soerenschneider/dyndns/commit/66073b1f135957f60619689ecb3de6af1f31388f))

## [1.14.2](https://github.com/soerenschneider/dyndns/compare/v1.14.1...v1.14.2) (2023-07-05)


### Miscellaneous Chores

* release 1.14.2 ([d098ace](https://github.com/soerenschneider/dyndns/commit/d098ace3157d45a720c7bb129271537e02ff671a))

## [1.14.1](https://github.com/soerenschneider/dyndns/compare/v1.14.0...v1.14.1) (2023-07-05)


### Bug Fixes

* fix import ([cb10801](https://github.com/soerenschneider/dyndns/commit/cb108016b81cafea65703cc554eb3fcbca9053b9))

## [1.14.0](https://github.com/soerenschneider/dyndns/compare/v1.13.0...v1.14.0) (2023-07-05)


### Features

* add easy way to generate keypairs from the cli ([5632aa8](https://github.com/soerenschneider/dyndns/commit/5632aa842decaf19e35203535aa760bcdbf49fe5))

## [1.13.0](https://github.com/soerenschneider/dyndns/compare/v1.12.1...v1.13.0) (2023-07-04)


### Features

* use dedicated validator ([3f67832](https://github.com/soerenschneider/dyndns/commit/3f67832eba6c53ebce49e2fbb3b17df7eb9a9f71))


### Bug Fixes

* remove dead code and actually set metric ([1b08613](https://github.com/soerenschneider/dyndns/commit/1b086138b7e0d95a4fd4133d73c5703ca791430d))

## [1.12.1](https://github.com/soerenschneider/dyndns/compare/v1.12.0...v1.12.1) (2023-07-04)


### Bug Fixes

* do not print sensitive values ([dcf30a7](https://github.com/soerenschneider/dyndns/commit/dcf30a78cbabed9328c90291540531d1305637d3))
* fix printing config values ([33cd42f](https://github.com/soerenschneider/dyndns/commit/33cd42fb22ed49301d340551c7dcc38627eb39b1))
* retry reconnects ([aeb518b](https://github.com/soerenschneider/dyndns/commit/aeb518b02f0cf4c2a2826aaafa0ebc5aa553660a))

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
