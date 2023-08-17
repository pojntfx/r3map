# r3map

![Logo](./docs/logo-readme.png)

Re**m**ote **mm**ap: High-performance remote memory region mounts and migrations in user space.

[![hydrun CI](https://github.com/pojntfx/r3map/actions/workflows/hydrun.yaml/badge.svg)](https://github.com/pojntfx/r3map/actions/workflows/hydrun.yaml)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.20-61CFDD.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/pojntfx/r3map.svg)](https://pkg.go.dev/github.com/pojntfx/r3map)
[![Matrix](https://img.shields.io/matrix/r3map:matrix.org)](https://matrix.to/#/#r3map:matrix.org?via=matrix.org)

## Overview

🚧 This project is a work-in-progress! Instructions will be added as soon as it is usable. 🚧

r3map enables high-performance remote memory region mounts and migrations in user space.

It enables you to ...

- **Mount and migrate memory regions with a unified API**: r3map provides a consistent API, no matter if a memory region should simply be accessed or migrated between hosts.
- **Expose a resource with multiple frontends**: By providing multiple interfaces (such as a memory region and a file/path) for accessing or migrating a resource, integrating remote memory into existing applications is possible with little to no changes.
- **Transparently map almost any resource into memory, only fetching chunks as they are being read**: By exposing a simple backend interface and being fully transport-independent, r3map makes it possible to map resources such as a S3 bucket, Cassandra or Redis database, or even a tape drive into a memory region efficiently, as well as migrating a region over a framework and protocol of your choice, such as gRPC.
- **Use remote memory without the associated overhead**: Despite being in user space, r3map manages (on a [typical desktop system](https://pojntfx.github.io/networked-linux-memsync/main.html#testing-environment)) to achieve **high throughput (up to 3 GB/s)** with **minimal access latencies (~100µs)** and **short initialization times (~15ms)**.
- **Adapt to challenging network environments**: By implementing various optimizations such as background pull and push, two-phase protocols for migrations and concurrent device initialization, r3map can be deployed both in low-latency, high-throughput local datacenter networks and more constrained networks like the public internet.

The project is accompanied by a scientific thesis, which provides additional insights into design decisions, the internals of its implementation and comparisons to existing technologies and alternative approaches:

[Pojtinger, F. (2023). Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation](https://github.com/pojntfx/networked-linux-memsync)

## Installation

You can add r3map to your Go project by running the following:

```shell
$ go get github.com/pojntfx/r3map/...@latest
```

## Reference

### Mounts and Migrations

There are two fundamental use cases for r3map: Mounts and migrations. Mounting refers to accessing a resource, where the resource (such as a S3 bucket, remote file or memory region, tape drive etc.) is made available locally as either read-only or read-write, without having to download the entire resource first. Mounts work similarly to `mmap`, except the can map almost any resource into memory, not just files. To learn more about mounts, see the [Push-Pull Synchronization with Mounts](https://pojntfx.github.io/networked-linux-memsync/main.html#push-pull-synchronization-with-mounts) chapter in the accompanying thesis.

Migration refers to moving a resource like a memory region and moving it from one host to another. While mounts are optimized to have low initialization latencies and the best possible throughput performance, migrations are optimized to have the smallest possible downtime, where downtime refers to the typically short period in the migration process where neither the source nor the destination host can write to the resource that is being migrated. To optimize for this, migrations have a two-phase protocol which splits the device initialization and critical migration phases into two distinct parts, which keeps downtime to a minimum. To learn more about mounts, see the [Pull-Based Synchronization with Migrations](https://pojntfx.github.io/networked-linux-memsync/main.html#pull-based-synchronization-with-migrations) chapter in the accompanying thesis.

### Path, Slice and File Frontends

In order to make adoption of r3map for new and existing applications as frictionless as possible, multiple frontends with different layers of abstraction are provided, for both mounts and migrations. These frontends make it possible to access the resource with varying degrees of transparency, with individual chunks only being fetched as they are needed or being fetched pre-emptively, depending on the API chosen.

The path frontend is the simplest one; it simply exposes a resource as a block device and returns the path, which can then be read/written to/from by the application consuming the resource. The slice frontend adds a layer of indirection which exposes the resource as a `[]byte` by `mmap`ing the block device and integrating it with the resources' lifecycle, making it possible to access the resource in a more transparent way. Similarly so, the file frontend exposes the block device as a file integrated with the resource lifecycle, making it easier to use r3map for applications that already use a file interface.

> ⚠️ Note that the Go garbage collector is currently known to [cause deadlocks in some cases with the slice frontend](https://pojntfx.github.io/networked-linux-memsync/main.html#limitations) if the application using it runs in the same process. To work around this, prefer using the file frontend, or make sure that the client application is started in a separate process if the slice frontend is being used.

## License

r3map (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: Apache-2.0
