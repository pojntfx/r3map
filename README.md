# r3map

![Logo](./docs/logo-readme.png)

Re**m**ote **mm**ap: High-performance remote memory region mounts and migrations in user space.

[![hydrun CI](https://github.com/pojntfx/r3map/actions/workflows/hydrun.yaml/badge.svg)](https://github.com/pojntfx/r3map/actions/workflows/hydrun.yaml)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.20-61CFDD.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/pojntfx/r3map.svg)](https://pkg.go.dev/github.com/pojntfx/r3map)
[![Matrix](https://img.shields.io/matrix/r3map:matrix.org)](https://matrix.to/#/#r3map:matrix.org?via=matrix.org)

## Overview

r3map is a library that simplifies working with remote memory regions and migrating them between hosts.

It can ...

- **Create a virtual `[]byte` or a virtual file that transparently downloads remote chunks only when they are accessed**: By providing multiple frontends (such as a memory region and a file/path) for accessing or migrating a resource, integrating remote memory into existing applications is possible with little to no changes, and fully language-independent.
- **`mmap` any local or remote resource instead of just files**: By exposing a simple backend interface and being fully transport-independent, r3map makes it possible to map resources such as a **S3 bucket, Cassandra or Redis database**, or even a tape drive into a memory region efficiently, as well as migrating it over an RPC framework of your choice, such as gRPC.
- **Enable live migration features for any hypervisor or application**: r3map implements the APIs which allow for zero-downtime live migration of virtual machines, but makes them generic so that they can be used for any memory region, bringing live migration abilities to almost any hypervisor or application with minimal changes and overhead.
- **Overcome the performance issues typically associated with remote memory**: Despite being in user space, r3map manages (on a [typical desktop system](https://pojntfx.github.io/networked-linux-memsync/main.html#testing-environment)) to achieve **very high throughput (up to 3 GB/s)** with **minimal access latencies (~100µs)** and **short initialization times (~12ms)**.
- **Adapt to challenging network environments**: By implementing various optimizations such as **background pull and push**, two-phase protocols for migrations and concurrent device initialization, r3map can be deployed not only in low-latency, high-throughput local datacenter networks but also in more constrained networks like the public internet.

The project is **accompanied by a scientific thesis**, which provides additional insights into design decisions, the internals of its implementation and comparisons to existing technologies and alternative approaches:

<p align="center">
	<a href="https://pojntfx.github.io/networked-linux-memsync/main.pdf" rel="nofollow"><img src="./docs/thesis-badge.png" alt="Thesis badge for Pojtinger, F. (2023). Efficient Synchronization of Linux Memory Regions over a Network: A Comparative Study and Implementation" width="650"></a>
</p>

For more information, you can find the related benchmark and LaTeX code in the [thesis repository](https://github.com/pojntfx/networked-linux-memsync).

## Installation

You can add r3map to your Go project by running the following:

```shell
$ go get github.com/pojntfx/r3map/...@latest
```

## Usage

### 1. Mapping a Remote Resource into Memory with the Direct Mount API

> TL;DR: Create a backend, find a free block device, start the direct mount and interact with the resource through a frontend of your choice (`[]byte`/file etc.)

The direct mount API is the simplest way of accessing a resource. In order to make a resource available, either a custom backend can be created or one of the available example backends can be used (see the [backends reference](#backends) for more information), such as a S3 bucket or a Redis database. Similarly so, multiple frontends, which represent means to access a resource, are also available: The path frontend, which simply exposes the path to a block device with the resource, the file frontend, which exposes the resource by opening the block device and integrating the file's lifecycle, and the slice/`[]byte` frontend, which **makes the resource available as a `[]byte`, only fetching chunks from the backend as they are being accessed**. For the file frontend, this means that if call `.Read(offset, chunk)`, only `len(chunk)` at `offset` will be fetched, while for a slice, if you access `sliceFrontend[lowerBound:upperBound]`, only the chunks from `lowerBound` to `upperBound` will be fetched as they are being accessed. For more information, check out the [frontends reference](#path-slice-and-file-frontends).

#### 1.1 Setting up a Backend

<details>
  <summary>Expand instructions</summary>

For this usage example, we'll use a simple, local example backend: The file backend, which can be set up with following:

```go
f, err := os.CreateTemp("", "")
if err != nil {
	panic(err)
}
defer os.RemoveAll(f.Name())

if err := f.Truncate(*size); err != nil {
	panic(err)
}

b := backend.NewFileBackend(f)
```

Next, a free block device (which provides the mechanism for intercepting reads and writes) needs to be opened:

```go
devPath, err := utils.FindUnusedNBDDevice()
if err != nil {
	panic(err)
}

devFile, err := os.Open(devPath)
if err != nil {
	panic(err)
}
defer devFile.Close()
```

</details>

#### 1.2 Setting up a Frontend

<details>
  <summary>Expand instructions</summary>

For this simple example, we'll use the file frontend, which can be set up and initialized like this; it provides a virtual file, which fetches offsets on demand:

```go
mnt := mount.NewDirectFileMount(
	b,
	devFile,

	nil,
	nil,
)

var wg sync.WaitGroup
wg.Add(1)
go func() {
	defer wg.Done()

	if err := mnt.Wait(); err != nil {
		panic(err)
	}
}()

done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt)
go func() {
	<-done

	log.Println("Exiting gracefully")

	_ = mnt.Close()
}()

defer mnt.Close()
file, err := mnt.Open()
if err != nil {
	panic(err)
}

log.Println("Resource available on", file.Name())

wg.Wait()
```

Note that the interrupt signal has been intercepted to gracefully close the mount, which helps prevent data loss when stopping the process. Here, the file provided by the file frontend is simply used to print the path to the resource; in real-world scenarios, the file (or `[]byte`) provided can be interacted with directly.

</details>

#### 1.3 Demonstration

<details>
  <summary>Expand instructions</summary>

The mount can then be started like this, and should output the following; see the [full code of the example for more](./cmd/r3map-example-direct-mount-file/main.go):

```shell
$ sudo modprobe nbd # This is only necessary once, and loads the NBD kernel module
$ go build -o /tmp/r3map-example-direct-mount-file ./cmd/r3map-example-direct-mount-file/ && sudo /tmp/r3map-example-direct-mount-file
2023/08/18 16:39:18 Resource available on /dev/nbd0
```

The resource can now be interacted as though it were any file, for example by reading and writing a string to/from it:

```shell
$ echo 'Hello, world!' | sudo tee /dev/nbd0
Hello, world!
$ sudo cat /dev/nbd0
Hello, world!
```

</details>

For more information on the direct mount, as well as available configuration options, usage examples for different frontends and backends, and using a remote resource instead of a local one, see the [direct mount benchmark](./cmd/r3map-benchmark-direct-mount/main.go) and [direct mount Go API reference](https://pkg.go.dev/github.com/pojntfx/r3map/pkg/mount#DirectFileMount).

### 2. Efficiently Mounting a Remote Resource with the Managed Mount API

> TL;DR: Create a server exposing a backend, create a client that connects to the server, and interact with the resource through a frontend of your choice (`[]byte`/file etc.)

While the direct mount API is a good choice for mounting a resource if there is little to no latency, the managed mount API is the better choice if the resource is remote, esp. in networks with high latencies such as the public internet. Instead of the reads and writes being forwarded synchronously to the backend, the asynchronous background push- and pull system can take advantage of multiple connections and concurrent push/pull to significantly increase throughput and decrease access latency, as well as pre-emptively pulling specific offsets first (see the [mounts reference](#direct-mounts-managed-mounts-and-pull-priority) for more information).

#### 2.1 Setting up a Server

<details>
  <summary>Expand instructions</summary>

While it is possible to use [any of the available backends](#backends) or creating a custom one, we'll be creating a client and server system, where a gRPC server exposes a resource backed by a file, and a managed mount uses a gRPC client to mount the resource. Note that since r3map is fully transport independent, there are other options available as well, such as fRPC and panrpc, which [can have different characteristics depending on network conditions and other factors](https://pojntfx.github.io/networked-linux-memsync/main.html#rpc-frameworks-1). To create the server exposing the resource, first the backend and gRPC need to be set up:

```go
f, err := os.CreateTemp("", "")
if err != nil {
	panic(err)
}
defer os.RemoveAll(f.Name())

if err := f.Truncate(*size); err != nil {
	panic(err)
}

srv := grpc.NewServer()

v1.RegisterBackendServer(
	srv,
	services.NewBackendServiceGrpc(
		services.NewBackend(
			backend.NewFileBackend(f),
			*verbose,
			services.MaxChunkSize,
		),
	),
)
```

After both are available, the server is attached to a TCP listener:

```go
lis, err := net.Listen("tcp", *laddr)
if err != nil {
	panic(err)
}
defer lis.Close()

log.Println("Listening on", *laddr)

if err := srv.Serve(lis); err != nil && !utils.IsClosedErr(err) {
	panic(err)
}
```

The server can then be started like this, and should output the following; see the [full code of the example for more](./cmd/r3map-example-mount-server/main.go):

```shell
$ go run ./cmd/r3map-example-mount-server/
2023/08/19 18:32:26 Listening on localhost:1337
```

</details>

#### 2.2 Setting up a Client

<details>
  <summary>Expand instructions</summary>

On the client side, we connect to this server and set up a local backend (which caches backgrounds reads/writes; here we use a simple file backend, but any backend can be used):

```go
f, err := os.CreateTemp("", "")
if err != nil {
	panic(err)
}
defer os.RemoveAll(f.Name())

if err := f.Truncate(*size); err != nil {
	panic(err)
}

conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
	panic(err)
}
defer conn.Close()

log.Println("Connected to", *raddr)
```

Similarly to direct mounts, any frontend can be chosen (such as the `[]byte` frontend), but for simplicity we'll use the file frontend here, which can be started like this:

```go
mnt := mount.NewManagedFileMount(
	ctx,

	lbackend.NewRPCBackend(
		ctx,
		services.NewBackendRemoteGrpc(
			v1.NewBackendClient(conn),
		),
		*size,
		false,
	),
	backend.NewFileBackend(f),

	&mount.ManagedMountOptions{
		Verbose: *verbose,
	},
	nil,

	nil,
	nil,
)

var wg sync.WaitGroup
wg.Add(1)
go func() {
	defer wg.Done()

	if err := mnt.Wait(); err != nil {
		panic(err)
	}
}()

done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt)
go func() {
	<-done

	log.Println("Exiting gracefully")

	_ = mnt.Close()
}()

defer mnt.Close()
file, err := mnt.Open()
if err != nil {
	panic(err)
}

log.Println("Resource available on", file.Name())

wg.Wait()
```

Just like with the direct mount API, the interrupt signal has been intercepted to gracefully close the mount, which helps prevent data loss when stopping the process by flushing the remaining changes to the backend. Here, the file provided by the file frontend is simply used to print the path to the resource; in real-world scenarios, the file (or `[]byte`) provided can be interacted with directly.

</details>

#### 2.3 Demonstration

<details>
  <summary>Expand instructions</summary>

Just like with the direct mount, the managed mount can be started like this, and should output the following; see the [full code of the example for more](./cmd/r3map-example-managed-mount-file/main.go):

```shell
$ sudo modprobe nbd # This is only necessary once, and loads the NBD kernel module
$ go build -o /tmp/r3map-example-managed-mount-client ./cmd/r3map-example-managed-mount-client/ && sudo /tmp/r3map-example-managed-mount-client
2023/08/19 18:59:16 Connected to localhost:1337
2023/08/19 18:59:16 Resource available on /dev/nbd0
```

The resource can now be interacted as though it were any file, for example by reading and writing a string to/from it:

```shell
$ echo 'Hello, world!' | sudo tee /dev/nbd0
Hello, world!
$ sudo cat /dev/nbd0
Hello, world!
```

Note that if the client is stopped and started again, the data of the resource will continue to be available, since it is stored on the server and remaining changes are flushed to the server on exit and by the background push system.

</details>

For more information on the managed mount, as well as available configuration options, usage examples for different frontends and backends, see the [managed mount benchmark](./cmd/r3map-benchmark-managed-mount/main.go) and [managed mount Go API reference](https://pkg.go.dev/github.com/pojntfx/r3map/pkg/mount#ManagedFileMount).

### 3. Migrating a Memory Region Between Two Hosts with the Migration API

> TL;DR: Create a backend, attach it to a migrator, start seeding the resource on one host, leech the resource from a second host, finalize the migration and interact with the resource through a frontend of your choice (`[]byte`/file etc.) or start seeding the resource again

While mounts offer a universal method for accessing resources, migrations are optimized for scenarios that involve moving a resource between two hosts. This is because they use a two-phase protocol to minimize the time the resource is unavailable during migration (see the [mounts and migrations reference](#mounts-and-migrations) for more information). Migrations are also peer-to-peer, meaning that no intermediary/remote backend is required, which reduces the impact of network latency on the migration. There are two actors in a migration: The seeder, from which a resource can be migrated from, and a leecher, which migrates a resource to itself. Similarly to mounts, different frontends (such as `[]byte` or the file frontend), backends (for locally storing the resource) and transports (like gRPC) can be chosen.

#### 3.1 Setting up a Migrator

<details>
  <summary>Expand instructions</summary>

For this example, we'll start by creating a migrator (the component that handles both seeding and leeching) with a file frontend, file backend and gRPC transport:

```go
f, err := os.CreateTemp("", "")
if err != nil {
	panic(err)
}
defer os.RemoveAll(f.Name())

if err := f.Truncate(*size); err != nil {
	panic(err)
}

mgr := migration.NewFileMigrator(
	ctx,

	backend.NewFileBackend(f),

	&migration.MigratorOptions{
		Verbose: *verbose,
	},
	&migration.MigratorHooks{
		OnBeforeSync: func() error {
			log.Println("Suspending app")

			return nil
		},
		OnAfterSync: func(dirtyOffsets []int64) error {
			delta := (len(dirtyOffsets) * client.MaximumBlockSize)

			log.Printf("Invalidated: %.2f MB (%.2f Mb)", float64(delta)/(1024*1024), (float64(delta)/(1024*1024))*8)

			return nil
		},

		OnBeforeClose: func() error {
			log.Println("Stopping app")

			return nil
		},

		OnChunkIsLocal: func(off int64) error {
			log.Printf("Chunk %v has been leeched")

			return nil
		},
	},

	nil,
	nil,
)
```

Note the use of the hook functions; these allow for integrating the migration with the application lifecycle, and can be used for notifying an application which is accessing it to suspend or shut down access to the resource when the migration lifecycle requires it, as well as monitoring the migration progress with `OnChunkIsLocal`. The migrator can be started similarly to how the mounts are started, including registering the interrupt handler to prevent data loss on exit:

```go
var wg sync.WaitGroup
wg.Add(1)
go func() {
	defer wg.Done()

	if err := mgr.Wait(); err != nil {
		panic(err)
	}
}()

done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt)
go func() {
	<-done

	log.Println("Exiting gracefully")

	_ = mgr.Close()
}()
```

</details>

#### 3.2 Seeding a Resource

<details>
  <summary>Expand instructions</summary>

Note that the migrator is able to both seed and leech a resource; to start seeding, call `Seed()` on the migrator:

```go
defer mgr.Close()
file, svc, err := mgr.Seed()
if err != nil {
	panic(err)
}

log.Println("Starting app on", file.Name())
```

The resulting file can be used to interact with the resource just like with mounts, and the service can then be attached to a gRPC server, making it available over a network:

```go
server := grpc.NewServer()

v1.RegisterSeederServer(server, services.NewSeederServiceGrpc(svc))

lis, err := net.Listen("tcp", *laddr)
if err != nil {
	panic(err)
}
defer lis.Close()

log.Println("Seeding on", *laddr)

go func() {
	if err := server.Serve(lis); err != nil {
		if !utils.IsClosedErr(err) {
			panic(err)
		}

		return
	}
}()
```

Finally, we register an invalidation handler for this example; this allows simulating writes done by the application using a resource as it is being seeded by pressing <kbd>Enter</kbd>:

```go
go func() {
	log.Println("Press <ENTER> to invalidate resource")

	bufio.NewScanner(os.Stdin).Scan()

	log.Println("Invalidating resource")

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		panic(err)
	}

	if _, err := io.CopyN(
		file,
		rand.Reader,
		int64(math.Floor(
			float64(*size)*(float64(*invalidate)/float64(100)),
		)),
	); err != nil {
		panic(err)
	}
}()
```

</details>

#### 3.3 Leeching a Resource

<details>
  <summary>Expand instructions</summary>

Setting up a leecher is similar to setting up a seeder, and starts by connecting to the gRPC server provided by the seeder as well as calling `Leech()` on the migrator:

```go
conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
	panic(err)
}
defer conn.Close()

log.Println("Leeching from", *raddr)

defer mgr.Close()
finalize, err := mgr.Leech(services.NewSeederRemoteGrpc(v1.NewSeederClient(conn)))
if err != nil {
	panic(err)
}
```

This will start leeching the resource in the background, and the `OnChunkLocal` callback can be used to monitor the download progress. In order to "finalize" the migration, which tells the seeder to suspend the application using the resource and marks chunks that were changed since the migration started as "dirty", we can call `finalize()` once the <kbd>Enter</kbd> key is pressed (for more information on the migration protocol, see the [Pull-Based Synchronization with Migrations](https://pojntfx.github.io/networked-linux-memsync/main.html#pull-based-synchronization-with-migrations) chapter in the accompanying thesis):

```go
log.Println("Press <ENTER> to finalize migration")

bufio.NewScanner(os.Stdin).Scan()
// ...

seed, file, err := finalize()
if err != nil {
	panic(err)
}

log.Println("Resuming app on", file.Name())
```

The resulting file can be used to interact with the resource, just like with mounts. It is also possible to for a leecher to start seeding the resource; this works by calling the resulting `seed()` closure, and attaching it to a gRPC server just like when calling `Seed()` on the migrator:

```go
svc, err := seed()
if err != nil {
	panic(err)
}

server := grpc.NewServer()
// ...
```

</details>

#### 3.4 Demonstration

<details>
  <summary>Expand instructions</summary>

To demonstrate the migration, a seeder can now be started like this (with the migrator being instructed to seed by providing `--laddr`), and should output the following; see the [full code of the example for more, which covers both the seeder and the leecher](./cmd/r3map-example-migration/main.go):

```shell
$ sudo modprobe nbd # This is only necessary once, and loads the NBD kernel module
$ go build -o /tmp/r3map-example-migration ./cmd/r3map-example-migration/ && sudo /tmp/r3map-example-migration --laddr localhost:1337 --invalidate 10
2023/08/20 00:27:19 Starting app on /dev/nbd0
2023/08/20 00:27:19 Seeding on localhost:1337
2023/08/20 00:27:19 Press <ENTER> to invalidate resource
```

This makes the resource available on the specified path, and starts seeding; invalidating the resource (which simulates an application using it) is also possible by pressing <kbd>Enter</kbd>. To start migrating the application away from this first seeder, a leecher can be started by specifying a `--raddr`:

```shell
$ sudo modprobe nbd # This is only necessary once, and loads the NBD kernel module
$ go build -o /tmp/r3map-example-migration ./cmd/r3map-example-migration/ && sudo /tmp/r3map-example-migration --raddr localhost:1337 --laddr localhost:1337 --invalidate 10
2023/08/20 01:05:32 Leeching from localhost:1337
2023/08/20 01:05:32 Press <ENTER> to finalize migration
Pulling 100% [==============================================] (512/512 MB, 510 MB/s) [1s:0s]
```

This starts pulling chunks in the background, and the migration can be finalized by pressing <kbd>Enter</kbd>:

```shell
2023/08/20 01:06:38 Invalidated: 0.00 MB (0.00 Mb)
2023/08/20 01:06:38 Resuming app on /dev/nbd1
2023/08/20 01:06:38 Seeding on localhost:1337
2023/08/20 01:06:38 Press <ENTER> to invalidate resource
```

The resource can now be interacted with on the provided path. Since the `--laddr` flag was also provided, the resource is also being seeded after being fully leeched, which makes it possible to migrate it to another leecher. This time the seeder mode is disabled by not specifying `--laddr`, and the migration is finalized with <kbd>Enter</kbd>:

```shell
$ sudo modprobe nbd # This is only necessary once, and loads the NBD kernel module
$ go build -o /tmp/r3map-example-migration ./cmd/r3map-example-migration/ && sudo /tmp/r3map-example-migration --raddr localhost:1337
2023/08/20 01:10:52 Leeching from localhost:1337
2023/08/20 01:10:52 Press <ENTER> to finalize migration
Pulling 100% [==============================================] (512/512 MB, 508 MB/s) [1s:0s]

2023/08/20 01:10:58 Invalidated: 0.00 MB (0.00 Mb)
2023/08/20 01:10:58 Resuming app on /dev/nbd1
```

The resource can now be interacted with as though it were any file again, for example by reading and writing a string to/from it; no additional seeder has been started, thus terminating the migration chain:

```shell
$ echo 'Hello, world!' | sudo tee /dev/nbd0
Hello, world!
$ sudo cat /dev/nbd0
Hello, world!
```

</details>

For more information on the migration, as well as available configuration options, usage examples for different frontends and backends, see the [migration benchmark](./cmd/r3map-benchmark-migration/main.go) and [migration Go API reference](https://pkg.go.dev/github.com/pojntfx/r3map/pkg/migration#FileMigrator).

🚀 That's it! We can't wait to see what you're going to build with r3map. Be sure to take a look at the [reference](#reference), [additional examples and related projects](#examples) and the [accompanying research paper](https://pojntfx.github.io/networked-linux-memsync/main.pdf) for more information.

## Examples

To make getting started with r3map easier, take a look at the following simple usage examples:

- [Direct Mount Example (file backend)](./cmd/r3map-example-direct-mount-file/main.go)
- [Direct and Managed Mount Example Server (gRPC-based)](./cmd/r3map-example-mount-server/main.go)
- [Direct Mount Example Client (gRPC-based)](./cmd/r3map-example-direct-mount-client/main.go)
- [Managed Mount Example Client (gRPC-based)](./cmd/r3map-example-managed-mount-client/main.go)
- [Migration Example (gRPC-based)](./cmd/r3map-example-migration/main.go)

The benchmarks also serve as much more detailed examples, highlighting different configuration options, transports and backends:

- [Direct and Managed Mount Benchmark Server](./cmd/r3map-benchmark-mount-server/main.go)
- [Direct Mount Benchmark](./cmd/r3map-benchmark-direct-mount/main.go)
- [Managed Mount Benchmark](./cmd/r3map-benchmark-managed-mount/main.go)
- [Migration Benchmark Server](./cmd/r3map-benchmark-migration-server/main.go)
- [Migration Benchmark](./cmd/r3map-benchmark-migration/main.go)

For more in-depth, real-world use cases, check out these related projects which use r3map:

- [tapisk](https://github.com/pojntfx/tapisk) exposes a tape drive as a block device using r3map's managed mounts and chunking system.
- [ram-dl](https://github.com/pojntfx/ram-dl) uses r3map to share RAM/swap space between two hosts with direct mounts.

## Reference

### Mounts and Migrations

There are two fundamental use cases for r3map: Mounts and migrations. Mounting refers to accessing a resource, where the resource (such as a S3 bucket, remote file or memory region, tape drive etc.) is made available locally as either read-only or read-write, without having to download the entire resource first. Mounts work similarly to `mmap`, except the can map almost any resource into memory, not just files. To learn more about mounts, see the [Push-Pull Synchronization with Mounts](https://pojntfx.github.io/networked-linux-memsync/main.html#push-pull-synchronization-with-mounts) chapter in the accompanying thesis.

Migration refers to moving a resource like a memory region and moving it from one host to another. While mounts are optimized to have low initialization latencies and the best possible throughput performance, migrations are optimized to have the smallest possible downtime, where downtime refers to the typically short period in the migration process where neither the source nor the destination host can write to the resource that is being migrated. To optimize for this, migrations have a two-phase protocol which splits the device initialization and critical migration phases into two distinct parts, which keeps downtime to a minimum. To learn more about migrations, see the [Pull-Based Synchronization with Migrations](https://pojntfx.github.io/networked-linux-memsync/main.html#pull-based-synchronization-with-migrations) chapter in the accompanying thesis.

### Path, Slice and File Frontends

In order to make adoption of r3map for new and existing applications as frictionless as possible, multiple frontends with different layers of abstraction are provided, for both mounts and migrations. These frontends make it possible to access the resource with varying degrees of transparency, with individual chunks only being fetched as they are needed or being fetched pre-emptively, depending on the API chosen.

The path frontend is the simplest one; it simply exposes a resource as a block device and returns the path, which can then be read/written to/from by the application consuming the resource. The slice frontend adds a layer of indirection which exposes the resource as a `[]byte` by `mmap`ing the block device and integrating it with the resources' lifecycle, making it possible to access the resource in a more transparent way. Similarly so, the file frontend exposes the block device as a file integrated with the resource lifecycle, making it easier to use r3map for applications that already use a file interface.

> ⚠️ Note that the Go garbage collector is currently known to [cause deadlocks in some cases with the slice frontend](https://pojntfx.github.io/networked-linux-memsync/main.html#limitations) if the application using it runs in the same process. To work around this, prefer using the file frontend, or make sure that the client application is started in a separate process if the slice frontend is being used.

### Direct Mounts, Managed Mounts and Pull Priority

Direct mounts serve as the simplest mount API, and allow directly mapping a resource into memory. These mounts can be either read-only or read-write, and simply forwards reads/writes between a backend (such as a S3 bucket) and the memory region/frontend. Direct mounts work well in LAN deployments, but since chunks are only fetched from the backend as they are being accessed, and writes are immediately forwarded to the backend too, this can lead to performance issues in high-latency deployments like the public internet, since reads need to be synchronous.

In contrast to this, the managed mount API allows for smart background pull and push mechanisms. This makes it possible to pre-emptively fetch chunks before they are being accessed, and writing back changes periodically. Since this can be done concurrently and asynchronously, managed mounts are much less vulnerable to networks with high RTT like the internet, where they [can significantly increase throughput and decrease access latency](https://pojntfx.github.io/networked-linux-memsync/main.html#access-methods), allowing for deployment in WAN.

Managed mounts also allow for the use of a [pull priority function](https://pojntfx.github.io/networked-linux-memsync/main.html#background-pull-and-push); this allows an application to specify which chunks should be pulled first and in which order, which can be used to further increase throughput and decrease latency by having the most important chunks be available as quickly as possible. This is particularly useful if the resource has a known structure: For example, if the metadata is available at the end of a media file but needs to be available first to start playback, the pull priority function can [help optimize the pull process without having to change the format or re-encoding](https://pojntfx.github.io/networked-linux-memsync/main.html#universal-database-media-and-asset-streaming).

### Backends

Backends represent a way of accessing a resource (in the case of mounts) and locally storing a resource (in the case of migrations). They are defined as by a simple interface, as [introduced by go-nbd](https://github.com/pojntfx/go-nbd#1-define-a-backend):

```go
type Backend interface {
	ReadAt(p []byte, off int64) (n int, err error)
	WriteAt(p []byte, off int64) (n int, err error)
	Size() (int64, error)
	Sync() error
}
```

Since the interface is so simple, it is possible to represent almost any resource with it. There are also a few example backends available:

- [Memory](https://github.com/pojntfx/go-nbd/blob/main/pkg/backend/memory.go): Exposes a memory region as a resource
- [File](https://github.com/pojntfx/go-nbd/blob/main/pkg/backend/file.go): Exposes a file as a resource
- [Directory](./pkg/backend/directory.go): Exposes a directory of chunks as a resource
- [Redis](./pkg/backend/redis.go): Exposes a Redis database as a resource
- [Cassandra](./pkg/backend/cassandra.go): Exposes a Cassandra/ScyllaDB database as a resource
- [S3](./pkg/backend/s3.go): Exposes a S3 bucket as a resource
- [RPC](pkg/backend/rpc.go): Exposes any backend over an RPC framework of choice, such as gRPC

Different backends tend to have different characteristics, and behave [differently depending on network conditions and access patterns](https://pojntfx.github.io/networked-linux-memsync/main.html#backends). Depending on the backend used, it might also require a chunking system, which [can be implemented on both the client and server side](https://pojntfx.github.io/networked-linux-memsync/main.html#chunking-3); see the [mount benchmarks](./cmd/r3map-benchmark-managed-mount/main.go) for more information.

## Acknowledgements

- [pojntfx/go-bd](https://github.com/pojntfx/go-nbd) provides the Go NBD client and server.
- [pojntfx/panrpc](https://github.com/pojntfx/panrpc) provides one of the example RPC frameworks for mounts and migrations.
- [gRPC](https://grpc.io/) provides a reliable RPC framework for mounts and migrations.
- [fRPC](https://frpc.io/) provides a high-performance RPC framework for mounts and migrations.

## Contributing

To contribute, please use the [GitHub flow](https://guides.github.com/introduction/flow/) and follow our [Code of Conduct](./CODE_OF_CONDUCT.md).

To build and start a development version of one of the examples locally, run the following:

```shell
$ git clone https://github.com/pojntfx/r3map.git
$ cd go-nbd

# Load the NBD kernel module
$ sudo modprobe nbd

# Run unit tests
$ make test

# Build integration tests/benchmarks
$ make -j$(nproc)
# Run integration tests/benchmarks
$ sudo make integration

# Run the migration examples
$ go build -o /tmp/r3map-example-migration ./cmd/r3map-example-migration/ && sudo /tmp/r3map-example-migration --laddr localhost:1337 --invalidate 10 # Starts the first seeder
$ go build -o /tmp/r3map-example-migration ./cmd/r3map-example-migration/ && sudo /tmp/r3map-example-migration --raddr localhost:1337 --laddr localhost:1337 --invalidate 10 # First leeches the resource from a first seeder, then starts seeding
$ go build -o /tmp/r3map-example-migration ./cmd/r3map-example-migration/ && sudo /tmp/r3map-example-migration --raddr localhost:1337 # Leeches the resource from a seeder again, but doesn't start seeding afterwards

# Run the mount examples
$ go build -o /tmp/r3map-example-direct-mount-file ./cmd/r3map-example-direct-mount-file/ && sudo /tmp/r3map-example-direct-mount-file # Mounts a temporary file with a direct mount
$ go run ./cmd/r3map-example-mount-server/ # Starts the server exposing the resource
$ go build -o /tmp/r3map-example-direct-mount-client ./cmd/r3map-example-direct-mount-client/ && sudo /tmp/r3map-example-direct-mount-client # Mounts the resource with a direct mount
$ go build -o /tmp/r3map-example-managed-mount-client ./cmd/r3map-example-managed-mount-client/ && sudo /tmp/r3map-example-managed-mount-client # Mounts the resource with a managed mount
```

Have any questions or need help? Chat with us [on Matrix](https://matrix.to/#/#r3map:matrix.org?via=matrix.org)!

## License

r3map (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: Apache-2.0
