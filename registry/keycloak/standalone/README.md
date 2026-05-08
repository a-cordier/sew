---
title: "Keycloak"
description: "Keycloak identity provider in dev mode"
tags: [security]
---

# Keycloak

Deploys a Keycloak server in development mode into a local Kind cluster
with host access on port 30880. Uses an embedded H2 database with no
persistence -- data is lost on restart.

## Install sew

```bash
go install github.com/a-cordier/sew@latest
```

For other installation methods, see [Installation](https://a-cordier.github.io/sew/docs/getting-started/installation/).

## Usage

### Create

```bash
sew create --from keycloak/standalone
```

### Cleanup

```bash
sew delete
```

## Quick Start

Open the Keycloak admin console from your host:

```bash
open http://localhost:30880
```

| Parameter | Value                   |
|-----------|-------------------------|
| Admin URL | http://localhost:30880   |
| Username  | admin                   |
| Password  | admin                   |
