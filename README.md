# NanoHUB

> **QA Wolf fork.** Branch `qawolf` is an upstream release tag plus our commits. Our releases are tagged
> `vX.Y.Z-qawolf.N`. CI runs gofmt, vet, build and tests on Linux with the Go version of our image build; no
> images or release archives are published from this fork.
> Changes on top of upstream:
>
> - `DELETE /api/v1/nanomdm/queue/{ids}` clears the command queue of each device ID (`cmd/nanohub/queue.go`).
>   It follows the queue API proposed in https://github.com/micromdm/nanomdm/pull/269. Drop it when upstream ships one.

[![CI/CD](https://github.com/micromdm/nanohub/actions/workflows/on-push-pr.yml/badge.svg)](https://github.com/micromdm/nanohub/actions/workflows/on-push-pr.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/micromdm/nanohub.svg)](https://pkg.go.dev/github.com/micromdm/nanohub)

NanoHUB adapts and unifies NanoMDM, NanoCMD, and KMFDDM into a single MDM server. Intended as a library it includes a reference command-line server as well.

## Getting started & Documentation

- [Go Reference](https://pkg.go.dev/github.com/micromdm/nanohub)  
Go package documentation for using NanoHUB as a library.

- [Operations Guide](docs/operations-guide.md)  
A brief overview of configuring and running the NanoHUB reference server.

## Getting the latest version

* Release `.zip` files containing the project should be attached to every [GitHub release](https://github.com/micromdm/nanohub/releases).
  * Release zips are also [published](https://github.com/micromdm/nanohub/actions) for every `main` branch commit.
* A Docker container is built and [published to the GHCR.io registry](http://ghcr.io/micromdm/nanohub) for every release.
  * `docker pull ghcr.io/micromdm/nanohub:latest` — `docker run ghcr.io/micromdm/nanohub:latest`
  * A Docker container is also published for every `main` branch commit (and tagged with `:main`)
* If you have a [Go toolchain installed](https://go.dev/doc/install) you can checkout the source and simply run `make`.
