---
title: "Getting Started"
weight: 1
type: docs
---

Get sew installed and deploy your first application stack in under a minute.

## Prerequisites

- **Go 1.25+** -- sew is installed via `go install`
- **Docker** -- sew uses it under the hood for Kind clusters

## Install

```bash
go install github.com/a-cordier/sew@latest
```

Make sure Docker is running before you continue.

## Create your first cluster

You don't even need a config file to get started. Pick a context from the registry and deploy it in one command:

```bash
sew create --from gravitee.io/apim
```

That's it. sew creates a Kind cluster, installs the Helm repos and components defined by the context, and gives you a full Gravitee API Management stack.

Once the cluster is ready, you can inspect it:

```bash
sew list        # shows running clusters
sew describe    # shows components, ports, and networking for the current cluster
```

When you're done:

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

You didn't need to write any Helm commands, manage chart repos, or wire up port mappings. The context handled it. See [Architecture]({{< ref "/docs/reference/architecture" >}}) for the full picture of how these components interact.

## What's next?

- **[Why sew?]({{< ref "/docs/getting-started/why-sew" >}})** -- the problems that motivated this tool and how each feature addresses them.
- **[Browse the registry]({{< ref "/docs/guides/sew-registry" >}})** to find contexts for databases, API gateways, and full application stacks.
- **[Compose contexts]({{< ref "/docs/guides/composing-contexts" >}})** to build complex stacks from simple building blocks.
- **[Set up networking]({{< ref "/docs/guides/networking" >}})** with load balancers, Gateway API, and local DNS.
- **[Build and iterate locally]({{< ref "/docs/guides/developer-loop" >}})** to rebuild images and reload them into the cluster in one command.
- **[Explore the CLI]({{< ref "/docs/reference/commands" >}})** to see everything sew can do.
