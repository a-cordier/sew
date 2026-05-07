---
title: "MailHog"
description: "SMTP testing server with web UI"
tags: [sink]
---

# MailHog

Deploys a MailHog email testing server into a local Kind cluster with
host access on port 30825 (web UI) and 31025 (SMTP). All received
emails are displayed in the web UI and stored in memory.

## Usage

```bash
sew create --from mailhog/standalone
```

## Quick Start

Open the MailHog web UI from your host:

```bash
open http://localhost:30825
```

Configure your application to send emails to `mailhog:1025` (in-cluster)
or `localhost:31025` (from the host).

| Parameter | Value                   |
|-----------|-------------------------|
| Web UI    | http://localhost:30825   |
| SMTP host | localhost               |
| SMTP port | 31025                   |
| Storage   | in-memory               |
