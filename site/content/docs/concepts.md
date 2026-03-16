---
title: "Concepts"
weight: 2
type: docs
---

- **Registry** — A tree of context directories, either on the filesystem (`file:///path`) or over HTTP. The binary does not ship a registry; you use your own or a remote one.
- **Context** — A path inside the registry following `org/product/variant` (e.g. `gravitee.io/apim/db-less`). Each context has a `sew.yaml` that lists Helm repos and components (charts + values). If you omit the variant (e.g. `gravitee.io/apim`), sew looks for a `.default` file to pick one automatically (see [Default variant resolution]({{< ref "context-format#default-variant-resolution" >}})).
- **Config** — Configuration is layered. A **user-level** config at `$SEW_HOME/sew.yaml` (default `~/.sew/sew.yaml`) provides base settings; a **project-level** `./sew.yaml` (or `--config`) is merged on top. Each layer sets the registry URL, the contexts to compose via `from`, the Kind cluster definition, and optional local components and repos. Set the `SEW_HOME` environment variable to change the user-level config location.
