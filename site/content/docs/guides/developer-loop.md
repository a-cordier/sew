---
title: "Developer Loop"
weight: 5
type: docs
---

When you're iterating on application code, the last thing you want is a slow feedback loop. sew's `builds` feature lets you compile, build a Docker image, push it to the cluster, and restart workloads -- all in a single command, without leaving your terminal.

## Defining builds

Add a `builds` section to your `sew.yaml`. Each entry describes one image you build locally:

```yaml
builds:
  - name: gateway
    image: graviteeio/apim-gateway:latest-debian
    dir: $HOME/src/gravitee/gravitee-api-management
    pre:
      - mvn clean install -DskipTests -T 2C -pl gateway-standalone -am
    context: gateway-standalone/target
    dockerfile: gateway/docker/Dockerfile

  - name: console-ui
    image: graviteeio/apim-management-ui:latest
    dir: $HOME/src/gravitee/gravitee-api-management
    pre:
      - npx nx build console
    context: gravitee-apim-console-webui
    dockerfile: gravitee-apim-console-webui/docker/Dockerfile
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Short identifier, used to select builds on the CLI |
| `image` | yes | Docker image tag to build (e.g. `myapp:latest`) |
| `dir` | no | Working directory for `pre` commands and base for relative paths. Supports env vars (`$HOME`). Defaults to `.` |
| `pre` | no | Shell commands run sequentially before `docker build` (compilation, packaging, etc.) |
| `context` | no | Docker build context, relative to `dir`. Defaults to `.` |
| `dockerfile` | no | Path to the Dockerfile, relative to `dir`. Defaults to `Dockerfile` in the context |

## Running builds

```bash
sew build
```

This builds every entry, pushes each image to the cluster's preload registry, and restarts any Deployment or StatefulSet that references the image. Build output (pre-build commands, docker build logs) is captured in `~/.sew/logs/build/build.log` and the terminal shows a clean progress view.

### Building a subset

Pass one or more names to build only what you need:

```bash
sew build gateway
sew build gateway console-ui
```

### Skipping pre-build commands

When you've already compiled locally and just want to rebuild the Docker image:

```bash
sew build --skip-pre gateway
```

### Building without restarting

Push the image to the registry but don't trigger a rollout restart:

```bash
sew build --no-restart gateway
```

## One command from zero to dev loop

If you don't have a cluster yet, `--create` creates one before building. When the cluster already exists, the flag is silently ignored:

```bash
sew build --create
```

This runs the full `sew create` flow (preload, cluster, component install) then proceeds with the builds. You can pass context flags too -- they're forwarded to the creation step:

```bash
sew build --create --no-es --skip-pre gateway
```

This is the fastest way to go from a clean machine to running your local code on a cluster.

## How builds interact with preloading

When `builds` is configured, sew automatically excludes build images from the preload list during `sew create`. If your context preloads `graviteeio/apim-gateway:latest-debian` and you define a build for the same image, sew won't waste time pulling it from a remote registry -- it knows you'll push a local version.

This works transparently with both merge and replace preload modes. You don't need to manually add `skip` entries for your build images.

## Flags reference

| Flag | Description |
|------|-------------|
| `--create` | Create the cluster if it doesn't exist |
| `--skip-pre` | Skip pre-build commands, go straight to `docker build` |
| `--no-restart` | Build and push but don't restart workloads |
| `--name <cluster>` | Target a specific cluster (default: from config) |

## Build logs

All build output is written to `~/.sew/logs/build/build.log`. When a step fails, sew points you to the log file:

```
  ✗ Running pre-build commands (failed after 12.3s)
    See logs: /home/user/.sew/logs/build/build.log
```
