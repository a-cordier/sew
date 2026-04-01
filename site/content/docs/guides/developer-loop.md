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
  - name: aes
    image: docker.io/datawire/aes:3.12.7
    dir: $HOME/src/gravitee/edge-stack/apro
    pre:
      - make -C $HOME/src/gravitee/edge-stack/emissary docker/emissary.docker.tag.local
    buildArgs:
      EMISSARY_BASE: emissary.local/emissary
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Short identifier, used to select builds on the CLI |
| `image` | yes | Docker image tag to build (e.g. `myapp:latest`) |
| `dir` | no | Working directory for `pre` commands and base for relative paths. Supports env vars (`$HOME`). Defaults to `.` |
| `pre` | no | Shell commands run sequentially before `docker build` (compilation, packaging, etc.) |
| `buildArgs` | no | Docker build arguments passed to `docker build --build-arg`. Keys are argument names, values are argument values. Supports env var expansion |
| `context` | no | Docker build context, relative to `dir`. Defaults to `.` |
| `dockerfile` | no | Path to the Dockerfile, relative to `dir`. Defaults to `Dockerfile` in the context |
| `platform` | no | Target platform for `docker build --platform` (e.g. `linux/amd64`). Useful when the base image is only available for a specific architecture |

## Running builds

```bash
sew build
```

This builds every entry, pushes each image to the cluster's preload registry, and restarts any Deployment or StatefulSet that references the image. Build output (pre-build commands, docker build logs) is captured in `~/.sew/logs/build/build.log` and the terminal shows a clean progress view.

### Building a subset

Pass one or more names to build only what you need:

```bash
sew build aes
```

### Skipping pre-build commands

When you've already compiled locally and just want to rebuild the Docker image:

```bash
sew build --skip-pre aes
```

### Building without restarting

Push the image to the registry but don't trigger a rollout restart:

```bash
sew build --no-restart aes
```

## Creating and building in one step

If you don't have a cluster yet, `--create` creates one before building. When the cluster already exists, the flag is silently ignored:

```bash
sew build --create
```

This runs the full `sew create` flow (preload, cluster, component install) then proceeds with the builds. You can pass context flags too -- they're forwarded to the creation step:

```bash
sew build --create --skip-pre aes
```

This is the fastest way to go from a clean machine to running your local code on a cluster.

## How builds interact with preloading

When `builds` is configured, sew automatically excludes build images from the preload list during `sew create`. If your context preloads `docker.io/datawire/aes:3.12.7` and you define a build for the same image, sew won't waste time pulling it from a remote registry -- it knows you'll push a local version.

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

## Patching a running cluster

`sew build` is for local code changes. When you need to bump an upstream image tag, change a Helm chart version, or tweak values on a running cluster without recreating it, use `sew patch`.

Write a patch file with the components you want to upgrade:

```yaml
components:
  - name: edge-stack
    helm:
      values:
        emissary-ingress:
          image:
            tag: 3.13.0
```

Then apply it:

```bash
sew patch upgrade.yaml
```

sew merges the patch into the resolved context and upgrades only the named components -- everything else is left untouched. You can preview the changes before applying them:

```bash
sew patch upgrade.yaml --dry-run
```

This runs a server-side dry-run and prints a colored diff of what would change.

See the [patch command reference]({{< ref "/docs/reference/commands#sew-patch" >}}) for the full set of flags and merge rules.
