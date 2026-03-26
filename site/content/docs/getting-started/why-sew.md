---
title: "Why sew?"
weight: 2
type: docs
---

sew came from day-to-day frictions faced by the Gravitee developer experience team while running local Kubernetes clusters for tests and demos. These are a few that kept coming back.

## Cluster setup is repetitive boilerplate

Every developer on the team had their own approach: a shell script here, a Kind config there, a README with a list of Helm commands to run in the right order. It worked, but it was fragile and nobody's setup matched anyone else's.

We wanted a single command that would produce a known-good cluster from a known-good definition. That's the core idea behind **registry contexts** -- curated stack descriptions that capture everything: Kind configuration, Helm charts, raw manifests, port mappings, networking. Pick one and deploy it:

```bash
sew create --from gravitee.io/apim
```

No scripts, no "follow the README and hope it's up to date."

## Stacks are duplicated across teams

The same MongoDB setup appeared in half a dozen repos, each with slightly different Helm values. When someone found a better configuration, it didn't propagate. Bug fixes in one copy didn't reach the others.

sew solves this with **composable contexts**. A MongoDB context lives in one place in the registry. An application context pulls it in with `from` and layers its own components on top. Abstract base contexts capture shared configuration that concrete variants extend. Context flags toggle optional pieces without forking the whole stack.

The result: one source of truth for each building block, reused everywhere it's needed.

## Pulling images takes forever

Destroy a cluster, recreate it, and wait 5 minutes for Docker to re-pull the same 2 GB of images you had 30 seconds ago. On a shared office connection, multiply that by the number of developers doing the same thing.

sew attacks this from two angles. **Image mirrors** are persistent pull-through cache containers that survive cluster restarts -- once an image is cached, it stays cached. **Preloading** pulls images on the host and pushes them to a local registry inside the cluster, so Kind nodes don't hit the network at all. Between the two, a second `sew create` is dramatically faster than the first.

## The edit-build-deploy loop is too slow

The inner dev loop on Kubernetes involves at least three manual steps: build the Docker image, load it into Kind, and trigger a rollout restart. Each step has its own flags and failure modes. It's the kind of friction that makes developers avoid testing on a real cluster.

**`sew build`** collapses the whole sequence into one command. Define your builds in `sew.yaml` -- source directory, pre-build commands, Dockerfile -- and sew handles the rest. It even skips preloading for images you build locally, since it knows you'll push a fresh copy.

## Config mistakes surface too late

A typo in a Helm value or a missing field in `sew.yaml` doesn't show up until Kind is running and the Helm install fails -- minutes into the process, after you've already waited for images to load.

**`sew validate`** checks your config against the embedded JSON Schema before anything starts. The same schema powers editor autocompletion and inline validation, so most mistakes never make it to the command line at all.

sew is the tool we wished we had from the start. If any of the above sounds familiar, give it a try -- the [Getting Started tutorial]({{< ref "/docs/getting-started" >}}) will get you to a running cluster in under a minute.
