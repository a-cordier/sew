# Changelog

## [0.4.0](https://github.com/a-cordier/sew/compare/v0.3.1...v0.4.0) (2026-03-23)


### Features

* add sew validate registry command ([7381dff](https://github.com/a-cordier/sew/commit/7381dffd00c275d0eaf5ca59c169f5656089da07))
* introduce context flags for maintainers ([2c9b1de](https://github.com/a-cordier/sew/commit/2c9b1de32a5a1bc858c71733e5bd2229c5d388f4))

## [0.3.1](https://github.com/a-cordier/sew/compare/v0.3.0...v0.3.1) (2026-03-21)


### Documentation

* fix broken link in registry section ([6041cf4](https://github.com/a-cordier/sew/commit/6041cf42ac4c1d99e282e07198822bcaf553b10e))
* fix links in maintainer section of readme ([be90647](https://github.com/a-cordier/sew/commit/be90647598dc89039111604d541952e1b9af1d05))
* make gravitee context stick on top in registry ([d462531](https://github.com/a-cordier/sew/commit/d462531b26e86ffbed2c8e29f69c6b4150ad5292))

## [0.3.0](https://github.com/a-cordier/sew/compare/v0.2.0...v0.3.0) (2026-03-21)


### Features

* support anchor links in doc site ([002c0f6](https://github.com/a-cordier/sew/commit/002c0f65ab63cb055f5c8a8381af9d6e5d7bafdf))

## [0.2.0](https://github.com/a-cordier/sew/compare/v0.1.0...v0.2.0) (2026-03-21)


### Features

* add ai tooling for context maintainers ([aad0dfa](https://github.com/a-cordier/sew/commit/aad0dfa6a2bb7cb741e928e1fe9f87d9c5a9e095))
* add cloud provider kind integration ([e2bb595](https://github.com/a-cordier/sew/commit/e2bb59597b2e01e73b7f81c24434898a5a4b697c))
* add component level readiness ([fc785c8](https://github.com/a-cordier/sew/commit/fc785c8da157aa02b55175f7067a9bbce686f08c))
* add create and delete notes for end users ([d3b2929](https://github.com/a-cordier/sew/commit/d3b292977a87c75c1821948d2fd57187689d8bb8))
* add default feature for registries ([1d2ccda](https://github.com/a-cordier/sew/commit/1d2ccda67d81fdeda3b3b0f05c5f432d5aabe979))
* add dry run to patch command ([3f62096](https://github.com/a-cordier/sew/commit/3f620961739125fe9be73d209c527c0c2aa44eae))
* add gateway and  dns resolution ([241afc3](https://github.com/a-cordier/sew/commit/241afc3859f18fe9a82afb8ded0428d0a61a8108))
* add gravitee apim aio context ([c76b287](https://github.com/a-cordier/sew/commit/c76b2872aa9333f4554d9b7465b91f1c08f22f3c))
* add Gravitee APIM enterprise Kafka context ([3a08db6](https://github.com/a-cordier/sew/commit/3a08db667ed914c813cc4021a5ba10e462460fb9))
* add kafka registry ([9e48e4b](https://github.com/a-cordier/sew/commit/9e48e4b9371cc0eef5778dcb81bb9fa7d2fb820c))
* add license volume mounts to ee/kafka gateway and api ([903924f](https://github.com/a-cordier/sew/commit/903924f39732d26e179d1ecc85f7ea60cf77a4b9))
* add notes.create with connection info for all contexts ([a2bd8d6](https://github.com/a-cordier/sew/commit/a2bd8d646feb52dbac1700dc10ca179e29971362))
* add patch command ([773c7ee](https://github.com/a-cordier/sew/commit/773c7ee171bfb59c83e97cb0dc206969ce304756))
* add pg variant for apim aio ([0890681](https://github.com/a-cordier/sew/commit/08906811aa5d9c9382b0b97d6569d022845e22a5))
* address component dependencies ([87cd00a](https://github.com/a-cordier/sew/commit/87cd00aabaef3dc42098e644e9ceab7c9866040b))
* allow context maintainers to add their logo ([f02300d](https://github.com/a-cordier/sew/commit/f02300d420905b9f64587d5d6a626fb69558695f))
* allow end user to add components ([d819ddf](https://github.com/a-cordier/sew/commit/d819ddf2adfd3f5ecdffd620ae653388df70d72f))
* allow multiple contexts composition ([cf54ad6](https://github.com/a-cordier/sew/commit/cf54ad6730133770db1691149689001ac2951c04))
* allow to create k8s secrets from local files and env ([bfa7b02](https://github.com/a-cordier/sew/commit/bfa7b02e939430cec97da2c5967c121e73af4180))
* allow to define k8s manifests inline ([aba8480](https://github.com/a-cordier/sew/commit/aba84808675b94204cea6a1dff621a2b9ed4cfae))
* boostrap helm installer implementation ([68ba282](https://github.com/a-cordier/sew/commit/68ba282e81edfdd635ff1b426556e6edce6dd95c))
* bring context composition ([0fec6c7](https://github.com/a-cordier/sew/commit/0fec6c781eb5cb81ea3deeaa7bb724c22e878ec6))
* define deps to user define components ([32dec8e](https://github.com/a-cordier/sew/commit/32dec8e2151d044a1fc60861767c75b4c4c8cd7d))
* enable image pre-loading for patch command ([57fbdc8](https://github.com/a-cordier/sew/commit/57fbdc8f7223432c3dcab3aaee1767eb7326ccb4))
* expose standalone contexts via NodePort for host access ([4fb14ac](https://github.com/a-cordier/sew/commit/4fb14ac60b7630d46715e1cb473aa302f301a78b))
* introduce abstract contexts for maintainers ([066ec3c](https://github.com/a-cordier/sew/commit/066ec3c2c107258fc69e5f985f7f30012de60401))
* leverage docker layer caching with mirrors ([25e9b60](https://github.com/a-cordier/sew/commit/25e9b60788587634e11454973abeed71fd4a5501))
* leverage docker layer caching with preloading ([1dfbe11](https://github.com/a-cordier/sew/commit/1dfbe11ab9b79c51df774ec6ab5a086d1efc72aa))
* merge extra port mapping when composing contexts ([1e8896a](https://github.com/a-cordier/sew/commit/1e8896a41212e5b2062c1840313a5ad4f8dd8379))
* output diff when running dry run for patch ([7aa1e0a](https://github.com/a-cordier/sew/commit/7aa1e0af580146d22ca92c6b8b77a993006d6e17))
* support wild card domain for DNS ([a96529a](https://github.com/a-cordier/sew/commit/a96529a332924a4bbdc1238575eadb63ccfb3317))


### Bug Fixes

* add elastic image to apim aio preloading ([d4e3416](https://github.com/a-cordier/sew/commit/d4e3416bc235b9f98771043b3d15d2a9ebc767d3))
* expand env vars in fromFile paths before absolute path check ([ce71449](https://github.com/a-cordier/sew/commit/ce714499ab5a4d3f03bf49615d0b24c5f4fab270))
* handle ns create in manifest installer ([1b591ac](https://github.com/a-cordier/sew/commit/1b591ac04229dcc06da6b4f918847cebc8f9ec31))
* improve CPK lifecycle and DNS introspection ([fc25c6f](https://github.com/a-cordier/sew/commit/fc25c6f81fa2e436e73e1a61ceb633282e2e9628))
* install helm repos for user defined components ([ba10606](https://github.com/a-cordier/sew/commit/ba106065da21255ced6b66e69e208577c4063051))
* make all commands context aware ([66eaeaf](https://github.com/a-cordier/sew/commit/66eaeaf7417b69e3505d859a10cee83ab29b1311))
* make lb routing and dns resolution consistent across runs ([1f3fbc9](https://github.com/a-cordier/sew/commit/1f3fbc9440d95e3df2bde736f47fcb060b487a85))
* merge named lists by name in deepMergeValues ([33dc3c9](https://github.com/a-cordier/sew/commit/33dc3c9a1acb1633864af63d48a21f9acd095de1))
* override standalone services to ClusterIP in parent contexts ([057a3a3](https://github.com/a-cordier/sew/commit/057a3a3f29ad21a3b961619c014e969ae8a15f21))
* remove images from aio apim values ([a36c591](https://github.com/a-cordier/sew/commit/a36c59164dc34890a1e3e769a27aeb201e649b52))
* resolve manifests from http ([0658caf](https://github.com/a-cordier/sew/commit/0658caff87a212b569d64289887c1158ac280906))
* save cluster state early so delete works after failed installs ([fd919ae](https://github.com/a-cordier/sew/commit/fd919aec4d0f423680999c613b17e20abec2299b))
* skip sudo prompt when no CPK process is running ([536f780](https://github.com/a-cordier/sew/commit/536f780bde3c8f2371aefafff60abd96b6788740))
* slice reallocated when merging components ([047edab](https://github.com/a-cordier/sew/commit/047edab2bb3f8d43150b78e38c92a560fa2565f1))
* strip root sew.yaml to prevent feature leaking ([11876d3](https://github.com/a-cordier/sew/commit/11876d3704090670ff09aae8178842a56e3f432a))


### Performance

* add mirrors for gravitee.io/apim/aio ([d781229](https://github.com/a-cordier/sew/commit/d7812293de18a29cb4758e7201fcfdd5cd0a66fb))
* remove taint from kind control plane ([347fab4](https://github.com/a-cordier/sew/commit/347fab4a6761dde2485e64dcf987368bd577b0c2))
