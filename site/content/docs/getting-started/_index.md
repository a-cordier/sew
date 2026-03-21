---
title: "Getting Started"
weight: 1
type: docs
---

Get sew installed and deploy your first application stack in under a minute.

## Install

```bash
go install github.com/a-cordier/sew@latest
```

Make sure you have [Docker](https://docs.docker.com/get-docker/) running -- sew uses it under the hood for Kind clusters.

## Create your first cluster

You don't even need a config file to get started. Pick a context from the registry and deploy it in one command:

```bash
sew create --from gravitee.io/apim
```

That's it. sew creates a Kind cluster, installs the Helm repos and components defined by the context, and gives you a full Gravitee API Management stack. When you're done:

```bash
sew delete
```

## Using a config file

For anything beyond a quick test, you'll want a `sew.yaml` file. It lets you compose multiple contexts, add your own components, and configure networking:

```yaml
from:
  - gravitee.io/apim
```

Then just run `sew create` without flags. The config file is where things get interesting -- you can layer contexts, override values, enable DNS, and more. See [Composing Contexts]({{< ref "/docs/guides/composing-contexts" >}}) for the full story.

## What just happened?

Whether you used `--from` or a config file, sew did the same thing behind the scenes:

1. **Fetched the context** from the registry -- a `sew.yaml` describing which Helm charts and Kubernetes manifests to deploy, and how.
2. **Created a Kind cluster** with the port mappings, nodes, and settings specified by the context.
3. **Installed components** in dependency order -- adding Helm repos, running `helm upgrade --install`, and applying Kubernetes manifests.

You didn't need to write any Helm commands, manage chart repos, or wire up port mappings. The context handled it.

## What's next?

- **[Browse the registry]({{< ref "/docs/guides/registry" >}})** to find contexts for databases, API gateways, and full application stacks.
- **[Compose contexts]({{< ref "/docs/guides/composing-contexts" >}})** to build complex stacks from simple building blocks.
- **[Set up networking]({{< ref "/docs/guides/networking" >}})** with load balancers, Gateway API, and local DNS.
- **[Explore the CLI]({{< ref "/docs/reference/commands" >}})** to see everything sew can do.
