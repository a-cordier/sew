---
title: "AI Toolchain"
weight: 5
type: docs
---

sew is designed from the ground up to work well with AI coding assistants. Whether you use Cursor, Codex, Copilot, or another agent, everything in the repo is structured so your assistant can understand the project, validate configs, and follow product-specific rules without guessing.

This page explains what makes sew agent-friendly, and how you as a maintainer can extend these capabilities for your own products.

## What makes sew AI-friendly

Three things work together:

1. **A typed config schema** -- `schema/sew.schema.yaml` is a JSON Schema that describes every field in `sew.yaml`. Agents use it to validate configs, discover available options, and generate correct YAML without reading Go source code.
2. **A structured registry** -- The `org/product/variant` convention, `.default` files, and YAML front matter in READMEs make the registry machine-parseable. An agent can navigate it, understand what each context does, and compose them together.
3. **Explicit agent rules** -- `AGENTS.md` at the repo root gives agents project-wide guidance, and product-specific instruction files in `agents/` provide domain rules scoped to particular registry subtrees.

## AGENTS.md

The `AGENTS.md` file sits at the repository root. AI coding assistants (Cursor, Codex, etc.) automatically discover and read it when they open the project. It tells the agent:

- **What sew is** -- a one-line description so the agent has context.
- **Where the schema lives** -- points to `schema/sew.schema.yaml` and tells the agent to prefer it over reading Go source code.
- **Agent-specific rules** -- guardrails like "always run `task lint` and `task test` before proposing changes" and "validate `sew.yaml` files against the schema."
- **Product-specific instructions** -- a table linking to dedicated instruction files for each product in the registry.

Here's what the structure looks like:

```markdown
# AGENTS.md

sew is a CLI tool that spins up local Kubernetes clusters and deploys
ready-to-use applications from a composable registry of contexts.

## Guidelines
Read and follow CONTRIBUTING.md for toolchain, commit conventions, ...

## Schema
The machine-readable JSON Schema for sew.yaml lives at
schema/sew.schema.yaml. Use it as the primary reference ...

## Agent-specific rules
- Do not add AI-attribution footers to commits
- Always run task lint and task test before proposing changes
- When authoring sew.yaml files, validate against the schema
- ...

## Product-specific instructions
| Product  | Instructions                  | Applies to            |
|----------|-------------------------------|-----------------------|
| Gravitee | agents/gravitee-agent.md      | registry/gravitee.io/ |
```

## Product-specific instructions

When a product has domain rules that agents should follow, you create a dedicated instruction file under `agents/`. These files are scoped to a specific subtree of the registry and contain rules that only apply when working on that product's contexts.

### File format

Each instruction file uses YAML front matter with two fields:

```yaml
---
product: MyProduct
paths:
  - registry/mycompany/
---

# MyProduct rules

These instructions apply when working on contexts under registry/mycompany/.

## Your domain rules here
...
```

- **`product`** -- The display name for the product table in `AGENTS.md`.
- **`paths`** -- Which registry paths these rules apply to. An agent working on files outside these paths can ignore the file.

### What belongs in a product instruction file

Product instructions should cover domain knowledge that an agent wouldn't otherwise have:

- **Organizational rules** -- how contexts are structured (e.g. `oss/` vs `ee/` splits, naming conventions).
- **Required patterns** -- things every context in this product must include (license handling, specific Helm values, mandatory components).
- **Override conventions** -- how users are expected to customize the product (which fields to override, which to leave alone).
- **README requirements** -- what sections or content must appear in each context's README for consistency.

For example, the Gravitee product instructions cover the OSS/EE split, how license keys are mounted as Kubernetes Secrets with `onMissing: ignore`, the required Helm values for license volumes, and the verbatim README section every EE context must include.

### Keeping the table in sync

The product table in `AGENTS.md` is auto-generated. When you add or remove an instruction file, run:

```bash
task agents:update
```

This scans `agents/*-agent.md` files, reads their front matter, and regenerates the table between the `<!-- agents:begin -->` and `<!-- agents:end -->` markers in `AGENTS.md`.

## The config schema

The JSON Schema at `schema/sew.schema.yaml` is the machine-readable contract for `sew.yaml`. It defines every field, its type, description, default value, and allowed values. Agents use it to:

- **Validate** existing configs -- catch typos, missing required fields, or invalid values before running `sew create`.
- **Generate** new configs -- produce correct YAML by following the schema's structure and constraints.
- **Discover** options -- understand what features are available (DNS, load balancers, Gateway API, image mirrors) without reading documentation.

The `AGENTS.md` file explicitly tells agents to prefer the schema over Go source code in `internal/config/`, which avoids the risk of the agent misinterpreting implementation details.

See the [Configuration reference]({{< ref "/docs/reference/configuration" >}}) for the full schema documentation.

## Registry structure

The registry's conventions are designed to be both human-readable and machine-parseable:

- **`org/product/variant` paths** -- predictable, navigable hierarchy.
- **`.default` files** -- plain-text pointers that agents can follow to resolve shorthand paths (e.g. `mycompany/product` resolves to `mycompany/product/dev` if `.default` contains `dev`).
- **README front matter** -- `title`, `description`, and `tags` in YAML front matter give agents structured metadata about each context without parsing free-text.
- **`sew.yaml` in every context** -- a single, schema-validated file that describes the entire stack. No implicit configuration, no magic.

## Making your product AI-friendly

If you maintain contexts in the sew registry and want AI assistants to work well with your product, follow this checklist:

1. **Create an agent instruction file** at `agents/<product>-agent.md` with YAML front matter (`product`, `paths`) and your domain rules.
2. **Run `task agents:update`** to regenerate the product table in `AGENTS.md`.
3. **Set `.default` files** in your registry subtree so agents (and users) can reference your product with shorthand paths.
4. **Add front matter to every README** (`title`, `description`, `tags`) so the registry browser and agents can discover your contexts.
5. **Validate your `sew.yaml` files against the schema** -- if the schema doesn't cover a field you need, contribute to the schema first.
6. **Document required patterns** in your agent file -- license handling, naming conventions, mandatory components, anything an agent needs to know to produce correct contexts.
