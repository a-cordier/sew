---
title: "Sew Registry"
weight: 1
type: docs
---

The registry is where you find ready-to-use application stacks. It's a structured tree of **contexts** -- each one a self-contained recipe for deploying a particular stack on a local Kind cluster.

## The default registry

sew ships with a curated registry of contexts for common building blocks -- databases, message brokers, search engines, and full application stacks. You can [browse it here]({{< ref "/registry" >}}). It's used by default, so you don't need to set `registry` in your config -- just pick a context and go:

```yaml
from:
  - elastic/elasticsearch/standalone
```

The default registry follows an `org/edition/product/variant` convention:

```
registry/
  mongodb/
    standalone/
      sew.yaml
  elastic/
    elasticsearch/
      standalone/
        sew.yaml
  postgresql/
    standalone/
      sew.yaml
  kafka/
    standalone/
      sew.yaml
```

Each leaf directory with a `sew.yaml` is a deployable context. You pick a context path, and sew fetches the config, creates a cluster, and deploys everything it defines.

### Default variants

When a product has multiple variants, the registry can define a **default** so you don't have to spell out the full path. A `.default` file in a directory contains the name of the variant to use:

```
registry/mycompany/myproduct/
├── .default          # contains "dev"
├── dev/
│   └── sew.yaml
└── staging/
    └── sew.yaml
```

With this setup, `from: [mycompany/myproduct]` resolves to `mycompany/myproduct/dev`. Defaults can chain across multiple levels -- sew reads `.default` at each level until it finds a `sew.yaml`.

## Hosting your own registry

A registry is just a directory tree served over HTTP. There's no server to run, no database to manage -- just files. If you can host static files, you can host a registry.

To create your own:

1. Create a directory structure following the `org/edition/product/variant` convention.
2. Add a `sew.yaml` to each leaf context (see [Context Format]({{< ref "/docs/reference/context-format" >}})).
3. Serve it however you like -- a Git repo's raw URL, a static file server, or an internal CDN.

Your team points their configs at your registry URL and uses your contexts just like the central ones:

```yaml
registry: https://registry.mycompany.com/sew
from:
  - myproduct/dev
```

You can also compose across registries. A context in your private registry can reference contexts from the central registry (or any other) through cross-registry composition -- see [Composing Contexts]({{< ref "/docs/guides/composing-contexts" >}}).

## Local filesystem registries

During development, you don't need to push your registry to a server. Point sew at a directory on your machine using a `file://` URL:

```yaml
registry: file://./my-registry
from:
  - myproduct/dev
```

This is the fastest way to iterate on new contexts. You edit your `sew.yaml`, run `sew create`, and see the result immediately -- no deploy step, no HTTP server.

A few common patterns:

```yaml
# Relative to the current directory
registry: file://./registry

# Absolute path
registry: file:///home/me/sew-contexts

# Navigate up from the project directory
registry: file://../shared-registry
```

Local registries behave exactly like remote ones -- composition, default variants, and all other features work the same way. Once your context is ready, push the directory tree to a server and switch the URL.
