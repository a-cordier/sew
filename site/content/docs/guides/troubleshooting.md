---
title: "Troubleshooting"
weight: 6
type: docs
---

Common issues and how to fix them.

## Docker is not running

**Symptom:** `sew create` or `sew delete` fails with `"docker is not running — start Docker and try again"`.

**Cause:** sew checks Docker availability before doing any work. If the Docker daemon is unreachable, the command exits immediately with this message.

**Fix:** Start Docker Desktop (or `systemctl start docker` on Linux), then verify with:

```bash
docker info
```

## Port already in use

**Symptom:** Container creation or the DNS server fails to start; the error mentions "address already in use" or "port is already allocated."

**Cause:** sew uses fixed host ports: **15353** (DNS), **5000+** (mirror proxies), **5100** (preload registry). Kind contexts also bind host ports (e.g. 80, 443, 9200). There is no preflight port check.

**Fix:** Identify what's using the port:

```bash
lsof -i :<port>
```

Stop the conflicting process, or change the port in config (`features.dns.port` for DNS; Kind port mappings can be overridden in `sew.yaml`). Mirror and preload ports are currently not configurable — stop the conflicting service.

## DNS not resolving after `sew create`

**Symptom:** `curl api.sew.local` times out or returns NXDOMAIN.

**Cause:** OS-level routing was never configured (one-time `sew setup dns` step).

**Fix:** Run the DNS setup (requires sudo) and check status afterwards:

```bash
sew setup dns
sew status
```

The DNS section of `sew status` shows resolver, server, and records state. See the [Networking guide]({{< ref "/docs/guides/networking/#local-dns" >}}) for details.

## `sudo` prompt on macOS when creating a cluster

**Symptom:** macOS asks for your password during `sew create` with load balancers enabled.

**Cause:** Docker on macOS runs in a VM; sew sets up a packet tunnel to make container IPs routable from the host, which requires root.

**Fix:** This is expected. Enter your password. Subsequent `sew create` calls in the same session won't re-prompt unless the tunnel was torn down.

## `gateway requires lb, but lb is explicitly disabled`

**Symptom:** `sew create` fails with this validation error.

**Cause:** Gateway API needs load balancers to assign IPs to Gateways. The config has `features.gateway.enabled: true` but `features.lb.enabled: false`.

**Fix:** Either enable load balancers explicitly:

```yaml
features:
  lb:
    enabled: true
```

Or let sew auto-enable it by removing the explicit `lb` block entirely.

## Component not ready (timeout)

**Symptom:** `sew create` hangs and eventually fails with "timeout waiting for X to be ready."

**Cause:** A pod isn't reaching Ready state within the timeout (default 5 min). Common reasons: image pull failure, crash loop, missing dependency, resource limits.

**Fix:** Check pod status and logs:

```bash
kubectl get pods -A
kubectl logs -n <ns> <pod>
```

Increase the timeout in `sew.yaml` if the component just needs more time. Check `sew status` for image preloading or mirror issues.

## Image pull errors / Docker Hub rate limiting

**Symptom:** Pods stuck in `ImagePullBackOff` or `ErrImagePull`. Docker Hub returns 429 (Too Many Requests).

**Cause:** Anonymous Docker Hub pulls are rate-limited (100 pulls / 6 hours). Heavy image contexts can hit this on first run.

**Fix:** Enable image mirrors to cache layers locally:

```yaml
images:
  mirrors: {}
```

For authenticated pulls, configure `~/.docker/config.json` — sew's mirror proxies forward credentials.

## Stale cluster after a crash

**Symptom:** `sew create` fails because the Kind cluster already exists, or Docker containers from a previous run are still around.

**Cause:** A previous `sew delete` didn't complete (crash, Ctrl+C, Docker restart).

**Fix:** Run a clean teardown:

```bash
sew delete --name <cluster>
```

If no state file exists, sew does best-effort cleanup. As a last resort:

```bash
kind delete cluster --name <name>
docker rm -f $(docker ps -aq --filter "name=sew-mirror-")
docker rm -f sew-preload
```

## Context not found (404)

**Symptom:** `sew create` fails with "fetching context: 404 Not Found."

**Cause:** Typo in the `from` path, or the registry URL doesn't point to a valid registry tree.

**Fix:** Check the path in `from` against the [registry browser]({{< ref "/registry" >}}). For custom registries, verify the URL is reachable and the directory structure contains a `sew.yaml` at the expected path.

## `context is abstract and cannot be deployed directly`

**Symptom:** Error when running `sew create` with `--from` pointing to an abstract context.

**Cause:** Abstract contexts are shared bases; they must be composed via `from` in a concrete context.

**Fix:** Use a concrete variant instead. Check the registry for available variants under the same product path.
