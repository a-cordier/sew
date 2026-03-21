<p align="center"><img src="logo.svg" alt="sew" width="140"/><br><em>Your local Kubernetes tailor shop</em></p>

<p align="center">
<a href="https://github.com/a-cordier/sew/actions/workflows/lint.yml"><img src="https://github.com/a-cordier/sew/actions/workflows/lint.yml/badge.svg" alt="CI"></a>
<a href="https://github.com/a-cordier/sew/releases/latest"><img src="https://img.shields.io/github/v/release/a-cordier/sew" alt="Release"></a>
<img src="https://img.shields.io/github/go-mod/go-version/a-cordier/sew" alt="Go version">
<img src="https://img.shields.io/github/license/a-cordier/sew" alt="License">
<a href="https://goreportcard.com/report/github.com/a-cordier/sew"><img src="https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat" alt="Go Report Card"></a>
</p>

<p align="center">
<a href="https://a-cordier.github.io/sew/docs/">Documentation</a> · <a href="https://a-cordier.github.io/sew/registry/">Registry</a> · <a href="#quick-start">Getting Started</a>
</p>

## Quick start

Install:

```bash
go install github.com/a-cordier/sew@latest
```

Create a cluster from a registry context:

```bash
sew create --registry https://a-cordier.github.io/sew --from gravitee.io/apim
```

Tear it down:

```bash
sew delete
```

## Features

- **Discoverable** — Browse available contexts in the [registry](https://a-cordier.github.io/sew/registry/) and deploy them with a single command.
- **Composable** — Contexts build on each other via `from`. Mix databases, brokers, and applications into a tailored stack without duplicating configuration.
- **Convenient** — Built-in image mirrors, preloading, local DNS, and readiness ordering get a full environment running with minimal effort.
- **Agent-friendly** — Machine-readable schema, structured registry, and `AGENTS.md` make sew a first-class target for AI-assisted workflows.

## Prerequisites

- **Go 1.25+**
- **Docker**

## Contributing

### Developers

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide. Key commands:

```bash
task lint          # run Go linter
task test          # run all tests
task fmt:yaml      # format YAML files
task site:build    # rebuild the doc site after registry/ or site/ changes
```

### Context maintainers

Contexts live under `registry/` following the `org/product/variant` convention. Each context has a `sew.yaml` describing Helm repos, components, and features. Refer to the [Context Format](https://a-cordier.github.io/sew/docs/reference/context-format/) and [AI Toolchain](https://a-cordier.github.io/sew/docs/guides/ai-toolchain/) docs for authoring guidelines.

## License

[Apache License 2.0](LICENSE)
