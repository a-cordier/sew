# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

Initial release of sew — a CLI tool for composing and deploying Kubernetes
stacks from a registry of reusable contexts.

### Added

- `sew create` / `sew delete` commands for full lifecycle management of Kind clusters.
- `sew patch` command with diff output and dry-run support.
- `sew status` and `sew dns` commands for cluster introspection.
- Context composition via `from` with deep-merge of Helm values and named lists.
- Remote registry support — fetch and resolve contexts from any HTTP registry.
- Abstract contexts (`abstract: true`) for shared base configurations.
- `.default` files for automatic variant resolution from partial paths.
- Helm chart installation with inline `values`, `valueFiles`, and repo management.
- Kubernetes manifest support (inline and remote) with namespace handling.
- K8s secrets from local files and environment variables (`onMissing: ignore` for optional resources).
- Component dependency ordering with readiness conditions and timeouts.
- Image preloading into Kind nodes for faster startup.
- Docker layer caching with image mirrors.
- Local DNS resolution with wildcard domain support.
- Cloud provider (Kind) integration with load balancer routing.
- `notes.create` templates for post-create connection info.
- Machine-readable JSON Schema for `sew.yaml` validation.
- `AGENTS.md` and AI tooling for agent-assisted context authoring.
- Documentation site built with Hugo, auto-generated from the registry.
- CI pipeline with Go linting (revive) and tests (gotestsum).
- Registry contexts for Gravitee APIM (OSS all-in-one, OSS Kubernetes, EE Kafka), Kafka, MongoDB, PostgreSQL, and Elasticsearch.
