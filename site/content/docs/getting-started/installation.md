---
title: "Installation"
weight: 3
type: docs
---

## Homebrew

```bash
brew tap a-cordier/sew https://github.com/a-cordier/sew
brew install sew
```

## Linux packages

Download the `.deb` or `.rpm` from the [latest release](https://github.com/a-cordier/sew/releases/latest):

```bash
# Debian / Ubuntu
sudo dpkg -i sew_*_linux_amd64.deb

# Fedora / RHEL
sudo rpm -i sew_*_linux_amd64.rpm
```

## go install

```bash
go install github.com/a-cordier/sew@latest
```

Requires Go 1.25+.

## From source

```bash
git clone https://github.com/a-cordier/sew.git
cd sew
go build -o sew .
```
