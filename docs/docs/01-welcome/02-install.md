---
sidebar_position: 1
description: Steps to install Ignite CLI on your local computer.
---

# Install Ignite CLI

You can run [Ignite CLI](https://github.com/ignite/cli) in a web-based IDE or you can install Ignite CLI on your local computer.

## Prerequisites

Be sure you have met the prerequisites before you install and use Ignite CLI.

### Operating systems

Ignite CLI is supported for the following operating systems:

- GNU/Linux
- macOS
- Windows Subsystem for Linux (WSL)

### Go

Ignite CLI is written in the Go programming language. To use Ignite CLI on a local system:

- Install [Go](https://golang.org/doc/install) (**version 1.23** or higher)
- Ensure the Go environment variables are [set properly](https://golang.org/doc/gopath_code#GOPATH) on your system

## Verify your Ignite CLI version

To verify the version of Ignite CLI you have installed, run the following command:

```bash
ignite version
```

## Installing Ignite CLI

To install the latest version of Ignite use [HomeBrew](https://formulae.brew.sh/formula/ignite) on macOS and GNU/Linux:

```sh
brew install ignite
```

Or use Snap on GNU/Linux:

```sh
snap install ignite --classic
```

### Install manually

Alternatively, you can install the latest version of the `ignite` binary use the following command:

```bash
curl https://get.ignite.com/cli! | bash
```

This command invokes `curl` to download the installation script and pipes the output to `bash` to perform the
installation.  The `ignite` binary is installed in `/usr/local/bin`.

Ignite CLI installation requires write permission to the `/usr/local/bin/` directory. If the installation fails because
you do not have write permission to `/usr/local/bin/`, run the following command:

```bash
curl https://get.ignite.com/cli | bash
```

Then run this command to move the `ignite` executable to `/usr/local/bin/`:

```bash
sudo mv ignite /usr/local/bin/
```

On some machines, a permissions error occurs:

```bash
mv: rename ./ignite to /usr/local/bin/ignite: Permission denied
============
Error: mv failed
```

In this case, use sudo before `curl` and before `bash`:

```bash
sudo curl https://get.ignite.com/cli | sudo bash
```

To learn more or customize the installation process, see the [installer docs](https://github.com/ignite/installer) on
GitHub.

## Upgrading your Ignite CLI installation {#upgrade}

Before you install a new version of Ignite CLI, remove all existing Ignite CLI installations.

To remove the current Ignite CLI installation:

1. On your terminal window, press `Ctrl+C` to stop the chain that you started with `ignite chain serve`.
2. Remove the Ignite CLI binary with `rm $(which ignite)`.
   Depending on your user permissions, run the command with or without `sudo`.
3. Repeat this step until all `ignite` installations are removed from your system.

After all existing Ignite CLI installations are removed, follow the  [Installing Ignite CLI](#installing-ignite-cli)
instructions.

For details on version features and changes, see
the [changelog.md](https://github.com/ignite/cli/blob/main/changelog.md)
in the repo.

## Build from source

To experiment with the source code, you can build from source:

```bash
git clone https://github.com/ignite/cli --depth=1
cd cli && make install
```

## Summary

- Verify the prerequisites.
- To set up a local development environment, install Ignite CLI locally on your computer.
- Install Ignite CLI by fetching the binary using cURL or by building from source.
- The latest version is installed by default. You can install previous versions of the precompiled `ignite` binary.
- Stop the chain and remove existing versions before installing a new version.
