---
title: "Eureka"
description: "Spring Cloud Eureka service registry"
tags: [networking]
---

# Eureka

Deploys a Spring Cloud Eureka server into a local Kind cluster with
host access on port 30761. Useful for service discovery in Spring Cloud
microservice stacks.

## Usage

```bash
sew create --from eureka/standalone
```

## Quick Start

Open the Eureka dashboard from your host:

```bash
open http://localhost:30761
```

| Parameter | Value                   |
|-----------|-------------------------|
| Dashboard | http://localhost:30761   |
| Auth      | none                    |
