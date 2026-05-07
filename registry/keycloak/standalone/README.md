---
title: "Keycloak"
description: "Keycloak identity provider in dev mode"
tags: [security]
---

# Keycloak

Deploys a Keycloak server in development mode into a local Kind cluster
with host access on port 30880. Uses an embedded H2 database with no
persistence -- data is lost on restart.

## Usage

```bash
sew create --from keycloak/standalone
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
