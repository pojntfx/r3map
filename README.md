# r3map

![Logo](./docs/logo-readme.png)

Re**m**ote **mm**ap: High-performance remote memory region mounts and migrations in user space.

[![hydrun CI](https://github.com/pojntfx/r3map/actions/workflows/hydrun.yaml/badge.svg)](https://github.com/pojntfx/r3map/actions/workflows/hydrun.yaml)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.20-61CFDD.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/pojntfx/r3map.svg)](https://pkg.go.dev/github.com/pojntfx/r3map)
[![Matrix](https://img.shields.io/matrix/r3map:matrix.org)](https://matrix.to/#/#r3map:matrix.org?via=matrix.org)

## Overview

r3map enables high-performance remote memory region mounts and migrations in user space.

It enables you to ...

- **Mount and migrate memory regions with a unified API**: r3map provides a consistent API, no matter if a memory region should simply be accessed or migrated between hosts.
- **Expose a resource with multiple frontends**: By providing multiple interfaces (such as a memory region and a file/path) for accessing or migrating a resource, integrating remote memory into existing applications is possible with little to no changes.
- **Map almost any resource into memory**: By exposing a simple backend interface and being fully transport-independent, r3map makes it possible to map resources such as a S3 bucket, Cassandra or Redis database, or even a tape drive into a memory region efficiently, as well as migrating a region over a framework and protocol of your choice, such as gRPC.
- **Use remote memory without the associated overhead**: Despite being in user space, r3map manages (on a [typical desktop system](https://pojntfx.github.io/networked-linux-memsync/main.html#testing-environment)) to achieve **high throughput (up to 3 GB/s)** with **minimal access latencies (~100µs)** and **short initialization times (~15ms)**.
- **Adapt to challenging network environments**: By implementing various optimizations such as background pull and push, two-phase protocols for migrations and concurrent device initialization, r3map can be deployed both in low-latency, high-throughput local datacenter networks and more constrained networks like the public internet.

The project is accompanied by a scientific thesis, which provides additional insights into design decisions, the internals of its implementation and comparisons to existing technologies and alternative approaches:

[Pojtinger, F. (2023). Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation](https://github.com/pojntfx/networked-linux-memsync)

## Installation

You can add r3map to your Go project by running the following:

```shell
$ go get github.com/pojntfx/r3map/...@latest
```

## License

r3map (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: Apache-2.0
